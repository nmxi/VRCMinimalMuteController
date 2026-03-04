//go:build windows

package main

const (
	appName                = "ミュトコン"
	appVersion             = "1.0.0"
	windowClassName        = "VRCMinimalMuteController.HiddenWindow"
	dialogClassName        = "VRCMinimalMuteController.ShortcutDialog"
	singleInstanceMutex    = "VRCMinimalMuteController.SingleInstance"
	startupRegistryPath    = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	settingsRegistryPath   = `HKCU\Software\VRCMinimalMuteController`
	shortcutRegistryValue  = "ShortcutKey"
	startupRegistryValue   = "VRCMinimalMuteController"
	trayUID                = 1
	hotKeyID               = 1
	shortcutShiftMask      = 0x00010000
	shortcutControlMask    = 0x00020000
	shortcutAltMask        = 0x00040000
	modAlt                 = 0x0001
	modControl             = 0x0002
	modShift               = 0x0004
	wmApp                  = 0x8000
	wmTrayIcon             = wmApp + 1
	wmDestroy              = 0x0002
	wmCommand              = 0x0111
	wmClose                = 0x0010
	wmNull                 = 0x0000
	wmKeyDown              = 0x0100
	wmSysKeyDown           = 0x0104
	wmHotKey               = 0x0312
	wmLButtonDblClk        = 0x0203
	wmRButtonUp            = 0x0205
	wmContextMenu          = 0x007B
	wsOverlapped           = 0x00000000
	wsCaption              = 0x00C00000
	wsSysMenu              = 0x00080000
	wsVisible              = 0x10000000
	wsChild                = 0x40000000
	bsPushButton           = 0x00000000
	ssLeft                 = 0x00000000
	swShow                 = 5
	cwUseDefault           = 0x80000000
	nimAdd                 = 0x00000000
	nimDelete              = 0x00000002
	nifMessage             = 0x00000001
	nifIcon                = 0x00000002
	nifTip                 = 0x00000004
	tpmLeftAlign           = 0x0000
	tpmRightButton         = 0x0002
	mfString               = 0x00000000
	mfGrayed               = 0x00000001
	mfSeparator            = 0x00000800
	idiApplication         = 32512
	vkCancel               = 0x03
	vkShift                = 0x10
	vkControl              = 0x11
	vkMenu                 = 0x12
	vkPause                = 0x13
	errorAlreadyExists     = 183
	menuShortcutSettingsID = 1001
	menuStartupToggleID    = 1002
	menuExitID             = 1003
	dialogSaveButtonID     = 2001
	dialogDeleteButtonID   = 2002
	dialogCancelButtonID   = 2003
)

const (
	hKeyCurrentUser    = 0x80000001
	keyQueryValue      = 0x0001
	keySetValue        = 0x0002
	keyCreateSubKey    = 0x0004
	keyRead            = 0x20019
	keyWrite           = 0x20006
	regOptionNonVolatile = 0x00000000
	regSZ              = 1
	regDWORD           = 4
	errorSuccess       = 0
	errorFileNotFound  = 2
)

type point struct {
	X int32
	Y int32
}

type msg struct {
	HWnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       point
	LPrivate uint32
}

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type notifyIconData struct {
	CbSize            uint32
	HWnd              uintptr
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             uintptr
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          guid
	HBalloonIcon      uintptr
}

type app struct {
	mutexHandle        uintptr
	hInstance          uintptr
	hwnd               uintptr
	hIcon              uintptr
	configuredShortcut uint32
	hotKeyRegistered   bool
}

type shortcutDialog struct {
	hwnd             uintptr
	currentLabel     uintptr
	selectedLabel    uintptr
	selectedShortcut uint32
}

var (
	currentApp          *app
	currentDialog       *shortcutDialog
	oscSequenceRunning  int32
)
