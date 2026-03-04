//go:build windows

package main

import (
	"fmt"
	"strings"
	"unsafe"
)

func (a *app) loadConfiguredShortcut() {
	shortcut, err := readShortcutSetting()
	if err != nil {
		return
	}

	if !isValidShortcut(shortcut) {
		_ = deleteShortcutSetting()
		return
	}

	a.configuredShortcut = shortcut
	a.hotKeyRegistered = a.tryRegisterShortcut(shortcut)
}

func (a *app) showShortcutDialog() {
	if currentDialog != nil {
		procShowWindow.Call(currentDialog.hwnd, swShow)
		procSetForegroundWindow.Call(currentDialog.hwnd)
		procSetFocus.Call(currentDialog.hwnd)
		return
	}

	hwnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(toUTF16Ptr(dialogClassName))),
		uintptr(unsafe.Pointer(toUTF16Ptr("ショートカット設定"))),
		wsOverlapped|wsCaption|wsSysMenu|wsVisible,
		cwUseDefault,
		cwUseDefault,
		360,
		200,
		0,
		0,
		a.hInstance,
		0,
	)
	if hwnd == 0 {
		showMessageBox(a.hwnd, fmt.Sprintf("ダイアログを開けませんでした: %v", err), "ショートカット設定", 0x10)
		return
	}

	currentDialog = &shortcutDialog{
		hwnd:             hwnd,
		selectedShortcut: a.configuredShortcut,
	}
	currentDialog.buildControls(a.configuredShortcut)
	procShowWindow.Call(hwnd, swShow)
	procUpdateWindow.Call(hwnd)
	procSetForegroundWindow.Call(hwnd)
	procSetFocus.Call(hwnd)
}

func (d *shortcutDialog) buildControls(configured uint32) {
	createStatic(d.hwnd, "キーを入力してください。", 16, 16, 300, 20)
	d.currentLabel = createStatic(d.hwnd, "現在の設定値: "+formatShortcut(configured), 16, 44, 320, 20)
	d.selectedLabel = createStatic(d.hwnd, "入力値: "+formatShortcut(d.selectedShortcut), 16, 74, 320, 24)

	createButton(d.hwnd, "登録", dialogSaveButtonID, 16, 118, 90, 28)
	if configured != 0 {
		createButton(d.hwnd, "削除", dialogDeleteButtonID, 116, 118, 90, 28)
	}
	createButton(d.hwnd, "キャンセル", dialogCancelButtonID, 214, 118, 110, 28)
}

func (d *shortcutDialog) handleKey(vk uint32) {
	mods := currentModifierMask()
	switch vk {
	case vkControl, vkShift, vkMenu:
		d.selectedShortcut = mods
	default:
		d.selectedShortcut = mods | (vk & 0xFFFF)
	}
	d.refreshSelection()
}

func (d *shortcutDialog) refreshSelection() {
	if d.selectedLabel == 0 {
		return
	}

	procSetWindowTextW.Call(
		d.selectedLabel,
		uintptr(unsafe.Pointer(toUTF16Ptr("入力値: "+formatShortcut(d.selectedShortcut)))),
	)
}

func (d *shortcutDialog) save() {
	if !isValidShortcut(d.selectedShortcut) {
		showMessageBox(d.hwnd, "登録するキーを入力してください。修飾キーのみでは登録できません。", "ショートカット設定", 0x40)
		return
	}
	if currentApp == nil {
		return
	}

	if err := currentApp.saveShortcut(d.selectedShortcut); err != nil {
		showMessageBox(d.hwnd, err.Error(), "ショートカット設定", 0x30)
		return
	}

	currentText := "現在の設定値: " + formatShortcut(currentApp.configuredShortcut)
	procSetWindowTextW.Call(d.currentLabel, uintptr(unsafe.Pointer(toUTF16Ptr(currentText))))
	procDestroyWindow.Call(d.hwnd)
}

func (d *shortcutDialog) remove() {
	if currentApp == nil {
		return
	}

	if err := currentApp.clearShortcut(); err != nil {
		showMessageBox(d.hwnd, "ショートカット設定を削除できませんでした。", "ショートカット設定", 0x10)
		return
	}
	procDestroyWindow.Call(d.hwnd)
}

func (a *app) saveShortcut(shortcut uint32) error {
	previous := a.configuredShortcut
	previousRegistered := a.hotKeyRegistered

	if a.hotKeyRegistered {
		procUnregisterHotKey.Call(a.hwnd, hotKeyID)
		a.hotKeyRegistered = false
	}

	if !a.tryRegisterShortcut(shortcut) {
		a.configuredShortcut = previous
		if previousRegistered && previous != 0 {
			a.hotKeyRegistered = a.tryRegisterShortcut(previous)
		}
		return fmt.Errorf("ショートカットを登録できませんでした。他のアプリで使用中の可能性があります。")
	}

	a.configuredShortcut = shortcut
	if err := writeShortcutSetting(shortcut); err != nil {
		procUnregisterHotKey.Call(a.hwnd, hotKeyID)
		a.hotKeyRegistered = false
		a.configuredShortcut = previous
		if previousRegistered && previous != 0 {
			a.hotKeyRegistered = a.tryRegisterShortcut(previous)
		}
		return fmt.Errorf("ショートカット設定を保存できませんでした。")
	}

	return nil
}

func (a *app) clearShortcut() error {
	previous := a.configuredShortcut
	previousRegistered := a.hotKeyRegistered

	if a.hotKeyRegistered {
		procUnregisterHotKey.Call(a.hwnd, hotKeyID)
		a.hotKeyRegistered = false
	}

	a.configuredShortcut = 0
	if err := deleteShortcutSetting(); err != nil {
		a.configuredShortcut = previous
		if previousRegistered && previous != 0 {
			a.hotKeyRegistered = a.tryRegisterShortcut(previous)
		}
		return err
	}

	return nil
}

func (a *app) tryRegisterShortcut(shortcut uint32) bool {
	keyCode := shortcut & 0xFFFF
	if keyCode == 0 {
		return false
	}

	modifiers := uint32(0)
	if shortcut&shortcutControlMask != 0 {
		modifiers |= modControl
	}
	if shortcut&shortcutShiftMask != 0 {
		modifiers |= modShift
	}
	if shortcut&shortcutAltMask != 0 {
		modifiers |= modAlt
	}

	r1, _, _ := procRegisterHotKey.Call(a.hwnd, hotKeyID, uintptr(modifiers), uintptr(keyCode))
	a.hotKeyRegistered = r1 != 0
	return a.hotKeyRegistered
}

func isValidShortcut(shortcut uint32) bool {
	return shortcut&0xFFFF != 0
}

func currentModifierMask() uint32 {
	mask := uint32(0)
	if keyDown(vkControl) {
		mask |= shortcutControlMask
	}
	if keyDown(vkShift) {
		mask |= shortcutShiftMask
	}
	if keyDown(vkMenu) {
		mask |= shortcutAltMask
	}
	return mask
}

func keyDown(vk uint32) bool {
	state, _, _ := procGetKeyState.Call(uintptr(vk))
	return uint16(state)&0x8000 != 0
}

func formatShortcut(shortcut uint32) string {
	if shortcut == 0 {
		return "未設定"
	}

	parts := make([]string, 0, 4)
	if shortcut&shortcutControlMask != 0 {
		parts = append(parts, "Ctrl")
	}
	if shortcut&shortcutShiftMask != 0 {
		parts = append(parts, "Shift")
	}
	if shortcut&shortcutAltMask != 0 {
		parts = append(parts, "Alt")
	}

	keyCode := shortcut & 0xFFFF
	if keyCode != 0 {
		parts = append(parts, keyName(keyCode))
	}
	if len(parts) == 0 {
		return "未設定"
	}

	return strings.Join(parts, "+")
}

func keyName(vk uint32) string {
	switch {
	case vk >= '0' && vk <= '9':
		return string(rune(vk))
	case vk >= 'A' && vk <= 'Z':
		return string(rune(vk))
	case vk >= 0x70 && vk <= 0x87:
		return fmt.Sprintf("F%d", vk-0x6F)
	}

	switch vk {
	case 0x08:
		return "Backspace"
	case 0x13:
		return "Pause"
	case 0x14:
		return "CapsLock"
	case 0x20:
		return "Space"
	case 0x2C:
		return "PrintScreen"
	case 0x25:
		return "Left"
	case 0x26:
		return "Up"
	case 0x27:
		return "Right"
	case 0x28:
		return "Down"
	case 0x2D:
		return "Insert"
	case 0x2E:
		return "Delete"
	case 0x5B:
		return "LWin"
	case 0x5C:
		return "RWin"
	case 0x5D:
		return "Apps"
	case 0x60:
		return "NumPad0"
	case 0x61:
		return "NumPad1"
	case 0x62:
		return "NumPad2"
	case 0x63:
		return "NumPad3"
	case 0x64:
		return "NumPad4"
	case 0x65:
		return "NumPad5"
	case 0x66:
		return "NumPad6"
	case 0x67:
		return "NumPad7"
	case 0x68:
		return "NumPad8"
	case 0x69:
		return "NumPad9"
	case 0x6A:
		return "NumPad*"
	case 0x6B:
		return "NumPad+"
	case 0x6C:
		return "Separator"
	case 0x6D:
		return "NumPad-"
	case 0x6E:
		return "NumPad."
	case 0x6F:
		return "NumPad/"
	case 0x90:
		return "NumLock"
	case 0x91:
		return "ScrollLock"
	case 0x24:
		return "Home"
	case 0x23:
		return "End"
	case 0x21:
		return "PageUp"
	case 0x22:
		return "PageDown"
	case 0x0D:
		return "Enter"
	case 0x1B:
		return "Esc"
	case 0x09:
		return "Tab"
	}

	return fmt.Sprintf("VK_%02X", vk)
}
