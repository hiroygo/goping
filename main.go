package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"time"

	"github.com/hiroygo/goping/ping"
)

func args() (string, time.Duration, uint16, uint16, error) {
	remoteName := flag.String("d", "", "送信先をホスト名または IPv4 アドレスで指定します")
	timeout := flag.Duration("w", time.Second*5, "タイムアウト時間を指定します")
	try := flag.Uint("c", 5, "実行回数を指定します")
	dataSize := flag.Uint("l", 32, "ペイロードのサイズをバイトで指定します")
	flag.Parse()

	if *remoteName == "" {
		return "", 0, 0, 0, errors.New("送信先が空です")
	}
	if *try > math.MaxUint16 {
		return "", 0, 0, 0, fmt.Errorf("試行回数 %v は不正です", *try)
	}
	if *dataSize > math.MaxUint16 {
		return "", 0, 0, 0, fmt.Errorf("ペイロードのサイズ %v は不正です", *dataSize)
	}

	return *remoteName, *timeout, uint16(*dataSize), uint16(*try), nil
}

func main() {
	remoteName, timeout, dataSize, try, err := args()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	remote, err := net.ResolveIPAddr("ip4", remoteName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", fmt.Errorf("ResolveIPAddr error: %w", err))
		os.Exit(1)
	}
	identifier := uint16(os.Getpid())

	fmt.Printf("PING %v (%v) %v(%v) bytes of data.\n", remoteName, remote, dataSize, dataSize+28)
	for i := uint16(0); i < try; i++ {
		rtt, err := ping.Do(remote, timeout, identifier, i, dataSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		} else {
			fmt.Printf("reply recv: seq=%v rtt=%v ms\n", i, rtt.Milliseconds())
		}
		time.Sleep(time.Second)
	}
	os.Exit(0)
}
