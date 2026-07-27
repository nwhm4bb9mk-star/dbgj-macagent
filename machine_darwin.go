//go:build darwin

package main

import (
	"os/exec"
	"strings"
)

func readMacHardwareUUID() string {
	out, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, "IOPlatformUUID") {
			continue
		}
		parts := strings.Split(line, "\"")
		if len(parts) >= 4 {
			return parts[len(parts)-2]
		}
	}
	return ""
}

// 真采屏：后续在此接 ScreenCaptureKit / CGDisplay（自研，不链第三方远控 SDK）
func init() {
	// captureImpl 暂不覆盖 → 用 screen.go 占位 JPEG，真机联调协议；采屏实现下一迭代
}
