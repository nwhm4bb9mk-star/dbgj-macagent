package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"
)

// AgentMessage 与 server_monitor/internal/ws.AgentMessage 同形
type AgentMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
	Ts      int64           `json:"ts,omitempty"`
	Sig     string          `json:"sig,omitempty"`
}

func signRaw(signKey []byte, typ, id string, payload []byte) (ts int64, sig string) {
	if len(signKey) == 0 {
		return 0, ""
	}
	ts = time.Now().Unix()
	raw := typ + "|" + id + "|" + strconv.FormatInt(ts, 10) + "|" + string(payload)
	mac := hmac.New(sha256.New, signKey)
	mac.Write([]byte(raw))
	return ts, hex.EncodeToString(mac.Sum(nil))
}

func verifyMsg(signKey []byte, m *AgentMessage) bool {
	if len(signKey) == 0 {
		return true
	}
	if m.Sig == "" || m.Ts == 0 {
		return false
	}
	if abs64(time.Now().Unix()-m.Ts) > 60 {
		return false
	}
	raw := m.Type + "|" + m.ID + "|" + strconv.FormatInt(m.Ts, 10) + "|" + string(m.Payload)
	mac := hmac.New(sha256.New, signKey)
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil)) == m.Sig
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

func marshalAgentMsg(signKey []byte, typ, id string, payload json.RawMessage) ([]byte, error) {
	m := AgentMessage{Type: typ, ID: id, Payload: payload}
	if m.Payload == nil {
		m.Payload = json.RawMessage("null")
	}
	ts, sig := signRaw(signKey, m.Type, m.ID, m.Payload)
	m.Ts, m.Sig = ts, sig
	return json.Marshal(m)
}
