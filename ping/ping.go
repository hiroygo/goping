package ping

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"reflect"
	"time"
)

const (
	ipv4HeaderMinBytes = 20
	ipv4TotalLengthMax = 65535
	icmpHeaderBytes    = 8
	// ICMPEchoDataMaxBytes ICMP ペイロードの最大バイト数
	ICMPEchoDataMaxBytes = ipv4TotalLengthMax - ipv4HeaderMinBytes - icmpHeaderBytes
)

type ICMPEchoHeader struct {
	Type           byte
	Code           byte
	Checksum       uint16
	Identifier     uint16
	SequenceNumber uint16
}

// EchoRequest と EchoReply を表す
type ICMPEchoMessage struct {
	ICMPEchoHeader
	Data []byte
}

func NewEchoRequest(identifier, sequenceNumber uint16, data []byte) *ICMPEchoMessage {
	e := &ICMPEchoMessage{
		ICMPEchoHeader: ICMPEchoHeader{
			Type: 8, Code: 0, Checksum: 0, Identifier: identifier, SequenceNumber: sequenceNumber,
		},
		Data: data,
	}
	return e
}

func MarshalEcho(e *ICMPEchoMessage) ([]byte, error) {
	if len(e.Data) > ICMPEchoDataMaxBytes {
		return nil, fmt.Errorf("ペイロードのサイズ %v は最大長 %v を超えています", len(e.Data), ICMPEchoDataMaxBytes)
	}

	b := make([]byte, icmpHeaderBytes+len(e.Data))
	b[0] = e.Type
	b[1] = e.Code

	// ネットワークバイトオーダに変換する
	// uint16 をビッグエンディアンで 2byte に変換する
	binary.BigEndian.PutUint16(b[4:6], e.Identifier)
	binary.BigEndian.PutUint16(b[6:8], e.SequenceNumber)
	for i := range e.Data {
		b[icmpHeaderBytes+i] = e.Data[i]
	}
	// b[2], b[3] はチェックサムなので最後に設定する
	binary.BigEndian.PutUint16(b[2:4], Checksum(b))

	return b, nil
}

func UnmarshalEcho(b []byte) (*ICMPEchoMessage, error) {
	if len(b) < icmpHeaderBytes {
		return nil, fmt.Errorf("バイト列のサイズ %v は最小長 %v を満たしていません", len(b), icmpHeaderBytes)
	}

	if sum := Checksum(b); sum != 0x0000 {
		return nil, fmt.Errorf("チェックサム 0x%x は再計算で 0x0000 になりません", sum)
	}

	e := &ICMPEchoMessage{}
	e.Type = b[0]
	e.Code = b[1]
	e.Checksum = binary.BigEndian.Uint16(b[2:4])
	e.Identifier = binary.BigEndian.Uint16(b[4:6])
	e.SequenceNumber = binary.BigEndian.Uint16(b[6:8])

	// ペイロードが存在すればコピーする
	if len(b) > icmpHeaderBytes {
		e.Data = append([]byte(nil), b[icmpHeaderBytes:]...)
	}

	return e, nil
}

// EchoRequest と EchoReply が返答として成立しているか調べる
// Type と Checksum は確認しない
func IsPair(request, reply *ICMPEchoMessage) bool {
	if request.Code != reply.Code {
		return false
	}

	if request.Identifier != reply.Identifier {
		return false
	}

	if request.SequenceNumber != reply.SequenceNumber {
		return false
	}

	if !reflect.DeepEqual(request.Data, reply.Data) {
		return false
	}

	return true
}

// チェックサムを計算する
// bytes はビッグエンディアンで並んでいること
func Checksum(b []byte) uint16 {
	var sum uint32

	// 16 ビットずつ走査していく
	for i := 0; i+1 < len(b); i += 2 {
		// 初めの 8 ビット分を足す
		sum += uint32(b[i]) << 8
		// 後の 8 ビット分を足す
		sum += uint32(b[i+1])
	}

	// ICMP の全体長が奇数の時は 0 埋めして末尾を 16 ビットとする
	if len(b)%2 != 0 {
		sum += uint32(b[len(b)-1]) << 8
	}

	// チェックサムは 16 ビットなのであふれた桁を取り出す
	overflowDigit := sum >> 16

	// あふれた分を消して、足す
	sum &= 0x0000FFFF
	sum += overflowDigit

	// ビット反転
	sum = ^sum

	return uint16(sum)
}

// ping を行い、RTT を返す
// identifier が 0 のとき、宛先によっては返答のチェックサムが再計算で 0x0000 にならない場合がある
func Do(remote net.Addr, timeout time.Duration, identifier, sequenceNumber, dataBytes uint16) (rtt time.Duration, rerr error) {
	// ペイロードはすべて 0 で作成する
	request := NewEchoRequest(identifier, sequenceNumber, make([]byte, dataBytes))
	writeData, err := MarshalEcho(request)
	if err != nil {
		return 0, fmt.Errorf("MarshalEcho error, %w", err)
	}

	/*
		接続を作成する
		DialIP ではなく、ListenPacket を使う
		DialIP で生成した場合は WriteTo ではなく、Write を使う
	*/
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return 0, fmt.Errorf("ListenPacket error, %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			rerr = err
		}
	}()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return 0, fmt.Errorf("SetDeadline error, %w", err)
	}

	// 送信
	requestTime := time.Now()
	if _, err := conn.WriteTo(writeData, remote); err != nil {
		return 0, fmt.Errorf("WriteTo error, %w", err)
	}

	// 受信
	readData := make([]byte, ipv4TotalLengthMax)
	readBytes, fromIP, err := conn.ReadFrom(readData)
	replyTime := time.Now()
	if err != nil {
		return 0, fmt.Errorf("ReadFrom error, %w", err)
	}
	if fromIP.String() != remote.String() {
		return 0, fmt.Errorf("パケット送信元 %v が リクエスト先 %v と異なります", fromIP, remote)
	}

	// 受信データを構造体にする
	readData = readData[:readBytes]
	reply, err := UnmarshalEcho(readData)
	if err != nil {
		return 0, fmt.Errorf("UnmarshalEcho error, %w", err)
	}

	if !IsPair(request, reply) {
		return 0, errors.New("ICMP フィールドが一致しません")
	}

	return replyTime.Sub(requestTime), nil
}
