package ping_test

import (
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/hiroygo/goping/ping"
)

func TestNewEchoRequest(t *testing.T) {
	identifier := uint16(1)
	sequenceNumber := uint16(2)
	data := []byte{0}
	expected := &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 8, Code: 0, Checksum: 0, Identifier: identifier, SequenceNumber: sequenceNumber}, Data: data}
	if actual := ping.NewEchoRequest(identifier, sequenceNumber, data); !reflect.DeepEqual(actual, expected) {
		t.Errorf("want NewEchoRequest(%v, %v, %v) = %v, got %v", identifier, sequenceNumber, data, expected, actual)
	}
}

func TestMarshalEcho(t *testing.T) {
	cases := []struct {
		name        string
		echoRequest *ping.ICMPEchoMessage
		expected    []byte
	}{
		{
			name: "ペイロードが存在する",
			echoRequest: &ping.ICMPEchoMessage{
				ICMPEchoHeader: ping.ICMPEchoHeader{Type: 8, Code: 0, Checksum: 0, Identifier: 1, SequenceNumber: 11},
				Data: []byte{
					0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a,
					0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74,
					0x75, 0x76, 0x77, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67,
					0x68, 0x69},
			},
			expected: []byte{
				0x08, 0x00, 0x4d, 0x50, 0x00, 0x01, 0x00, 0x0b, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68,
				0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x61,
				0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69,
			},
		},
		{
			name: "ペイロードが存在しない",
			echoRequest: &ping.ICMPEchoMessage{
				ICMPEchoHeader: ping.ICMPEchoHeader{Type: 8, Code: 0, Checksum: 0, Identifier: 1, SequenceNumber: 11},
				Data:           nil,
			},
			expected: []byte{
				0x08, 0x00, 0xf7, 0xf3, 0x00, 0x01, 0x00, 0x0b,
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual, err := ping.MarshalEcho(c.echoRequest)
			if err != nil {
				t.Fatalf("MarshalEcho error %v", err)
			}

			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("want MarshalEcho(%v) = %v, got %v", c.echoRequest, c.expected, actual)
			}
		})
	}
}

func TestUnmarshalEcho(t *testing.T) {
	cases := []struct {
		name      string
		marshaled []byte
		expected  *ping.ICMPEchoMessage
	}{
		{
			name: "ペイロードが存在する",
			marshaled: []byte{
				0x08, 0x00, 0x4f, 0xfe, 0xff, 0x1e, 0xfe, 0x3e, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68,
				0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x61,
				0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69,
			},
			expected: &ping.ICMPEchoMessage{
				ICMPEchoHeader: ping.ICMPEchoHeader{Type: 8, Code: 0, Checksum: 0x4ffe, Identifier: 0xff1e, SequenceNumber: 0xfe3e},
				Data: []byte{
					0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a,
					0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74,
					0x75, 0x76, 0x77, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67,
					0x68, 0x69},
			},
		},
		{
			name: "ペイロードが存在しない",
			marshaled: []byte{
				0x08, 0x00, 0xf7, 0xf3, 0x00, 0x01, 0x00, 0x0b,
			},
			expected: &ping.ICMPEchoMessage{
				ICMPEchoHeader: ping.ICMPEchoHeader{Type: 8, Code: 0, Checksum: 0xf7f3, Identifier: 1, SequenceNumber: 11},
				Data:           nil,
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual, err := ping.UnmarshalEcho(c.marshaled)
			if err != nil {
				t.Fatalf("UnmarshalEcho error %v", err)
			}

			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("want UnmarshalEcho(%v) = %v, got %v", c.marshaled, c.expected, actual)
			}
		})
	}
}

func TestPair(t *testing.T) {
	cases := []struct {
		name        string
		echoRequest *ping.ICMPEchoMessage
		echoReply   *ping.ICMPEchoMessage
		wantErr     bool
	}{
		{
			name:        "Code が異なる",
			echoRequest: &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Code: 1, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			echoReply:   &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Code: 0, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			wantErr:     true,
		},
		{
			name:        "Identifier が異なる",
			echoRequest: &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Code: 1, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			echoReply:   &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Code: 1, Identifier: 0, SequenceNumber: 3}, Data: []byte{0}},
			wantErr:     true,
		},
		{
			name:        "SequenceNumber が異なる",
			echoRequest: &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Code: 1, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			echoReply:   &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Code: 1, Identifier: 2, SequenceNumber: 0}, Data: []byte{0}},
			wantErr:     true,
		},
		{
			name:        "Data が異なる",
			echoRequest: &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Code: 1, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			echoReply:   &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Code: 1, Identifier: 2, SequenceNumber: 3}, Data: []byte{0, 0}},
			wantErr:     true,
		},
		{
			name:        "一致する",
			echoRequest: &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Code: 1, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			echoReply:   &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Code: 1, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			wantErr:     false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := ping.Pair(c.echoRequest, c.echoReply)
			if c.wantErr && actual == nil || !c.wantErr && actual != nil {
				t.Errorf("wantErr Pair(%v, %v) = %v, got %v", c.echoRequest, c.echoReply, c.wantErr, actual)
			}
		})
	}
}

func TestChecksum(t *testing.T) {
	cases := []struct {
		name     string
		input    []byte
		expected uint16
	}{
		{
			name: "バイトスライス中のチェックサムが 0x00 0x00 に設定されているとき = ICMP を送信するとき",
			input: []byte{
				0x08, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x0b, 0x61, 0x62, 0x63,
				0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e,
				0x6f, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x61, 0x62,
				0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69,
			},
			expected: 0x4d50,
		},
		{
			name: "バイトスライス中のチェックサムがすでに設定されているとき = 受信したチェックサムは計算すると 0x00 0x00 になる",
			input: []byte{
				0x08, 0x00, 0x4d, 0x50, 0x00, 0x01, 0x00, 0x0b, 0x61, 0x62, 0x63,
				0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e,
				0x6f, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x61, 0x62,
				0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69,
			},
			expected: 0x0000,
		},
		{
			name: "1 バイトのとき",
			input: []byte{
				0x00,
			},
			expected: 0xFFFF,
		},
		{
			name:     "バイトスライスが nil のとき",
			input:    nil,
			expected: 0xFFFF,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if actual := ping.Checksum(c.input); actual != c.expected {
				t.Errorf("want Checksum(%v) = %v, got %v", c.input, c.expected, actual)
			}
		})
	}
}

func TestDo(t *testing.T) {
	cases := []struct {
		name           string
		ip4Remote      *net.IPAddr
		timeout        time.Duration
		identifier     uint16
		sequenceNumber uint16
		dataSize       uint16
		wantErr        bool
	}{
		{
			name:           "ローカルループバックアドレス",
			ip4Remote:      &net.IPAddr{IP: net.ParseIP("127.0.0.1")},
			timeout:        time.Second,
			identifier:     1,
			sequenceNumber: 1,
			dataSize:       1,
			wantErr:        false,
		},
		{
			name:           "例示用アドレス",
			ip4Remote:      &net.IPAddr{IP: net.ParseIP("192.0.2.1")},
			timeout:        time.Millisecond * 500,
			identifier:     1,
			sequenceNumber: 1,
			dataSize:       1,
			wantErr:        true,
		},
	}

	// Linux 環境では "socket: operation not permitted" が発生する場合がある
	// この場合 sudo してテストする
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 5; i++ {
				_, err := ping.Do(c.ip4Remote, c.timeout, c.identifier, c.sequenceNumber, c.dataSize)
				if c.wantErr && err == nil || !c.wantErr && err != nil {
					t.Errorf("wantErr Do(%v, %v, %v, %v, %v) = %v, got %v", c.ip4Remote, c.timeout, c.identifier, c.sequenceNumber, c.dataSize, c.wantErr, err)
				}
			}
		})
	}
}
