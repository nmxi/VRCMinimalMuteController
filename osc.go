//go:build windows

package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"sync/atomic"
	"time"
)

func sendOscSequence() {
	address, err := net.ResolveUDPAddr("udp4", "127.0.0.1:9000")
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp4", nil, address)
	if err != nil {
		return
	}
	defer conn.Close()

	sendOscInt(conn, "/input/Voice", 0)
	time.Sleep(32 * time.Millisecond)
	sendOscInt(conn, "/input/Voice", 1)
	time.Sleep(32 * time.Millisecond)
	sendOscInt(conn, "/input/Voice", 0)
}

// 送信中の連打でシーケンスが重ならないように抑止する。
func triggerOscSequence() {
	if !atomic.CompareAndSwapInt32(&oscSequenceRunning, 0, 1) {
		return
	}

	go func() {
		defer atomic.StoreInt32(&oscSequenceRunning, 0)
		sendOscSequence()
	}()
}

func sendOscInt(conn *net.UDPConn, address string, value int32) {
	packet := buildOscMessage(address, value)
	_, _ = conn.Write(packet)
}

func buildOscMessage(address string, value int32) []byte {
	var buffer bytes.Buffer
	writeOscString(&buffer, address)
	writeOscString(&buffer, ",i")
	_ = binary.Write(&buffer, binary.BigEndian, value)
	return buffer.Bytes()
}

func writeOscString(buffer *bytes.Buffer, value string) {
	buffer.WriteString(value)
	buffer.WriteByte(0)
	for buffer.Len()%4 != 0 {
		buffer.WriteByte(0)
	}
}
