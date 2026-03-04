//go:build windows

package main

import (
	"sync/atomic"
	"syscall"
	"unsafe"
)

const (
	whKeyboardLL = 13

	// low-level hook の wParam に来る
	wmKeyDownLL    = 0x0100
	wmSysKeyDownLL = 0x0104
)

type kbdllhookstruct struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

var (
	keyboardHookHandle uintptr
	keyboardHookProc   = syscall.NewCallback(lowLevelKeyboardProc)

	// 長押し等の連発を軽く抑制
	triggerGuard int32
)

func installKeyboardHook() bool {
	if keyboardHookHandle != 0 {
		return true
	}

	h, _, _ := procSetWindowsHookExW.Call(
		uintptr(whKeyboardLL),
		keyboardHookProc,
		0, // WH_KEYBOARD_LL は 0 でOK
		0, // 全スレッド
	)
	if h == 0 {
		return false
	}
	keyboardHookHandle = h
	return true
}

func uninstallKeyboardHook() {
	if keyboardHookHandle != 0 {
		procUnhookWindowsHookEx.Call(keyboardHookHandle)
		keyboardHookHandle = 0
	}
}

func lowLevelKeyboardProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	next := func() uintptr {
		r, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return r
	}

	if nCode < 0 || currentApp == nil {
		return next()
	}
	if wParam != wmKeyDownLL && wParam != wmSysKeyDownLL {
		return next()
	}

	k := (*kbdllhookstruct)(unsafe.Pointer(lParam))

	shortcut := currentApp.configuredShortcut
	keyCode := shortcut & 0xFFFF
	if keyCode == 0 {
		return next()
	}

	if keyCode == vkPause {
		if k.VkCode != vkPause && k.VkCode != vkCancel {
			return next()
		}
		if !matchesConfiguredModifiers(shortcut) {
			return next()
		}

		if !atomic.CompareAndSwapInt32(&triggerGuard, 0, 1) {
			return next()
		}
		go func() {
			defer atomic.StoreInt32(&triggerGuard, 0)
			triggerOscSequence()
		}()
		return next()
	}

	// それ以外のキーは RegisterHotKey 側に任せる
	return next()
}

func matchesConfiguredModifiers(shortcut uint32) bool {
	needCtrl := (shortcut & shortcutControlMask) != 0
	needShift := (shortcut & shortcutShiftMask) != 0
	needAlt := (shortcut & shortcutAltMask) != 0

	ctrl := asyncKeyDown(vkControl)
	shift := asyncKeyDown(vkShift)
	alt := asyncKeyDown(vkMenu)

	return ctrl == needCtrl && shift == needShift && alt == needAlt
}

func asyncKeyDown(vk uint32) bool {
	state, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return (state & 0x8000) != 0
}