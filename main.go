package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/hiroygo/goping/ping"
)

func randomIdentifier() uint16 {
	rand.Seed(time.Now().UnixNano())
	return uint16(rand.Intn(math.MaxUint16) + 1)
}

func parseIPv4Addr(ipv4 string) (*net.IPAddr, error) {
	ip := net.ParseIP(ipv4)
	if ip == nil {
		return nil, fmt.Errorf("ParseIP error, %v をパースできません", ipv4)
	}
	return &net.IPAddr{IP: ip}, nil
}

func parseArgs() (string, time.Duration, uint16, uint16, error) {
	ipv4 := flag.String("d", "", "IPv4送信先を指定します")
	timeout := flag.Duration("w", time.Second, "タイムアウト時間を指定します")
	try := flag.Uint("c", 5, "実行回数を指定します")
	dataBytes := flag.Uint("l", 32, "ペイロードのサイズをバイトで指定します")
	flag.Parse()

	if *ipv4 == "" {
		return "", 0, 0, 0, errors.New("送信先が空です")
	}
	if *try > math.MaxUint16 {
		return "", 0, 0, 0, fmt.Errorf("試行回数 %v は不正です", *try)
	}
	if *dataBytes > math.MaxUint16 {
		return "", 0, 0, 0, fmt.Errorf("ペイロードのサイズ %v は不正です", *dataBytes)
	}

	return *ipv4, *timeout, uint16(*dataBytes), uint16(*try), nil
}

func main() {
	ipv4, timeoutSec, dataBytes, try, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	remoteAddr, err := parseIPv4Addr(ipv4)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	// FIXME: pid にする
	identifier := randomIdentifier()

	fmt.Printf("PING %v\n", remoteAddr.String())
	for i := uint16(0); i < try; i++ {
		if rtt, err := ping.Do(remoteAddr, timeoutSec, identifier, i, dataBytes); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		} else {
			fmt.Printf("recv reply: rtt=%v ms\n", rtt.Milliseconds())
		}
		time.Sleep(time.Second)
	}

	os.Exit(0)
}
