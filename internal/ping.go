package internal

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	icmpHeaderBytes = 8
)

// ICMPEchoHeader Echo と Echo Reply のICMPヘッダ
type ICMPEchoHeader struct {
	Type           byte
	Code           byte
	Checksum       uint16
	Identifier     uint16
	SequenceNumber uint16
}

// ICMPEchoMessage Echo/Echo Reply 構造体
type ICMPEchoMessage struct {
	ICMPEchoHeader
	Data []byte
}

func marshal(echo *ICMPEchoMessage) []byte {
	bytes := make([]byte, icmpHeaderBytes+len(echo.Data))

	bytes[0] = echo.Type
	bytes[1] = echo.Code
	// bytes[2], bytes[3] はチェックサムなのであとで設定する
	// uint16 をビッグエンディアンで byte に変換する
	binary.BigEndian.PutUint16(bytes[4:6], echo.Identifier)
	binary.BigEndian.PutUint16(bytes[6:8], echo.SequenceNumber)

	// ペイロードを設定する
	for i := range echo.Data {
		bytes[i+icmpHeaderBytes] = echo.Data[i]
	}

	// チェックサムを定する
	binary.BigEndian.PutUint16(bytes[2:4], GetChecksum(bytes))

	return bytes
}

// Unmarshal Echo または Echo Reply のバイトスライスから構造体を作成する
func Unmarshal(bytes []byte, echo *ICMPEchoMessage) error {
	if len(bytes) < icmpHeaderBytes {
		msg := fmt.Sprintf("Unmarshal error:バイト列の長さ %d が不正です。", len(bytes))
		return errors.New(msg)
	}

	echo.Type = bytes[0]
	echo.Code = bytes[1]
	echo.Checksum = binary.BigEndian.Uint16(bytes[2:4])
	echo.Identifier = binary.BigEndian.Uint16(bytes[4:6])
	echo.SequenceNumber = binary.BigEndian.Uint16(bytes[6:8])

	// ペイロードが存在すればコピーする
	if len(bytes) > icmpHeaderBytes {
		echo.Data = append([]byte(nil), bytes[icmpHeaderBytes:]...)
	}

	return nil
}

// MarshalEcho ICMP Echo メッセージのバイトスライスを作成する
func MarshalEcho(identifier uint16, sequenceNumber uint16, data []byte) []byte {
	echo := ICMPEchoMessage{ICMPEchoHeader: ICMPEchoHeader{Type: 8, Code: 0, Checksum: 0, Identifier: identifier, SequenceNumber: sequenceNumber}, Data: data}
	return marshal(&echo)
}

// GetChecksum ICMPチェックサムを返す
func GetChecksum(bytes []byte) uint16 {
	var ret uint32

	// 16ビットずつ走査していく
	for i := 0; i+1 < len(bytes); i += 2 {
		// 初めのバイトを16ビット用に変換
		ret += uint32(bytes[i]) << 8
		// 後のバイトを16ビット用に変換
		ret += uint32(bytes[i+1])
	}

	// ICMPの全体長が奇数の時は0埋めして末尾16ビットとする
	if len(bytes)%2 != 0 {
		ret += uint32(bytes[len(bytes)-1]) << 8
	}

	// チェックサムは16ビットなのであふれた桁を計算する
	overflowDigit := ret >> 16

	// あふれた分を消して、足す
	ret &= 0x0000FFFF
	ret += overflowDigit

	// ビット反転
	ret = ^ret

	return uint16(ret)
}
