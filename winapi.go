//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	modKernel32              = syscall.NewLazyDLL("kernel32.dll")
	modUser32                = syscall.NewLazyDLL("user32.dll")
	modShell32               = syscall.NewLazyDLL("shell32.dll")
	modAdvapi32              = syscall.NewLazyDLL("advapi32.dll")
	procCreateMutexW         = modKernel32.NewProc("CreateMutexW")
	procGetLastError         = modKernel32.NewProc("GetLastError")
	procGetModuleHandleW     = modKernel32.NewProc("GetModuleHandleW")
	procRegisterClassExW     = modUser32.NewProc("RegisterClassExW")
	procCreateWindowExW      = modUser32.NewProc("CreateWindowExW")
	procDefWindowProcW       = modUser32.NewProc("DefWindowProcW")
	procDestroyWindow        = modUser32.NewProc("DestroyWindow")
	procGetComboBoxInfo      = modUser32.NewProc("GetComboBoxInfo")
	procGetMessageW          = modUser32.NewProc("GetMessageW")
	procTranslateMessage     = modUser32.NewProc("TranslateMessage")
	procDispatchMessageW     = modUser32.NewProc("DispatchMessageW")
	procPostQuitMessage      = modUser32.NewProc("PostQuitMessage")
	procCreateIconFromResourceEx = modUser32.NewProc("CreateIconFromResourceEx")
	procLoadIconW            = modUser32.NewProc("LoadIconW")
	procDestroyIcon          = modUser32.NewProc("DestroyIcon")
	procCreatePopupMenu      = modUser32.NewProc("CreatePopupMenu")
	procAppendMenuW          = modUser32.NewProc("AppendMenuW")
	procDestroyMenu          = modUser32.NewProc("DestroyMenu")
	procTrackPopupMenu       = modUser32.NewProc("TrackPopupMenu")
	procSetForegroundWindow  = modUser32.NewProc("SetForegroundWindow")
	procPostMessageW         = modUser32.NewProc("PostMessageW")
	procGetCursorPos         = modUser32.NewProc("GetCursorPos")
	procRegisterHotKey       = modUser32.NewProc("RegisterHotKey")
	procUnregisterHotKey     = modUser32.NewProc("UnregisterHotKey")
	procMessageBoxW          = modUser32.NewProc("MessageBoxW")
	procRedrawWindow         = modUser32.NewProc("RedrawWindow")
	procSendMessageW         = modUser32.NewProc("SendMessageW")
	procSetWindowTextW       = modUser32.NewProc("SetWindowTextW")
	procShowWindow           = modUser32.NewProc("ShowWindow")
	procUpdateWindow         = modUser32.NewProc("UpdateWindow")
	procSetFocus             = modUser32.NewProc("SetFocus")
	procGetKeyState          = modUser32.NewProc("GetKeyState")
	procShellNotifyIconW     = modShell32.NewProc("Shell_NotifyIconW")
	procRegOpenKeyExW        = modAdvapi32.NewProc("RegOpenKeyExW")
	procRegCreateKeyExW      = modAdvapi32.NewProc("RegCreateKeyExW")
	procRegQueryValueExW     = modAdvapi32.NewProc("RegQueryValueExW")
	procRegSetValueExW       = modAdvapi32.NewProc("RegSetValueExW")
	procRegDeleteValueW      = modAdvapi32.NewProc("RegDeleteValueW")
	procRegCloseKey          = modAdvapi32.NewProc("RegCloseKey")
	mainWndProcPtr           = syscall.NewCallback(mainWndProc)
	dialogWndProcPtr         = syscall.NewCallback(dialogWndProc)
)

func createStatic(parent uintptr, text string, x, y, width, height int32) uintptr {
	return createControl("STATIC", text, ssLeft|wsChild|wsVisible, 0, x, y, width, height, parent, 0)
}

func createButton(parent uintptr, text string, id uintptr, x, y, width, height int32) uintptr {
	return createControl("BUTTON", text, bsPushButton|wsChild|wsVisible, 0, x, y, width, height, parent, id)
}

func createCheckBox(parent uintptr, text string, id uintptr, x, y, width, height int32) uintptr {
	return createControl("BUTTON", text, bsAutoCheckBox|wsChild|wsVisible|wsTabStop, 0, x, y, width, height, parent, id)
}

func createListBox(parent, id uintptr, x, y, width, height int32) uintptr {
	return createControl("LISTBOX", "", lbsNotify|wsBorder|wsChild|wsVisible|wsTabStop|wsVScroll, 0, x, y, width, height, parent, id)
}

func createComboBox(parent, id uintptr, x, y, width, height int32) uintptr {
	return createControl("COMBOBOX", "", cbsDropdownList|wsChild|wsVisible|wsTabStop|wsVScroll, 0, x, y, width, height, parent, id)
}

func createControl(className, text string, style, exStyle uint32, x, y, width, height int32, parent, id uintptr) uintptr {
	hwnd, _, _ := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(toUTF16Ptr(className))),
		uintptr(unsafe.Pointer(toUTF16Ptr(text))),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		parent,
		id,
		currentApp.hInstance,
		0,
	)
	return hwnd
}

func appendMenu(menu uintptr, flags uint32, id uint32, text string) {
	var textPtr *uint16
	if text != "" {
		textPtr = toUTF16Ptr(text)
	}
	procAppendMenuW.Call(menu, uintptr(flags), uintptr(id), uintptr(unsafe.Pointer(textPtr)))
}

func defWindowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	r1, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return r1
}

func lowWord(value uint32) uint32 {
	return value & 0xFFFF
}

func highWord(value uint32) uint32 {
	return value >> 16
}

func sendMessage(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	r1, _, _ := procSendMessageW.Call(hwnd, uintptr(message), wParam, lParam)
	return r1
}

func redrawWindow(hwnd uintptr, flags uint32) {
	procRedrawWindow.Call(hwnd, 0, 0, uintptr(flags))
}

func getComboBoxInfo(hwnd uintptr) (comboBoxInfo, bool) {
	info := comboBoxInfo{CbSize: uint32(unsafe.Sizeof(comboBoxInfo{}))}
	r1, _, _ := procGetComboBoxInfo.Call(hwnd, uintptr(unsafe.Pointer(&info)))
	return info, r1 != 0
}

func toUTF16Ptr(value string) *uint16 {
	ptr, _ := syscall.UTF16PtrFromString(value)
	return ptr
}

func showMessageBox(hwnd uintptr, text, title string, flags uint32) {
	procMessageBoxW.Call(
		hwnd,
		uintptr(unsafe.Pointer(toUTF16Ptr(text))),
		uintptr(unsafe.Pointer(toUTF16Ptr(title))),
		uintptr(flags),
	)
}
