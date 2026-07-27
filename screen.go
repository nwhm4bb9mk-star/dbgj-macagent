package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"sync"
	"time"
)

// ScreenSession 桌面会话：采屏实现自研，本期先出可联调帧（占位图）；darwin 真采屏另接系统 API。
type ScreenSession struct {
	id       string
	fps      int
	quality  int
	scale    int
	onFrame  func(jpeg []byte, w, h int)
	onError  func(string)
	stopCh   chan struct{}
	stopped  sync.Once
}

func NewScreenSession(id string, fps, quality, scale int, onFrame func([]byte, int, int), onError func(string)) *ScreenSession {
	return &ScreenSession{
		id: id, fps: fps, quality: quality, scale: scale,
		onFrame: onFrame, onError: onError, stopCh: make(chan struct{}),
	}
}

func (s *ScreenSession) Start() {
	go s.loop()
}

func (s *ScreenSession) Stop() {
	s.stopped.Do(func() { close(s.stopCh) })
}

func (s *ScreenSession) loop() {
	iv := time.Second / time.Duration(s.fps)
	if iv < 20*time.Millisecond {
		iv = 20 * time.Millisecond
	}
	t := time.NewTicker(iv)
	defer t.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-t.C:
			b, w, h, err := CaptureFrameJPEG(s.quality, s.scale)
			if err != nil {
				continue
			}
			if s.onFrame != nil {
				s.onFrame(b, w, h)
			}
		}
	}
}

// CaptureFrameJPEG：darwin 由 captureImpl（CGDisplay / screencapture）；其它平台占位图仅供协议联调。
func CaptureFrameJPEG(quality, scale int) ([]byte, int, int, error) {
	if captureImpl != nil {
		return captureImpl(quality, scale)
	}
	w, h := 1280, 720
	if scale > 0 && scale < 100 {
		w = w * scale / 100
		h = h * scale / 100
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	c := color.RGBA{R: 30, G: 30, B: 40, A: 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	q := quality
	if q <= 0 || q > 100 {
		q = 70
	}
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
		return nil, 0, 0, err
	}
	return buf.Bytes(), w, h, nil
}

func screenMode() string {
	if captureImpl != nil {
		return "cg"
	}
	return "stub"
}

// captureImpl 由 darwin 注入
var captureImpl func(quality, scale int) ([]byte, int, int, error)
