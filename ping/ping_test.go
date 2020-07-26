package ping_test

import (
	"reflect"
	"testing"

	"github.com/hiroygo/goping/ping"
)

func TestNewEchoRequest(t *testing.T) {
	identifier := uint16(1)
	sequenceNumber := uint16(2)
	data := []byte{0}
	expected := ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 8, Code: 0, Checksum: 0, Identifier: identifier, SequenceNumber: sequenceNumber}, Data: data}
	if actual := ping.NewEchoRequest(identifier, sequenceNumber, data); !reflect.DeepEqual(*actual, expected) {
		t.Errorf("want NewEchoRequest(%v, %v, %v) = %v, got %v", identifier, sequenceNumber, data, expected, actual)
	}
}

func TestMarshalEcho(t *testing.T) {
	expected := []byte{
		0x08, 0x00, 0x4d, 0x50, 0x00, 0x01, 0x00, 0x0b, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68,
		0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x61,
		0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69,
	}

	echoRequest := ping.ICMPEchoMessage{
		ICMPEchoHeader: ping.ICMPEchoHeader{Type: 8, Code: 0, Checksum: 0, Identifier: 1, SequenceNumber: 11},
		Data: []byte{
			0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a,
			0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74,
			0x75, 0x76, 0x77, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67,
			0x68, 0x69},
	}
	actual, err := ping.MarshalEcho(&echoRequest)
	if err != nil {
		t.Fatalf("MarshalEcho error %v", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("want MarshalEcho(%v) = %v, got %v", echoRequest, expected, actual)
	}
}

func TestUnmarshalEcho(t *testing.T) {
	expected := ping.ICMPEchoMessage{
		ICMPEchoHeader: ping.ICMPEchoHeader{Type: 8, Code: 0, Checksum: 0x4ffe, Identifier: 0xff1e, SequenceNumber: 0xfe3e},
		Data: []byte{
			0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a,
			0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74,
			0x75, 0x76, 0x77, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67,
			0x68, 0x69},
	}

	marshaled := []byte{
		0x08, 0x00, 0x4f, 0xfe, 0xff, 0x1e, 0xfe, 0x3e, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68,
		0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x61,
		0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69,
	}
	actual, err := ping.UnmarshalEcho(marshaled)
	if err != nil {
		t.Fatalf("UnmarshalEcho error %v", err)
	}

	if !reflect.DeepEqual(*actual, expected) {
		t.Errorf("want UnmarshalEcho(%v) = %v, got %v", marshaled, expected, actual)
	}
}

func TestIsSameEchoField(t *testing.T) {
	cases := []struct {
		name        string
		echoRequest *ping.ICMPEchoMessage
		echoReply   *ping.ICMPEchoMessage
		expected    bool
	}{
		{
			name:        "Code が異なる",
			echoRequest: &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 0, Code: 1, Checksum: 0, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			echoReply:   &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 0, Code: 0, Checksum: 0, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			expected:    false,
		},
		{
			name:        "Identifier が異なる",
			echoRequest: &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 0, Code: 1, Checksum: 0, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			echoReply:   &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 0, Code: 1, Checksum: 0, Identifier: 0, SequenceNumber: 3}, Data: []byte{0}},
			expected:    false,
		},
		{
			name:        "SequenceNumber が異なる",
			echoRequest: &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 0, Code: 1, Checksum: 0, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			echoReply:   &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 0, Code: 1, Checksum: 0, Identifier: 2, SequenceNumber: 0}, Data: []byte{0}},
			expected:    false,
		},
		{
			name:        "Data が異なる",
			echoRequest: &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 0, Code: 1, Checksum: 0, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			echoReply:   &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 0, Code: 1, Checksum: 0, Identifier: 2, SequenceNumber: 3}, Data: []byte{0, 0}},
			expected:    false,
		},
		{
			name:        "一致する",
			echoRequest: &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 0, Code: 1, Checksum: 0, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			echoReply:   &ping.ICMPEchoMessage{ICMPEchoHeader: ping.ICMPEchoHeader{Type: 0, Code: 1, Checksum: 0, Identifier: 2, SequenceNumber: 3}, Data: []byte{0}},
			expected:    true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if actual := ping.IsSameEchoField(c.echoRequest, c.echoReply); actual != c.expected {
				t.Errorf("want IsSameEchoField(%v, %v) = %v, got %v", c.echoRequest, c.echoReply, c.expected, actual)
			}
		})
	}
}

func TestGetChecksum(t *testing.T) {
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
			if actual := ping.GetChecksum(c.input); actual != c.expected {
				t.Errorf("want GetChecksum(%v) = %v, got %v", c.input, c.expected, actual)
			}
		})
	}
}
