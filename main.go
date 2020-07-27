package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hiroygo/goping/ping"
)

func identifier() uint16 {
	rand.Seed(time.Now().UnixNano())
	return uint16(rand.Intn(math.MaxUint16) + 1)
}

func parseArgs() (ipv4 string, timeoutSec, try int, err error) {
	flag.StringVar(&ipv4, "d", "", "送信先を指定します")
	flag.IntVar(&timeoutSec, "t", 1, "タイムアウト時間を秒で指定します")
	flag.IntVar(&try, "c", 5, "試行回数を指定します")
	flag.Parse()

	if ipv4 == "" {
		return "", 0, 0, errors.New("送信先が空です")
	}
	if timeoutSec < 1 {
		return "", 0, 0, fmt.Errorf("タイムアウト時間 %v は不正です", timeoutSec)
	}
	if try < 1 {
		return "", 0, 0, fmt.Errorf("試行回数 %v は不正です", try)
	}

	return
}

func main() {
	ipv4, timeoutSec, try, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	identifier := identifier()
	fmt.Printf("Pinging %v\n", ipv4)
	for i := 0; i < try; i++ {
		if duration, err := ping.Do(ipv4, time.Second*time.Duration(timeoutSec), identifier, uint16(i), 32); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		} else {
			fmt.Printf("返答を受信:RTT=%v ms\n", duration.Milliseconds())
		}
		time.Sleep(time.Second)
	}

	os.Exit(0)
}
