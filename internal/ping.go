package internal

import "fmt"

const (
	ipv4HeaderBytes    = 20
	ipv4TotalLengthMax = 65535
	icmpHeaderBytes    = 8
	payloadBytesMax    = ipv4TotalLengthMax - ipv4HeaderBytes - icmpHeaderBytes
)

// Test shows "PING.GO!"
func Test() {
	fmt.Println("PING.GO!")
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
