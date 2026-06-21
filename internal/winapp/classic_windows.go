//go:build windows

package winapp

import (
	"time"

	"mofumouse/internal/core"
)

const (
	vkA     = 0x41
	vkC     = 0x43
	vkLeft  = 0x25
	vkRight = 0x27
	vkD     = 0x44
	vkF     = 0x46
	vkT     = 0x54
	vkLWin  = 0x5B
	vkMenu  = 0x12
	vkCtrl  = 0x11

	keyEventFKeyUp = 0x0002
)

func ShowDesktop() {
	pressChord(vkLWin, vkD)
}

func OpenStartMenu() {
	keyDown(vkLWin)
	time.Sleep(40 * time.Millisecond)
	keyUp(vkLWin)
}

func BrowserBack() {
	pressChord(vkMenu, vkLeft)
}

func BrowserForward() {
	pressChord(vkMenu, vkRight)
}

func RunQuickKey(key core.QuickKeyID) bool {
	chord, ok := quickKeyChord(key)
	if !ok {
		return false
	}
	pressChord(chord.modifier, chord.key)
	return true
}

type keyChord struct {
	modifier byte
	key      byte
}

func quickKeyChord(key core.QuickKeyID) (keyChord, bool) {
	switch key {
	case core.QuickKeyFind:
		return keyChord{modifier: vkCtrl, key: vkF}, true
	case core.QuickKeyCopy:
		return keyChord{modifier: vkCtrl, key: vkC}, true
	case core.QuickKeySelectAll:
		return keyChord{modifier: vkCtrl, key: vkA}, true
	case core.QuickKeyNewTab:
		return keyChord{modifier: vkCtrl, key: vkT}, true
	default:
		return keyChord{}, false
	}
}

func pressChord(modifier, key byte) {
	keyDown(modifier)
	time.Sleep(20 * time.Millisecond)
	keyDown(key)
	time.Sleep(30 * time.Millisecond)
	keyUp(key)
	time.Sleep(20 * time.Millisecond)
	keyUp(modifier)
}

func keyDown(vk byte) {
	procKeybdEvent.Call(uintptr(vk), 0, 0, 0)
}

func keyUp(vk byte) {
	procKeybdEvent.Call(uintptr(vk), 0, keyEventFKeyUp, 0)
}
