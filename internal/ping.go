package internal

const (
	ipv4HeaderBytes    = 20
	ipv4TotalLengthMax = 65535
	icmpHeaderBytes    = 8
	payloadBytesMax    = ipv4TotalLengthMax - ipv4HeaderBytes - icmpHeaderBytes
)

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

type icmpHeader struct {
}

// CreateICMPEcho ICMP エコーメッセージのバイトスライスを作成する
func CreateICMPEcho(sequenceNumber uint16, payloadBytes uint) ([]byte, error) {
	if payloadBytes > payloadBytesMax {
		// err
	}

	message := make([]byte, icmpHeaderBytes+payloadBytes)

	// type
	message[0] = 0
	// code
	message[1] = 0
	// identifier
	message[4] = 0

	return message, nil
}
