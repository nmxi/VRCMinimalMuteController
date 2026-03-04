//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

func main() {
	// Win32 メッセージループは同一スレッド固定で扱う。
	runtime.LockOSThread()

	appInstance, alreadyRunning, err := newApp()
	if err != nil {
		showMessageBox(0, err.Error(), "VRC Minimal Mute Controller", 0x10)
		return
	}
	if alreadyRunning {
		return
	}

	currentApp = appInstance
	defer appInstance.cleanup()

	if err := appInstance.run(); err != nil {
		showMessageBox(0, err.Error(), "VRC Minimal Mute Controller", 0x10)
	}
}

func newApp() (*app, bool, error) {
	mutexHandle, _, err := procCreateMutexW.Call(0, 1, uintptr(unsafe.Pointer(toUTF16Ptr(singleInstanceMutex))))
	if mutexHandle == 0 {
		return nil, false, fmt.Errorf("CreateMutexW failed: %v", err)
	}

	lastErr, _, _ := procGetLastError.Call()
	if lastErr == errorAlreadyExists {
		syscall.CloseHandle(syscall.Handle(mutexHandle))
		return nil, true, nil
	}

	hInstance, _, err := procGetModuleHandleW.Call(0)
	if hInstance == 0 {
		syscall.CloseHandle(syscall.Handle(mutexHandle))
		return nil, false, fmt.Errorf("GetModuleHandleW failed: %v", err)
	}

	return &app{
		mutexHandle: mutexHandle,
		hInstance:   hInstance,
	}, false, nil
}

func (a *app) run() error {
	if err := a.registerMainWindowClass(); err != nil {
		return err
	}
	if err := a.registerDialogWindowClass(); err != nil {
		return err
	}
	if err := a.createHiddenWindow(); err != nil {
		return err
	}

	a.hIcon = loadTrayIcon()
	if err := a.addTrayIcon(); err != nil {
		return err
	}

	a.loadConfiguredShortcut()
	return runMessageLoop()
}

func (a *app) cleanup() {
	if currentDialog != nil {
		procDestroyWindow.Call(currentDialog.hwnd)
		currentDialog = nil
	}

	if a.hotKeyRegistered {
		procUnregisterHotKey.Call(a.hwnd, hotKeyID)
		a.hotKeyRegistered = false
	}

	if a.hwnd != 0 {
		a.deleteTrayIcon()
		procDestroyWindow.Call(a.hwnd)
		a.hwnd = 0
	}

	if a.hIcon != 0 {
		procDestroyIcon.Call(a.hIcon)
		a.hIcon = 0
	}

	if a.mutexHandle != 0 {
		syscall.CloseHandle(syscall.Handle(a.mutexHandle))
		a.mutexHandle = 0
	}
}

func (a *app) registerMainWindowClass() error {
	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		LpfnWndProc:   mainWndProcPtr,
		HInstance:     a.hInstance,
		LpszClassName: toUTF16Ptr(windowClassName),
	}

	r1, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if r1 == 0 {
		return fmt.Errorf("RegisterClassExW failed: %v", err)
	}
	return nil
}

func (a *app) registerDialogWindowClass() error {
	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		LpfnWndProc:   dialogWndProcPtr,
		HInstance:     a.hInstance,
		LpszClassName: toUTF16Ptr(dialogClassName),
	}

	r1, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if r1 == 0 {
		return fmt.Errorf("RegisterClassExW dialog failed: %v", err)
	}
	return nil
}

func (a *app) createHiddenWindow() error {
	hwnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(toUTF16Ptr(windowClassName))),
		uintptr(unsafe.Pointer(toUTF16Ptr(appName))),
		wsOverlapped,
		0,
		0,
		0,
		0,
		0,
		0,
		a.hInstance,
		0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW failed: %v", err)
	}

	a.hwnd = hwnd
	return nil
}

func (a *app) addTrayIcon() error {
	data := notifyIconData{
		CbSize:            uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:              a.hwnd,
		UID:               trayUID,
		UFlags:            nifMessage | nifIcon | nifTip,
		UCallbackMessage:  wmTrayIcon,
		HIcon:             a.hIcon,
	}
	copy(data.SzTip[:], syscall.StringToUTF16("VRC Minimal Mute Controller"))

	r1, _, err := procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&data)))
	if r1 == 0 {
		return fmt.Errorf("Shell_NotifyIconW add failed: %v", err)
	}
	return nil
}

func (a *app) deleteTrayIcon() {
	data := notifyIconData{
		CbSize: uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:   a.hwnd,
		UID:    trayUID,
	}
	procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&data)))
}

func runMessageLoop() error {
	var message msg
	for {
		ret, _, err := procGetMessageW.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		switch int32(ret) {
		case -1:
			return fmt.Errorf("GetMessageW failed: %v", err)
		case 0:
			return nil
		default:
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&message)))
		}
	}
}

func mainWndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	if currentApp == nil {
		return defWindowProc(hwnd, message, wParam, lParam)
	}

	switch message {
	case wmCommand:
		switch lowWord(uint32(wParam)) {
		case menuShortcutSettingsID:
			currentApp.showShortcutDialog()
			return 0
		case menuStartupToggleID:
			currentApp.toggleStartup()
			return 0
		case menuExitID:
			currentApp.requestExit()
			return 0
		}
	case wmTrayIcon:
		switch uint32(lParam) {
		case wmLButtonDblClk:
			triggerOscSequence()
		case wmRButtonUp, wmContextMenu:
			currentApp.showContextMenu()
		}
		return 0
	case wmHotKey:
		if wParam == hotKeyID {
			triggerOscSequence()
			return 0
		}
	case wmDestroy:
		if currentApp != nil {
			currentApp.hwnd = 0
		}
		procPostQuitMessage.Call(0)
		return 0
	}

	return defWindowProc(hwnd, message, wParam, lParam)
}

func dialogWndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	if currentDialog == nil {
		return defWindowProc(hwnd, message, wParam, lParam)
	}

	switch message {
	case wmKeyDown, wmSysKeyDown:
		currentDialog.handleKey(uint32(wParam))
		return 0
	case wmCommand:
		switch lowWord(uint32(wParam)) {
		case dialogSaveButtonID:
			currentDialog.save()
			return 0
		case dialogDeleteButtonID:
			currentDialog.remove()
			return 0
		case dialogCancelButtonID:
			procDestroyWindow.Call(hwnd)
			return 0
		}
	case wmClose:
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		currentDialog = nil
		return 0
	}

	return defWindowProc(hwnd, message, wParam, lParam)
}

func (a *app) showContextMenu() {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	appendMenu(menu, mfString, menuShortcutSettingsID, "ショートカット設定")

	startupLabel := "スタートアップを有効化"
	if isStartupEnabled() {
		startupLabel = "スタートアップを無効化"
	}
	appendMenu(menu, mfString, menuStartupToggleID, startupLabel)
	appendMenu(menu, mfSeparator, 0, "")
	appendMenu(menu, mfString|mfGrayed, 0, "バージョン: "+appVersion)
	appendMenu(menu, mfSeparator, 0, "")
	appendMenu(menu, mfString, menuExitID, "Exit")

	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWindow.Call(a.hwnd)
	procTrackPopupMenu.Call(menu, tpmLeftAlign|tpmRightButton, uintptr(pt.X), uintptr(pt.Y), 0, a.hwnd, 0)
	procPostMessageW.Call(a.hwnd, wmNull, 0, 0)
}

func (a *app) toggleStartup() {
	if isStartupEnabled() {
		if err := disableStartup(); err != nil {
			showMessageBox(a.hwnd, "スタートアップ設定を無効化できませんでした。", "VRC Minimal Mute Controller", 0x10)
		}
		return
	}

	if err := enableStartup(); err != nil {
		showMessageBox(a.hwnd, "スタートアップ設定を有効化できませんでした。", "VRC Minimal Mute Controller", 0x10)
	}
}

func (a *app) requestExit() {
	if currentDialog != nil {
		procDestroyWindow.Call(currentDialog.hwnd)
		currentDialog = nil
	}

	if a.hotKeyRegistered {
		procUnregisterHotKey.Call(a.hwnd, hotKeyID)
		a.hotKeyRegistered = false
	}

	a.deleteTrayIcon()
	if a.hwnd != 0 {
		procDestroyWindow.Call(a.hwnd)
	}
}

func loadTrayIcon() uintptr {
	if iconHandle := createEmbeddedTrayIcon(embeddedTrayIcon); iconHandle != 0 {
		return iconHandle
	}

	r1, _, _ := procLoadIconW.Call(0, idiApplication)
	return r1
}

// .ico 内の最大サイズの画像を選んで HICON を生成する。
func createEmbeddedTrayIcon(iconFile []byte) uintptr {
	if len(iconFile) < 22 {
		return 0
	}

	iconCount := binary.LittleEndian.Uint16(iconFile[4:6])
	if iconCount == 0 {
		return 0
	}

	bestOffset := uint32(0)
	bestSize := uint32(0)
	bestArea := int32(-1)
	for i := uint16(0); i < iconCount; i++ {
		entryOffset := 6 + int(i)*16
		if entryOffset+16 > len(iconFile) {
			break
		}

		width := int32(iconFile[entryOffset])
		height := int32(iconFile[entryOffset+1])
		if width == 0 {
			width = 256
		}
		if height == 0 {
			height = 256
		}

		imageSize := binary.LittleEndian.Uint32(iconFile[entryOffset+8 : entryOffset+12])
		imageOffset := binary.LittleEndian.Uint32(iconFile[entryOffset+12 : entryOffset+16])
		if imageOffset == 0 || imageSize == 0 || int(imageOffset+imageSize) > len(iconFile) {
			continue
		}

		if area := width * height; area > bestArea {
			bestArea = area
			bestSize = imageSize
			bestOffset = imageOffset
		}
	}

	if bestOffset == 0 || bestSize == 0 {
		return 0
	}

	imageData := iconFile[bestOffset : bestOffset+bestSize]
	r1, _, _ := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&imageData[0])),
		uintptr(len(imageData)),
		1,
		0x00030000,
		0,
		0,
		0,
	)
	return r1
}
