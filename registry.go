//go:build windows

package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

func isStartupEnabled() bool {
	key, err := openRegistryKey(settingsHivePath(startupRegistryPath), registrySubKey(startupRegistryPath), keyQueryValue)
	if err != nil {
		return false
	}
	defer closeRegistryKey(key)

	return hasRegistryValue(key, startupRegistryValue)
}

func enableStartup() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	key, err := createRegistryKey(settingsHivePath(startupRegistryPath), registrySubKey(startupRegistryPath))
	if err != nil {
		return err
	}
	defer closeRegistryKey(key)

	return setRegistryString(key, startupRegistryValue, `"`+exePath+`"`)
}

func disableStartup() error {
	key, err := openRegistryKey(settingsHivePath(startupRegistryPath), registrySubKey(startupRegistryPath), keySetValue)
	if err != nil {
		if isRegistryNotFound(err) {
			return nil
		}
		return err
	}
	defer closeRegistryKey(key)

	err = deleteRegistryValue(key, startupRegistryValue)
	if isRegistryNotFound(err) {
		return nil
	}
	return err
}

func readShortcutSetting() (uint32, error) {
	key, err := openRegistryKey(settingsHivePath(settingsRegistryPath), registrySubKey(settingsRegistryPath), keyQueryValue)
	if err != nil {
		return 0, err
	}
	defer closeRegistryKey(key)

	value, valueType, err := queryRegistryDWORD(key, shortcutRegistryValue)
	if err != nil {
		return 0, err
	}
	if valueType != regDWORD {
		return 0, fmt.Errorf("shortcut value has unexpected type: %d", valueType)
	}

	return value, nil
}

func writeShortcutSetting(shortcut uint32) error {
	key, err := createRegistryKey(settingsHivePath(settingsRegistryPath), registrySubKey(settingsRegistryPath))
	if err != nil {
		return err
	}
	defer closeRegistryKey(key)

	return setRegistryDWORD(key, shortcutRegistryValue, shortcut)
}

func deleteShortcutSetting() error {
	key, err := openRegistryKey(settingsHivePath(settingsRegistryPath), registrySubKey(settingsRegistryPath), keySetValue)
	if err != nil {
		if isRegistryNotFound(err) {
			return nil
		}
		return err
	}
	defer closeRegistryKey(key)

	err = deleteRegistryValue(key, shortcutRegistryValue)
	if isRegistryNotFound(err) {
		return nil
	}
	return err
}

func settingsHivePath(path string) uintptr {
	switch {
	case strings.HasPrefix(path, `HKCU\`):
		return hKeyCurrentUser
	default:
		return 0
	}
}

func registrySubKey(path string) string {
	switch {
	case strings.HasPrefix(path, `HKCU\`):
		return strings.TrimPrefix(path, `HKCU\`)
	default:
		return path
	}
}

func openRegistryKey(root uintptr, subKey string, access uint32) (uintptr, error) {
	if root == 0 {
		return 0, fmt.Errorf("unsupported registry root")
	}

	var key uintptr
	ret, _, _ := procRegOpenKeyExW.Call(
		root,
		uintptr(unsafe.Pointer(toUTF16Ptr(subKey))),
		0,
		uintptr(access),
		uintptr(unsafe.Pointer(&key)),
	)
	if ret != errorSuccess {
		return 0, syscall.Errno(ret)
	}

	return key, nil
}

func createRegistryKey(root uintptr, subKey string) (uintptr, error) {
	if root == 0 {
		return 0, fmt.Errorf("unsupported registry root")
	}

	var key uintptr
	ret, _, _ := procRegCreateKeyExW.Call(
		root,
		uintptr(unsafe.Pointer(toUTF16Ptr(subKey))),
		0,
		0,
		regOptionNonVolatile,
		uintptr(keyRead|keyWrite|keyCreateSubKey),
		0,
		uintptr(unsafe.Pointer(&key)),
		0,
	)
	if ret != errorSuccess {
		return 0, syscall.Errno(ret)
	}

	return key, nil
}

func closeRegistryKey(key uintptr) {
	if key != 0 {
		procRegCloseKey.Call(key)
	}
}

// 値の中身は不要なので、存在確認だけを行う。
func hasRegistryValue(key uintptr, valueName string) bool {
	namePtr := toUTF16Ptr(valueName)

	ret, _, _ := procRegQueryValueExW.Call(
		key,
		uintptr(unsafe.Pointer(namePtr)),
		0,
		0,
		0,
		0,
	)
	return ret == errorSuccess
}

func queryRegistryDWORD(key uintptr, valueName string) (uint32, uint32, error) {
	namePtr := toUTF16Ptr(valueName)
	var valueType uint32
	var value uint32
	dataLen := uint32(unsafe.Sizeof(value))

	ret, _, _ := procRegQueryValueExW.Call(
		key,
		uintptr(unsafe.Pointer(namePtr)),
		0,
		uintptr(unsafe.Pointer(&valueType)),
		uintptr(unsafe.Pointer(&value)),
		uintptr(unsafe.Pointer(&dataLen)),
	)
	if ret != errorSuccess {
		return 0, 0, syscall.Errno(ret)
	}

	return value, valueType, nil
}

func setRegistryString(key uintptr, valueName, value string) error {
	data := syscall.StringToUTF16(value)
	ret, _, _ := procRegSetValueExW.Call(
		key,
		uintptr(unsafe.Pointer(toUTF16Ptr(valueName))),
		0,
		regSZ,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)*2),
	)
	if ret != errorSuccess {
		return syscall.Errno(ret)
	}
	return nil
}

func setRegistryDWORD(key uintptr, valueName string, value uint32) error {
	ret, _, _ := procRegSetValueExW.Call(
		key,
		uintptr(unsafe.Pointer(toUTF16Ptr(valueName))),
		0,
		regDWORD,
		uintptr(unsafe.Pointer(&value)),
		uintptr(unsafe.Sizeof(value)),
	)
	if ret != errorSuccess {
		return syscall.Errno(ret)
	}
	return nil
}

func deleteRegistryValue(key uintptr, valueName string) error {
	ret, _, _ := procRegDeleteValueW.Call(
		key,
		uintptr(unsafe.Pointer(toUTF16Ptr(valueName))),
	)
	if ret != errorSuccess {
		return syscall.Errno(ret)
	}
	return nil
}

func isRegistryNotFound(err error) bool {
	if err == nil {
		return false
	}

	errno, ok := err.(syscall.Errno)
	return ok && errno == errorFileNotFound
}
