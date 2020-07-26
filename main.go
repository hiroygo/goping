package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hiroygo/goping/ping"
)

// TODO: linux だと実行に特権レベルが必要になる
func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "引数に宛先 ipv4 アドレスを入力してください\n")
		os.Exit(1)
	}

	ipv4 := args[0]
	fmt.Printf("Pinging %v\n", ipv4)

	// TODO:i == 0 のとき、宛先に関係なく "ping error:UnmarshalEcho error:チェックサム 65535 は再計算で 0x0000 になりません。" が発生することがある
	for i := 1; i < 5; i++ {
		if duration, err := ping.Do(ipv4, time.Second*5, uint16(i), uint16(i), 32); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		} else {
			fmt.Printf("返答を受信:RTT=%v ms\n", duration.Milliseconds())
		}
		time.Sleep(time.Second)
	}

	os.Exit(0)
}
