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
	wsTabStop              = 0x00010000
	wsVScroll              = 0x00200000
	wsVisible              = 0x10000000
	wsChild                = 0x40000000
	bsPushButton           = 0x00000000
	bsAutoCheckBox         = 0x00000003
	cbsDropdownList        = 0x0003
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
	cbGetCurSel            = 0x0147
	cbAddString            = 0x0143
	cbSetCurSel            = 0x014E
	bmGetCheck             = 0x00F0
	bmSetCheck             = 0x00F1
	bstUnchecked           = 0x0000
	bstChecked             = 0x0001
	cbnSelChange           = 1
	idiApplication         = 32512
	vkCancel               = 0x03
	vkBack                 = 0x08
	vkTab                  = 0x09
	vkReturn               = 0x0D
	vkShift                = 0x10
	vkControl              = 0x11
	vkMenu                 = 0x12
	vkPause                = 0x13
	vkCapital              = 0x14
	vkEscape               = 0x1B
	vkSpace                = 0x20
	vkPrior                = 0x21
	vkNext                 = 0x22
	vkEnd                  = 0x23
	vkHome                 = 0x24
	vkLeft                 = 0x25
	vkUp                   = 0x26
	vkRight                = 0x27
	vkDown                 = 0x28
	vkSnapshot             = 0x2C
	vkInsert               = 0x2D
	vkDelete               = 0x2E
	vkLWin                 = 0x5B
	vkRWin                 = 0x5C
	vkApps                 = 0x5D
	vkNumpad0              = 0x60
	vkNumpad1              = 0x61
	vkNumpad2              = 0x62
	vkNumpad3              = 0x63
	vkNumpad4              = 0x64
	vkNumpad5              = 0x65
	vkNumpad6              = 0x66
	vkNumpad7              = 0x67
	vkNumpad8              = 0x68
	vkNumpad9              = 0x69
	vkMultiply             = 0x6A
	vkAdd                  = 0x6B
	vkSeparator            = 0x6C
	vkSubtract             = 0x6D
	vkDecimal              = 0x6E
	vkDivide               = 0x6F
	vkNumLock              = 0x90
	vkScroll               = 0x91
	errorAlreadyExists     = 183
	menuShortcutSettingsID = 1001
	menuStartupToggleID    = 1002
	menuExitID             = 1003
	dialogSaveButtonID     = 2001
	dialogDeleteButtonID   = 2002
	dialogCancelButtonID   = 2003
	dialogKeyComboID       = 2004
	dialogCtrlCheckID      = 2005
	dialogShiftCheckID     = 2006
	dialogAltCheckID       = 2007
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
	ctrlCheck        uintptr
	shiftCheck       uintptr
	altCheck         uintptr
	keyCombo         uintptr
	selectedShortcut uint32
}

var (
	currentApp          *app
	currentDialog       *shortcutDialog
	oscSequenceRunning  int32
)
