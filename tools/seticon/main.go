//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	rtIcon      = 3
	rtGroupIcon = 14
)

var (
	modKernel32             = syscall.NewLazyDLL("kernel32.dll")
	procBeginUpdateResourceW = modKernel32.NewProc("BeginUpdateResourceW")
	procUpdateResourceW      = modKernel32.NewProc("UpdateResourceW")
	procEndUpdateResourceW   = modKernel32.NewProc("EndUpdateResourceW")
)

type iconEntry struct {
	width       byte
	height      byte
	colorCount  byte
	reserved    byte
	planes      uint16
	bitCount    uint16
	bytesInRes  uint32
	imageOffset uint32
}

func main() {
	if len(os.Args) != 3 {
		exitf("usage: go run ./tools/seticon <exe> <ico>")
	}

	exePath := os.Args[1]
	iconPath := os.Args[2]

	iconFile, err := os.ReadFile(iconPath)
	if err != nil {
		exitf("read ico failed: %v", err)
	}

	entries, err := parseIcon(iconFile)
	if err != nil {
		exitf("parse ico failed: %v", err)
	}

	if err := applyIconResources(exePath, iconFile, entries); err != nil {
		exitf("set exe icon failed: %v", err)
	}
}

func parseIcon(iconFile []byte) ([]iconEntry, error) {
	if len(iconFile) < 6 {
		return nil, fmt.Errorf("ico header too short")
	}

	if binary.LittleEndian.Uint16(iconFile[0:2]) != 0 || binary.LittleEndian.Uint16(iconFile[2:4]) != 1 {
		return nil, fmt.Errorf("invalid ico header")
	}

	count := int(binary.LittleEndian.Uint16(iconFile[4:6]))
	if count == 0 {
		return nil, fmt.Errorf("ico has no images")
	}

	entries := make([]iconEntry, 0, count)
	for i := 0; i < count; i++ {
		offset := 6 + i*16
		if offset+16 > len(iconFile) {
			return nil, fmt.Errorf("ico entry %d truncated", i)
		}

		entry := iconEntry{
			width:       iconFile[offset],
			height:      iconFile[offset+1],
			colorCount:  iconFile[offset+2],
			reserved:    iconFile[offset+3],
			planes:      binary.LittleEndian.Uint16(iconFile[offset+4 : offset+6]),
			bitCount:    binary.LittleEndian.Uint16(iconFile[offset+6 : offset+8]),
			bytesInRes:  binary.LittleEndian.Uint32(iconFile[offset+8 : offset+12]),
			imageOffset: binary.LittleEndian.Uint32(iconFile[offset+12 : offset+16]),
		}

		end := uint64(entry.imageOffset) + uint64(entry.bytesInRes)
		if entry.bytesInRes == 0 || end > uint64(len(iconFile)) {
			return nil, fmt.Errorf("ico entry %d has invalid bounds", i)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func applyIconResources(exePath string, iconFile []byte, entries []iconEntry) error {
	updateHandle, _, err := procBeginUpdateResourceW.Call(
		uintptr(unsafe.Pointer(toUTF16Ptr(exePath))),
		0,
	)
	if updateHandle == 0 {
		return fmt.Errorf("BeginUpdateResourceW failed: %v", err)
	}

	commit := false
	defer func() {
		discard := uintptr(1)
		if commit {
			discard = 0
		}
		procEndUpdateResourceW.Call(updateHandle, discard)
	}()

	for i, entry := range entries {
		imageData := iconFile[entry.imageOffset : entry.imageOffset+entry.bytesInRes]
		if len(imageData) == 0 {
			return fmt.Errorf("icon image %d empty", i)
		}

		resID := uintptr(i + 1)
		r1, _, updateErr := procUpdateResourceW.Call(
			updateHandle,
			makeIntResource(rtIcon),
			makeIntResource(resID),
			0,
			uintptr(unsafe.Pointer(&imageData[0])),
			uintptr(len(imageData)),
		)
		if r1 == 0 {
			return fmt.Errorf("UpdateResourceW RT_ICON %d failed: %v", i, updateErr)
		}
	}

	groupData := buildGroupIcon(entries)
	r1, _, updateErr := procUpdateResourceW.Call(
		updateHandle,
		makeIntResource(rtGroupIcon),
		makeIntResource(1),
		0,
		uintptr(unsafe.Pointer(&groupData[0])),
		uintptr(len(groupData)),
	)
	if r1 == 0 {
		return fmt.Errorf("UpdateResourceW RT_GROUP_ICON failed: %v", updateErr)
	}

	commit = true
	return nil
}

func buildGroupIcon(entries []iconEntry) []byte {
	groupData := make([]byte, 6+len(entries)*14)
	binary.LittleEndian.PutUint16(groupData[0:2], 0)
	binary.LittleEndian.PutUint16(groupData[2:4], 1)
	binary.LittleEndian.PutUint16(groupData[4:6], uint16(len(entries)))

	for i, entry := range entries {
		offset := 6 + i*14
		groupData[offset] = entry.width
		groupData[offset+1] = entry.height
		groupData[offset+2] = entry.colorCount
		groupData[offset+3] = entry.reserved
		binary.LittleEndian.PutUint16(groupData[offset+4:offset+6], entry.planes)
		binary.LittleEndian.PutUint16(groupData[offset+6:offset+8], entry.bitCount)
		binary.LittleEndian.PutUint32(groupData[offset+8:offset+12], entry.bytesInRes)
		binary.LittleEndian.PutUint16(groupData[offset+12:offset+14], uint16(i+1))
	}

	return groupData
}

func makeIntResource(id uintptr) uintptr {
	return id
}

func toUTF16Ptr(value string) *uint16 {
	ptr, _ := syscall.UTF16PtrFromString(value)
	return ptr
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
