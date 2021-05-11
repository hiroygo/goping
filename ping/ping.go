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
	// 通常は 20 バイト
	ipv4HeaderMinSize = 20
	ipv4TotalMaxSize  = 65535
	icmpHeaderSize    = 8
	// ICMP ペイロードの最大バイト数
	MaxDataSize = ipv4TotalMaxSize - ipv4HeaderMinSize - icmpHeaderSize
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
	return &ICMPEchoMessage{
		ICMPEchoHeader: ICMPEchoHeader{
			Type: 8, Code: 0, Checksum: 0, Identifier: identifier, SequenceNumber: sequenceNumber,
		},
		Data: data,
	}
}

func MarshalEcho(e *ICMPEchoMessage) ([]byte, error) {
	if len(e.Data) > MaxDataSize {
		return nil, fmt.Errorf("ペイロードのサイズ %v が %v を超えています", len(e.Data), MaxDataSize)
	}

	b := make([]byte, icmpHeaderSize+len(e.Data))
	b[0] = e.Type
	b[1] = e.Code
	// ネットワークバイトオーダに変換する
	// uint16 をビッグエンディアンで 2byte に変換する
	binary.BigEndian.PutUint16(b[4:6], e.Identifier)
	binary.BigEndian.PutUint16(b[6:8], e.SequenceNumber)
	for i := range e.Data {
		b[icmpHeaderSize+i] = e.Data[i]
	}
	// b[2], b[3] はチェックサムなので最後に設定する
	binary.BigEndian.PutUint16(b[2:4], Checksum(b))

	return b, nil
}

func icmpType(b []byte) (byte, error) {
	if len(b) < icmpHeaderSize {
		return 0, fmt.Errorf("バイト列のサイズ %v は最小長 %v を満たしていません", len(b), icmpHeaderSize)
	}
	return b[0], nil
}

func UnmarshalEcho(b []byte) (*ICMPEchoMessage, error) {
	if len(b) < icmpHeaderSize {
		return nil, fmt.Errorf("バイト列のサイズ %v は最小長 %v を満たしていません", len(b), icmpHeaderSize)
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
	if len(b) > icmpHeaderSize {
		e.Data = append([]byte(nil), b[icmpHeaderSize:]...)
	}

	return e, nil
}

// EchoRequest と EchoReply が対になっているか調べる
// Type と Checksum は確認しない
func Pair(request, reply *ICMPEchoMessage) error {
	if request.Code != reply.Code {
		return errors.New("Code が一致しません")
	}

	if request.Identifier != reply.Identifier {
		return errors.New("Identifier が一致しません")
	}

	if request.SequenceNumber != reply.SequenceNumber {
		return errors.New("SequenceNumber が一致しません")
	}

	if !reflect.DeepEqual(request.Data, reply.Data) {
		return errors.New("Data が一致しません")
	}

	return nil
}

// ICMP チェックサムを計算する
// b はビッグエンディアンで並んでいること
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

// IPv4 先に Ping する
func Do(ip4Remote *net.IPAddr, timeout time.Duration, identifier, sequenceNumber, dataSize uint16) (rtt time.Duration, rerr error) {
	// DialIP では送信先からのパケットは受信できるが他のマシンからの
	// パケットは受信できない。つまり Destination Unreachable などが受信できない
	conn, err := net.ListenIP("ip4:icmp", nil)
	if err != nil {
		return 0, fmt.Errorf("ListenIP error: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			rerr = err
		}
	}()

	// 送信する
	// ペイロードはすべて 0 で作成する
	request := NewEchoRequest(identifier, sequenceNumber, make([]byte, dataSize))
	sendData, err := MarshalEcho(request)
	if err != nil {
		return 0, fmt.Errorf("MarshalEcho error: %w", err)
	}
	if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return 0, fmt.Errorf("SetWriteDeadline error: %w", err)
	}
	start := time.Now()
	if _, err := conn.WriteToIP(sendData, ip4Remote); err != nil {
		return 0, fmt.Errorf("WriteToIP error: %w", err)
	}

	// 受信する
	timeoutCh := time.After(timeout)
	for {
		select {
		case <-timeoutCh:
			return 0, errors.New("返答受信がタイムアウトしました")
		default:
			recvData := make([]byte, ipv4TotalMaxSize)
			if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
				return 0, fmt.Errorf("SetReadDeadline error: %w", err)
			}
			recvSize, recvFrom, err := conn.ReadFromIP(recvData)
			end := time.Now()
			if err != nil {
				return 0, fmt.Errorf("ReadFromIP error: %w", err)
			}

			recvData = recvData[:recvSize]
			t, err := icmpType(recvData)
			if err != nil {
				continue
			}
			// Linux では localhost などに ping した場合
			// 自分が送った EchoRequest(8) を受信する
			// その後、カーネルのプロトコルモジュールから EchoReply を受信したりする
			if t == 8 {
				continue
			}
			if t == 0 && recvFrom.IP.Equal(ip4Remote.IP) {
				reply, err := UnmarshalEcho(recvData)
				if err != nil {
					continue
				}
				if err := Pair(request, reply); err != nil {
					continue
				}
				return end.Sub(start), nil
			}
			if t == 3 {
				return 0, fmt.Errorf("From %v Destination Unreachable", recvFrom.IP.String())
			}
			// 本当は他の Type も処理すべき
		}
	}
}
