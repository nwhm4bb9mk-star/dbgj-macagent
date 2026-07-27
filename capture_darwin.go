//go:build darwin

package main

/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation -framework ApplicationServices
#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>

// macOS 15+: CGDisplayCreateImage unavailable; capture via screencapture

static void dbgj_mouse(int action, int button, int x, int y, int delta) {
    CGPoint pt = CGPointMake((CGFloat)x, (CGFloat)y);
    CGMouseButton mb = kCGMouseButtonLeft;
    if (button == 2) mb = kCGMouseButtonRight;
    else if (button == 3) mb = kCGMouseButtonCenter;

    if (action == 6) {
        CGEventRef w = CGEventCreateScrollWheelEvent(NULL, kCGScrollEventUnitLine, 1, delta);
        if (w) { CGEventPost(kCGHIDEventTap, w); CFRelease(w); }
        return;
    }
    if (action == 4) {
        CGEventType dn = (mb == kCGMouseButtonRight) ? kCGEventRightMouseDown : kCGEventLeftMouseDown;
        CGEventType up = (mb == kCGMouseButtonRight) ? kCGEventRightMouseUp : kCGEventLeftMouseUp;
        CGEventRef d = CGEventCreateMouseEvent(NULL, dn, pt, mb);
        CGEventRef u = CGEventCreateMouseEvent(NULL, up, pt, mb);
        if (d) { CGEventPost(kCGHIDEventTap, d); CFRelease(d); }
        if (u) { CGEventPost(kCGHIDEventTap, u); CFRelease(u); }
        return;
    }
    if (action == 5) {
        CGEventRef d1 = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, pt, kCGMouseButtonLeft);
        CGEventRef u1 = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp, pt, kCGMouseButtonLeft);
        CGEventRef d2 = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, pt, kCGMouseButtonLeft);
        CGEventRef u2 = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp, pt, kCGMouseButtonLeft);
        if (d1) { CGEventSetIntegerValueField(d1, kCGMouseEventClickState, 1); CGEventPost(kCGHIDEventTap, d1); CFRelease(d1); }
        if (u1) { CGEventSetIntegerValueField(u1, kCGMouseEventClickState, 1); CGEventPost(kCGHIDEventTap, u1); CFRelease(u1); }
        if (d2) { CGEventSetIntegerValueField(d2, kCGMouseEventClickState, 2); CGEventPost(kCGHIDEventTap, d2); CFRelease(d2); }
        if (u2) { CGEventSetIntegerValueField(u2, kCGMouseEventClickState, 2); CGEventPost(kCGHIDEventTap, u2); CFRelease(u2); }
        return;
    }
    CGEventType et = kCGEventMouseMoved;
    if (action == 2) {
        et = (mb == kCGMouseButtonRight) ? kCGEventRightMouseDown :
             (mb == kCGMouseButtonCenter) ? kCGEventOtherMouseDown : kCGEventLeftMouseDown;
    } else if (action == 3) {
        et = (mb == kCGMouseButtonRight) ? kCGEventRightMouseUp :
             (mb == kCGMouseButtonCenter) ? kCGEventOtherMouseUp : kCGEventLeftMouseUp;
    }
    CGEventRef e = CGEventCreateMouseEvent(NULL, et, pt, mb);
    if (e) { CGEventPost(kCGHIDEventTap, e); CFRelease(e); }
}

static void dbgj_key_vk(int down, int macCode) {
    CGEventRef e = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)macCode, down ? true : false);
    if (!e) return;
    CGEventPost(kCGHIDEventTap, e);
    CFRelease(e);
}

static void dbgj_unicode(const UniChar *chars, int n) {
    CGEventRef d = CGEventCreateKeyboardEvent(NULL, 0, true);
    CGEventRef u = CGEventCreateKeyboardEvent(NULL, 0, false);
    if (d) {
        CGEventKeyboardSetUnicodeString(d, n, chars);
        CGEventPost(kCGHIDEventTap, d);
        CFRelease(d);
    }
    if (u) {
        CGEventKeyboardSetUnicodeString(u, n, chars);
        CGEventPost(kCGHIDEventTap, u);
        CFRelease(u);
    }
}
*/
import "C"
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"unicode/utf16"
	"unsafe"
)

func init() {
	captureImpl = darwinCaptureJPEG
	inputImpl = darwinInput
}

func darwinCaptureJPEG(quality, scale int) ([]byte, int, int, error) {
	_ = quality
	_ = scale
	return captureViaScreencapture()
}

func captureViaScreencapture() ([]byte, int, int, error) {
	dir, err := os.MkdirTemp("", "dbgj-cap-*")
	if err != nil {
		return nil, 0, 0, err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "f.jpg")
	cmd := exec.Command("screencapture", "-x", "-t", "jpg", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, 0, 0, fmt.Errorf("采屏失败(请开屏幕录制权限): %v %s", err, string(out))
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, 0, err
	}
	w, h := jpegSize(b)
	return b, w, h, nil
}

func jpegSize(b []byte) (int, int) {
	for i := 0; i+8 < len(b); i++ {
		if b[i] == 0xFF && (b[i+1] == 0xC0 || b[i+1] == 0xC2) {
			h := int(b[i+5])<<8 | int(b[i+6])
			w := int(b[i+7])<<8 | int(b[i+8])
			return w, h
		}
	}
	return 1280, 720
}

func darwinInput(inputType, button, action string, x, y, delta, vk int, text string, sw, sh int) {
	if inputType == "keyboard" {
		if text != "" {
			u := utf16.Encode([]rune(text))
			if len(u) == 0 {
				return
			}
			C.dbgj_unicode((*C.UniChar)(unsafe.Pointer(&u[0])), C.int(len(u)))
			return
		}
		if vk > 0 {
			code := winVKToMac(vk)
			if action == "down" || action == "press" {
				C.dbgj_key_vk(1, C.int(code))
			}
			if action == "up" || action == "press" {
				C.dbgj_key_vk(0, C.int(code))
			}
		}
		return
	}
	btn := 1
	switch button {
	case "right":
		btn = 2
	case "middle":
		btn = 3
	}
	act := 1
	switch action {
	case "move":
		act = 1
	case "down":
		act = 2
	case "up":
		act = 3
	case "click":
		act = 4
	case "dblclick":
		act = 5
	case "wheel":
		act = 6
	}
	_ = sw
	_ = sh
	C.dbgj_mouse(C.int(act), C.int(btn), C.int(x), C.int(y), C.int(delta))
}

// winVKToMac：前端为 Windows VK → Mac keycode（常用键）
func winVKToMac(vk int) int {
	switch vk {
	case 0x08:
		return 51
	case 0x09:
		return 48
	case 0x0D:
		return 36
	case 0x1B:
		return 53
	case 0x20:
		return 49
	case 0x25:
		return 123
	case 0x26:
		return 126
	case 0x27:
		return 124
	case 0x28:
		return 125
	case 0x2E:
		return 117
	case 'A', 'a':
		return 0
	case 'S', 's':
		return 1
	case 'D', 'd':
		return 2
	case 'F', 'f':
		return 3
	case 'H', 'h':
		return 4
	case 'G', 'g':
		return 5
	case 'Z', 'z':
		return 6
	case 'X', 'x':
		return 7
	case 'C', 'c':
		return 8
	case 'V', 'v':
		return 9
	case 'B', 'b':
		return 11
	case 'Q', 'q':
		return 12
	case 'W', 'w':
		return 13
	case 'E', 'e':
		return 14
	case 'R', 'r':
		return 15
	case 'Y', 'y':
		return 16
	case 'T', 't':
		return 17
	case 'O', 'o':
		return 31
	case 'U', 'u':
		return 32
	case 'I', 'i':
		return 34
	case 'P', 'p':
		return 35
	case 'L', 'l':
		return 37
	case 'J', 'j':
		return 38
	case 'K', 'k':
		return 40
	case 'N', 'n':
		return 45
	case 'M', 'm':
		return 46
	case '1':
		return 18
	case '2':
		return 19
	case '3':
		return 20
	case '4':
		return 21
	case '5':
		return 23
	case '6':
		return 22
	case '7':
		return 26
	case '8':
		return 28
	case '9':
		return 25
	case '0':
		return 29
	}
	letters := map[int]int{
		0x41: 0, 0x42: 11, 0x43: 8, 0x44: 2, 0x45: 14, 0x46: 3, 0x47: 5, 0x48: 4,
		0x49: 34, 0x4A: 38, 0x4B: 40, 0x4C: 37, 0x4D: 46, 0x4E: 45, 0x4F: 31, 0x50: 35,
		0x51: 12, 0x52: 15, 0x53: 1, 0x54: 17, 0x55: 32, 0x56: 9, 0x57: 13, 0x58: 7,
		0x59: 16, 0x5A: 6,
	}
	if c, ok := letters[vk]; ok {
		return c
	}
	digits := map[int]int{0x30: 29, 0x31: 18, 0x32: 19, 0x33: 20, 0x34: 21, 0x35: 23, 0x36: 22, 0x37: 26, 0x38: 28, 0x39: 25}
	if c, ok := digits[vk]; ok {
		return c
	}
	return 0
}
