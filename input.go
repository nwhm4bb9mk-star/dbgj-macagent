package main

import (
	"encoding/json"
	"fmt"
)

// HandleScreenInput 对齐 web ScreenFloatPanel：mouse / keyboard（兼 key）
func HandleScreenInput(payload []byte) {
	var p struct {
		InputType string `json:"inputType"`
		X         any    `json:"x"`
		Y         any    `json:"y"`
		Button    string `json:"button"`
		Action    string `json:"action"`
		Delta     any    `json:"delta"`
		VK        any    `json:"vk"`
		Text      string `json:"text"`
		SW        any    `json:"sw"`
		SH        any    `json:"sh"`
	}
	_ = json.Unmarshal(payload, &p)
	it := p.InputType
	if it == "key" {
		it = "keyboard"
	}
	if inputImpl != nil {
		inputImpl(it, p.Button, p.Action, toInt(p.X), toInt(p.Y), toInt(p.Delta), toInt(p.VK), p.Text, toInt(p.SW), toInt(p.SH))
	}
}

var inputImpl func(inputType, button, action string, x, y, delta, vk int, text string, sw, sh int)

func toInt(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case json.Number:
		n, _ := t.Int64()
		return int(n)
	case string:
		var n int
		fmt.Sscanf(t, "%d", &n)
		return n
	default:
		return 0
	}
}
