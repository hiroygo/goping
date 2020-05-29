package main

import (
	"fmt"
	"goping/internal"
)

func getChecksum(bytes []byte) uint16 {
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

func main() {
	internal.Test()

	sum := getChecksum([]byte{0x08, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x0b, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68,
		0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x61,
		0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69})
	fmt.Printf("%x\n", sum)
}
