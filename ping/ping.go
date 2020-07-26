package ping

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
)

const (
	ipv4HeaderMinBytes = 20
	ipv4TotalLengthMax = 65535
	icmpHeaderBytes    = 8
	// ICMPEchoDataMaxBytes ICMP ペイロードの最大バイト数
	ICMPEchoDataMaxBytes = ipv4TotalLengthMax - ipv4HeaderMinBytes - icmpHeaderBytes
)

// ICMPEchoHeader EchoRequest と EchoReply 用の ICMP ヘッダ
type ICMPEchoHeader struct {
	Type           byte
	Code           byte
	Checksum       uint16
	Identifier     uint16
	SequenceNumber uint16
}

// ICMPEchoMessage EchoRequest と EchoReply を表す
type ICMPEchoMessage struct {
	ICMPEchoHeader
	Data []byte
}

// NewEchoRequest EchoRequest のポインタを返す
func NewEchoRequest(identifier, sequenceNumber uint16, data []byte) *ICMPEchoMessage {
	e := &ICMPEchoMessage{ICMPEchoHeader: ICMPEchoHeader{Type: 8, Code: 0, Checksum: 0, Identifier: identifier, SequenceNumber: sequenceNumber}, Data: data}
	return e
}

// MarshalEcho EchoRequest または EchoReply のバイトスライスを作成する
func MarshalEcho(echo *ICMPEchoMessage) ([]byte, error) {
	if echo == nil {
		return nil, errors.New("レシーバが nil です")
	}

	if len(echo.Data) > ICMPEchoDataMaxBytes {
		return nil, fmt.Errorf("ペイロードのサイズ %v は最大長 %v を超えています", len(echo.Data), ICMPEchoDataMaxBytes)
	}

	bytes := make([]byte, icmpHeaderBytes+len(echo.Data))
	bytes[0] = echo.Type
	bytes[1] = echo.Code
	// bytes[2], bytes[3] はチェックサムなのであとで設定する
	// uint16 をビッグエンディアンで 2byte に変換する
	binary.BigEndian.PutUint16(bytes[4:6], echo.Identifier)
	binary.BigEndian.PutUint16(bytes[6:8], echo.SequenceNumber)

	// ペイロードを設定する
	for i := range echo.Data {
		bytes[icmpHeaderBytes+i] = echo.Data[i]
	}

	// チェックサムを設定する
	binary.BigEndian.PutUint16(bytes[2:4], GetChecksum(bytes))

	return bytes, nil
}

// UnmarshalEcho バイトスライスから EchoRequest または EchoReply を作成する
func UnmarshalEcho(bytes []byte) (*ICMPEchoMessage, error) {
	if len(bytes) < icmpHeaderBytes {
		return nil, fmt.Errorf("バイト列のサイズ %v は最小長 %v を満たしていません", len(bytes), icmpHeaderBytes)
	}

	if checksum := GetChecksum(bytes); checksum != 0x0000 {
		return nil, fmt.Errorf("チェックサム %v は再計算で 0x0000 になりません", checksum)
	}

	echo := &ICMPEchoMessage{}
	echo.Type = bytes[0]
	echo.Code = bytes[1]
	echo.Checksum = binary.BigEndian.Uint16(bytes[2:4])
	echo.Identifier = binary.BigEndian.Uint16(bytes[4:6])
	echo.SequenceNumber = binary.BigEndian.Uint16(bytes[6:8])

	// ペイロードが存在すればコピーする
	if len(bytes) > icmpHeaderBytes {
		echo.Data = append([]byte(nil), bytes[icmpHeaderBytes:]...)
	}

	return echo, nil
}

// IsSameEchoField ICMPEchoMessage のフィールドが一致しているか確認する
// Type と Checksum は確認しない
func IsSameEchoField(echoRequest *ICMPEchoMessage, echoReply *ICMPEchoMessage) bool {
	if echoRequest == nil || echoReply == nil {
		return false
	}

	if echoRequest.Code != echoReply.Code {
		return false
	}

	if echoRequest.Identifier != echoReply.Identifier {
		return false
	}

	if echoRequest.SequenceNumber != echoReply.SequenceNumber {
		return false
	}

	if !reflect.DeepEqual(echoRequest.Data, echoReply.Data) {
		return false
	}

	return true
}

// GetChecksum チェックサムを計算する
// bytes はビッグエンディアンで並んでいること
func GetChecksum(bytes []byte) uint16 {
	var ret uint32

	// 16 ビットずつ走査していく
	for i := 0; i+1 < len(bytes); i += 2 {
		// 初めの 8 ビット分を足す
		ret += uint32(bytes[i]) << 8
		// 後の 8 ビット分を足す
		ret += uint32(bytes[i+1])
	}

	// ICMP の全体長が奇数の時は 0 埋めして末尾を 16 ビットとする
	if len(bytes)%2 != 0 {
		ret += uint32(bytes[len(bytes)-1]) << 8
	}

	// チェックサムは 16 ビットなのであふれた桁を取り出す
	overflowDigit := ret >> 16

	// あふれた分を消して、足す
	ret &= 0x0000FFFF
	ret += overflowDigit

	// ビット反転
	ret = ^ret

	return uint16(ret)
}
