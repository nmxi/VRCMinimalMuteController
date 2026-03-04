//go:build windows

package main

import (
	"fmt"
	"strings"
	"unsafe"
)

type shortcutKeyOption struct {
	vk    uint32
	label string
}

type shortcutValidationResult int

const (
	shortcutValid shortcutValidationResult = iota
	shortcutMissingKey
	shortcutNeedsModifier
	shortcutDisallowedKey
)

var shortcutKeyOptions = buildShortcutKeyOptions()

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
		292,
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
	d.currentLabel = createStatic(d.hwnd, "現在の設定値: "+formatShortcut(configured), 16, 16, 320, 20)
	createStatic(d.hwnd, "キーを入力、または選択してください。", 16, 64, 320, 20)
	createStatic(d.hwnd, "修飾キー:", 16, 94, 76, 20)
	d.ctrlCheck = createCheckBox(d.hwnd, "Ctrl", dialogCtrlCheckID, 96, 92, 120, 20)
	d.shiftCheck = createCheckBox(d.hwnd, "Shift", dialogShiftCheckID, 96, 116, 120, 20)
	d.altCheck = createCheckBox(d.hwnd, "Alt", dialogAltCheckID, 96, 140, 120, 20)
	createStatic(d.hwnd, "入力キー:", 16, 173, 76, 20)
	d.keyCombo = createComboBox(d.hwnd, dialogKeyComboID, 96, 170, 228, 320)
	for _, option := range shortcutKeyOptions {
		sendMessage(d.keyCombo, cbAddString, 0, uintptr(unsafe.Pointer(toUTF16Ptr(option.label))))
	}

	d.syncModifierChecks()
	d.syncKeyCombo()

	createButton(d.hwnd, "登録", dialogSaveButtonID, 16, 204, 90, 28)
	if configured != 0 {
		createButton(d.hwnd, "削除", dialogDeleteButtonID, 116, 204, 90, 28)
	}
	createButton(d.hwnd, "キャンセル", dialogCancelButtonID, 214, 204, 110, 28)
}

func (d *shortcutDialog) handleKey(vk uint32) {
	mods := currentModifierMask()
	switch vk {
	case vkControl, vkShift, vkMenu:
		d.selectedShortcut = mods | (d.selectedShortcut & 0xFFFF)
	default:
		d.selectedShortcut = mods | (vk & 0xFFFF)
	}
	d.refreshSelection()
}

func (d *shortcutDialog) handleComboSelection() {
	if d.keyCombo == 0 {
		return
	}

	index := int(sendMessage(d.keyCombo, cbGetCurSel, 0, 0))
	if index < 0 || index >= len(shortcutKeyOptions) {
		return
	}

	mods := d.selectedShortcut &^ 0xFFFF
	d.selectedShortcut = mods | shortcutKeyOptions[index].vk
	d.refreshSelection()
	procSetFocus.Call(d.hwnd)
}

func (d *shortcutDialog) handleModifierChange() {
	d.selectedShortcut = d.readModifierMask() | (d.selectedShortcut & 0xFFFF)
	d.refreshSelection()
}

func (d *shortcutDialog) refreshSelection() {
	d.syncModifierChecks()
	d.syncKeyCombo()
}

func (d *shortcutDialog) syncModifierChecks() {
	if d.ctrlCheck != 0 {
		setCheckBoxState(d.ctrlCheck, d.selectedShortcut&shortcutControlMask != 0)
	}
	if d.shiftCheck != 0 {
		setCheckBoxState(d.shiftCheck, d.selectedShortcut&shortcutShiftMask != 0)
	}
	if d.altCheck != 0 {
		setCheckBoxState(d.altCheck, d.selectedShortcut&shortcutAltMask != 0)
	}
}

func (d *shortcutDialog) syncKeyCombo() {
	if d.keyCombo == 0 {
		return
	}

	index := shortcutKeyIndex(d.selectedShortcut & 0xFFFF)
	if index >= 0 {
		sendMessage(d.keyCombo, cbSetCurSel, uintptr(index), 0)
		return
	}
	sendMessage(d.keyCombo, cbSetCurSel, ^uintptr(0), 0)
}

func (d *shortcutDialog) readModifierMask() uint32 {
	mask := uint32(0)
	if checkBoxChecked(d.ctrlCheck) {
		mask |= shortcutControlMask
	}
	if checkBoxChecked(d.shiftCheck) {
		mask |= shortcutShiftMask
	}
	if checkBoxChecked(d.altCheck) {
		mask |= shortcutAltMask
	}
	return mask
}

func (d *shortcutDialog) save() {
	switch validateShortcut(d.selectedShortcut) {
	case shortcutMissingKey:
		showMessageBox(d.hwnd, "修飾キーのみでは登録できません。", "ショートカット設定", 0x40)
		return
	case shortcutNeedsModifier:
		showMessageBox(d.hwnd, "選択されたキーは単独で登録することができません。修飾キーを含めて登録する必要があります。", "ショートカット設定", 0x40)
		return
	case shortcutDisallowedKey:
		showMessageBox(d.hwnd, "LWin と RWin は登録できません。", "ショートカット設定", 0x40)
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
	showMessageBox(d.hwnd, "ショートカットを "+formatShortcut(currentApp.configuredShortcut)+" に設定しました。", "ショートカット設定", 0x40)
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
		return fmt.Errorf("RegisterHotKey に失敗しました。他のアプリで使用中の可能性があります。")
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
	return validateShortcut(shortcut) == shortcutValid
}

func validateShortcut(shortcut uint32) shortcutValidationResult {
	keyCode := shortcut & 0xFFFF
	if keyCode == 0 {
		return shortcutMissingKey
	}

	if keyCode == vkLWin || keyCode == vkRWin {
		return shortcutDisallowedKey
	}

	if shortcut&(shortcutControlMask|shortcutShiftMask|shortcutAltMask) != 0 {
		return shortcutValid
	}

	if isSingleKeyAllowed(keyCode) {
		return shortcutValid
	}

	return shortcutNeedsModifier
}

func isSingleKeyAllowed(keyCode uint32) bool {
	switch keyCode {
	case vkCancel, vkPause, vkHome, vkEnd, vkPrior, vkNext, vkInsert, vkDelete:
		return true
	}

	return keyCode >= 0x70 && keyCode <= 0x87
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

func checkBoxChecked(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}
	return sendMessage(hwnd, bmGetCheck, 0, 0) == bstChecked
}

func setCheckBoxState(hwnd uintptr, checked bool) {
	if hwnd == 0 {
		return
	}
	state := uintptr(bstUnchecked)
	if checked {
		state = bstChecked
	}
	sendMessage(hwnd, bmSetCheck, state, 0)
}

func buildShortcutKeyOptions() []shortcutKeyOption {
	options := make([]shortcutKeyOption, 0, 64)

	for vk := uint32('A'); vk <= 'Z'; vk++ {
		options = append(options, shortcutKeyOption{vk: vk, label: string(rune(vk))})
	}
	for vk := uint32('0'); vk <= '9'; vk++ {
		options = append(options, shortcutKeyOption{vk: vk, label: string(rune(vk))})
	}
	for vk := uint32(0x70); vk <= 0x87; vk++ {
		options = append(options, shortcutKeyOption{vk: vk, label: fmt.Sprintf("F%d", vk-0x6F)})
	}

	extraKeys := []uint32{
		vkCancel,
		vkPause,
		vkSnapshot,
		vkLeft,
		vkUp,
		vkRight,
		vkDown,
		vkInsert,
		vkDelete,
		vkHome,
		vkEnd,
		vkPrior,
		vkNext,
		vkReturn,
		vkEscape,
		vkTab,
		vkBack,
		vkSpace,
		vkCapital,
		vkApps,
		vkNumpad0,
		vkNumpad1,
		vkNumpad2,
		vkNumpad3,
		vkNumpad4,
		vkNumpad5,
		vkNumpad6,
		vkNumpad7,
		vkNumpad8,
		vkNumpad9,
		vkMultiply,
		vkAdd,
		vkSeparator,
		vkSubtract,
		vkDecimal,
		vkDivide,
		vkNumLock,
		vkScroll,
	}
	for _, vk := range extraKeys {
		options = append(options, shortcutKeyOption{vk: vk, label: keyName(vk)})
	}

	return options
}

func shortcutKeyIndex(vk uint32) int {
	for i, option := range shortcutKeyOptions {
		if option.vk == vk {
			return i
		}
	}
	return -1
}

func formatShortcut(shortcut uint32) string {
	if shortcut == 0 {
		return "未設定"
	}

	parts := formatModifierParts(shortcut)

	keyCode := shortcut & 0xFFFF
	if keyCode != 0 {
		parts = append(parts, keyName(keyCode))
	}
	if len(parts) == 0 {
		return "未設定"
	}

	return strings.Join(parts, "+")
}

func formatModifierParts(shortcut uint32) []string {
	parts := make([]string, 0, 3)
	if shortcut&shortcutControlMask != 0 {
		parts = append(parts, "Ctrl")
	}
	if shortcut&shortcutShiftMask != 0 {
		parts = append(parts, "Shift")
	}
	if shortcut&shortcutAltMask != 0 {
		parts = append(parts, "Alt")
	}
	return parts
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
	case vkCancel:
		return "Break"
	case vkBack:
		return "Backspace"
	case vkPause:
		return "Pause"
	case vkCapital:
		return "CapsLock"
	case vkSpace:
		return "Space"
	case vkSnapshot:
		return "PrintScreen"
	case vkLeft:
		return "Left"
	case vkUp:
		return "Up"
	case vkRight:
		return "Right"
	case vkDown:
		return "Down"
	case vkInsert:
		return "Insert"
	case vkDelete:
		return "Delete"
	case vkLWin:
		return "LWin"
	case vkRWin:
		return "RWin"
	case vkApps:
		return "Apps"
	case vkNumpad0:
		return "NumPad0"
	case vkNumpad1:
		return "NumPad1"
	case vkNumpad2:
		return "NumPad2"
	case vkNumpad3:
		return "NumPad3"
	case vkNumpad4:
		return "NumPad4"
	case vkNumpad5:
		return "NumPad5"
	case vkNumpad6:
		return "NumPad6"
	case vkNumpad7:
		return "NumPad7"
	case vkNumpad8:
		return "NumPad8"
	case vkNumpad9:
		return "NumPad9"
	case vkMultiply:
		return "NumPad*"
	case vkAdd:
		return "NumPad+"
	case vkSeparator:
		return "Separator"
	case vkSubtract:
		return "NumPad-"
	case vkDecimal:
		return "NumPad."
	case vkDivide:
		return "NumPad/"
	case vkNumLock:
		return "NumLock"
	case vkScroll:
		return "ScrollLock"
	case vkHome:
		return "Home"
	case vkEnd:
		return "End"
	case vkPrior:
		return "PageUp"
	case vkNext:
		return "PageDown"
	case vkReturn:
		return "Enter"
	case vkEscape:
		return "Esc"
	case vkTab:
		return "Tab"
	}

	return fmt.Sprintf("VK_%02X", vk)
}
