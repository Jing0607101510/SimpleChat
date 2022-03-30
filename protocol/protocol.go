package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	SenderID = 4 // 发送者id
	MsgType  = 4 // 消息类型
	MsgLen   = 4 // 消息长度
)

const (
	NAME = iota
	ID
	MSG
)

func Enpack(senderId int, msgType int, msg []byte) []byte {
	var b []byte
	b = append(b, IntToBytes(senderId)...)
	b = append(b, IntToBytes(msgType)...)
	b = append(b, IntToBytes(len(msg))...)
	b = append(b, msg...)
	return b
}

func Depack(buffer []byte) ([]byte, bool, int, int, string) {
	headerLen := SenderID + MsgType + MsgLen
	if len(buffer) < headerLen {
		return buffer, false, 0, 0, ""
	} else {
		senderId := BytesToInt(buffer[:SenderID])
		msgType := BytesToInt(buffer[SenderID : SenderID+MsgType])
		msgLen := BytesToInt(buffer[SenderID+MsgType : headerLen])
		if len(buffer) < headerLen+msgLen {
			return buffer, false, 0, 0, ""
		} else {
			msg := string(buffer[headerLen : headerLen+msgLen])
			buffer = buffer[headerLen+msgLen:]
			return buffer, true, senderId, msgType, msg
		}
	}
}

func IntToBytes(n int) []byte {
	n32 := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, n32)
	return bytesBuffer.Bytes()
}

func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var n32 int32
	binary.Read(bytesBuffer, binary.BigEndian, &n32)
	return int(n32)
}
