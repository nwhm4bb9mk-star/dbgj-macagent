package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const agentVersion = "0.2.0-mac-desktop"

type Agent struct {
	serverURL string
	token     string
	signKey   []byte

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	wsMu sync.Mutex
	ws   *websocket.Conn

	screen *ScreenSession
}

func NewAgent(serverURL, token, signKey string) *Agent {
	ctx, cancel := context.WithCancel(context.Background())
	return &Agent{
		serverURL: strings.TrimRight(serverURL, "/"),
		token:     token,
		signKey:   []byte(signKey),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (a *Agent) Start() error {
	if a.token == "" {
		tok, sk, err := a.register()
		if err != nil {
			return err
		}
		a.token = tok
		if sk != "" {
			a.signKey = []byte(sk)
		}
		log.Printf("registered token=%s…", a.token[:min(8, len(a.token))])
	}
	a.wg.Add(2)
	go func() { defer a.wg.Done(); a.reportLoop() }()
	go func() { defer a.wg.Done(); a.wsLoop() }()
	return nil
}

func (a *Agent) Stop() {
	a.cancel()
	a.stopScreen()
	a.wsMu.Lock()
	if a.ws != nil {
		_ = a.ws.Close()
		a.ws = nil
	}
	a.wsMu.Unlock()
	a.wg.Wait()
}

func (a *Agent) register() (token, signKey string, err error) {
	body, _ := json.Marshal(map[string]string{
		"hostname":  hostname(),
		"os":        "darwin",
		"ip":        "",
		"machineId": machineID(),
	})
	resp, err := httpPost(a.serverURL+"/api/agent/register", body)
	if err != nil {
		return "", "", err
	}
	var out struct {
		Success bool `json:"success"`
		Data    struct {
			Token   string `json:"token"`
			SignKey string `json:"signKey"`
		} `json:"data"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(resp, &out); err != nil {
		return "", "", err
	}
	if !out.Success || out.Data.Token == "" {
		return "", "", fmt.Errorf("register: %s", out.Error)
	}
	return out.Data.Token, out.Data.SignKey, nil
}

func (a *Agent) reportLoop() {
	iv := 10 * time.Second
	for {
		select {
		case <-a.ctx.Done():
			return
		default:
		}
		payload, _ := json.Marshal(map[string]any{
			"token":   a.token,
			"version": agentVersion,
			"uptime":  "",
		})
		resp, err := httpPost(a.serverURL+"/api/agent/report", payload)
		if err != nil {
			iv = minDuration(iv*2, 60*time.Second)
		} else {
			iv = 10 * time.Second
			var out struct {
				Data struct {
					ServerURL string `json:"serverUrl"`
				} `json:"data"`
			}
			_ = json.Unmarshal(resp, &out)
			if u := strings.TrimRight(out.Data.ServerURL, "/"); u != "" && !strings.EqualFold(u, a.serverURL) {
				a.serverURL = u
			}
		}
		t := time.NewTimer(iv)
		select {
		case <-a.ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
	}
}

func (a *Agent) wsLoop() {
	backoff := time.Second
	for {
		select {
		case <-a.ctx.Done():
			return
		default:
		}
		err := a.wsSession()
		if a.ctx.Err() != nil {
			return
		}
		log.Printf("ws disconnected: %v; retry %s", err, backoff)
		t := time.NewTimer(backoff)
		select {
		case <-a.ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (a *Agent) wsSession() error {
	u, err := url.Parse(a.serverURL)
	if err != nil {
		return err
	}
	scheme := "wss"
	if u.Scheme == "http" {
		scheme = "ws"
	}
	wsURL := fmt.Sprintf("%s://%s/ws/agent?token=%s", scheme, u.Host, url.QueryEscape(a.token))
	dialer := websocket.Dialer{HandshakeTimeout: 30 * time.Second}
	conn, _, err := dialer.DialContext(a.ctx, wsURL, http.Header{
		"User-Agent": []string{"DBGJ-MacAgent/" + agentVersion},
	})
	if err != nil {
		return err
	}
	a.wsMu.Lock()
	a.ws = conn
	a.wsMu.Unlock()
	defer func() {
		a.stopScreen()
		a.wsMu.Lock()
		_ = conn.Close()
		if a.ws == conn {
			a.ws = nil
		}
		a.wsMu.Unlock()
	}()

	log.Printf("ws connected %s", u.Host)
	for {
		select {
		case <-a.ctx.Done():
			return a.ctx.Err()
		default:
		}
		_ = conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		a.handleWS(data)
	}
}

func (a *Agent) handleWS(data []byte) {
	var m AgentMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return
	}
	if m.Type == "pong" {
		return
	}
	if m.Type == "ping" {
		_ = a.sendMsg("pong", m.ID, json.RawMessage("null"))
		return
	}
	if !verifyMsg(a.signKey, &m) {
		log.Printf("ws bad sig type=%s", m.Type)
		return
	}
	switch m.Type {
	case "screen_start":
		a.handleScreenStart(m.ID, m.Payload)
	case "screen_stop":
		a.stopScreen()
	case "screen_input":
		HandleScreenInput(m.Payload)
	default:
		// MVP 只做桌面；其余类型忽略
	}
}

func (a *Agent) sendMsg(typ, id string, payload json.RawMessage) error {
	// Agent→服务端与 Windows MiniAgent 一致：出站不签 HMAC（入站仍验签）
	if payload == nil {
		payload = json.RawMessage("null")
	}
	b, err := json.Marshal(AgentMessage{Type: typ, ID: id, Payload: payload})
	if err != nil {
		return err
	}
	a.wsMu.Lock()
	defer a.wsMu.Unlock()
	if a.ws == nil {
		return fmt.Errorf("ws nil")
	}
	_ = a.ws.SetWriteDeadline(time.Now().Add(30 * time.Second))
	return a.ws.WriteMessage(websocket.TextMessage, b)
}

func (a *Agent) sendTextThenBinary(text []byte, bin []byte) error {
	a.wsMu.Lock()
	defer a.wsMu.Unlock()
	if a.ws == nil {
		return fmt.Errorf("ws nil")
	}
	_ = a.ws.SetWriteDeadline(time.Now().Add(30 * time.Second))
	if err := a.ws.WriteMessage(websocket.TextMessage, text); err != nil {
		return err
	}
	_ = a.ws.SetWriteDeadline(time.Now().Add(30 * time.Second))
	return a.ws.WriteMessage(websocket.BinaryMessage, bin)
}

func (a *Agent) handleScreenStart(id string, payload json.RawMessage) {
	a.stopScreen()
	fps, quality, scale := 10, 70, 100
	var p struct {
		FPS     int `json:"fps"`
		Quality int `json:"quality"`
		Scale   int `json:"scale"`
	}
	_ = json.Unmarshal(payload, &p)
	if p.FPS > 0 && p.FPS <= 60 {
		fps = p.FPS
	}
	if p.Quality > 0 && p.Quality <= 100 {
		quality = p.Quality
	}
	if p.Scale > 0 && p.Scale <= 100 {
		scale = p.Scale
	}
	sess := NewScreenSession(id, fps, quality, scale, func(jpeg []byte, w, h int) {
		header, _ := json.Marshal(AgentMessage{
			Type: "screen_frame",
			ID:   id,
			Payload: mustJSON(map[string]any{
				"width": w, "height": h, "size": len(jpeg),
				"x": 0, "y": 0, "cw": w, "ch": h, "full": true,
			}),
		})
		_ = a.sendTextThenBinary(header, jpeg)
	}, func(errMsg string) {
		b, _ := json.Marshal(errMsg)
		_ = a.sendMsg("screen_error", id, b)
	})
	a.screen = sess
	sess.Start()
	_ = a.sendMsg("screen_started", id, mustJSON(map[string]string{"mode": screenMode()}))
}

func (a *Agent) stopScreen() {
	if a.screen != nil {
		a.screen.Stop()
		a.screen = nil
	}
}

func httpPost(urlStr string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, urlStr, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	cli := &http.Client{Timeout: 30 * time.Second}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "mac-unknown"
	}
	return h
}

func machineID() string {
	if id := readMacHardwareUUID(); id != "" {
		return id
	}
	return hostname() + "-" + os.Getenv("USER")
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
