package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/hiroygo/goping/ping"
)

func getIPv4Addr(ipv4 string) (*net.IPAddr, error) {
	ip := net.ParseIP(ipv4)
	if ip == nil {
		msg := fmt.Sprintf("getIPv4Addr error:%v をパースできません。", ipv4)
		return nil, errors.New(msg)
	}

	return &net.IPAddr{IP: ip}, nil
}

func do(remoteIP string, timeout time.Duration, identifier uint16, sequenceNumber uint16, dataBytes uint16) (time.Duration, error) {
	// ペイロードはすべて 0 で作成する
	echoRequest := ping.NewEchoRequest(identifier, sequenceNumber, make([]byte, dataBytes))
	writeBytes, err := ping.MarshalEcho(echoRequest)
	if err != nil {
		msg := fmt.Sprintf("ping error:%v", err)
		return time.Duration(0), errors.New(msg)
	}

	// 接続先
	remoteAddr, err := getIPv4Addr(remoteIP)
	if err != nil {
		msg := fmt.Sprintf("ping error:%v", err)
		return time.Duration(0), errors.New(msg)
	}

	/*
		ソケットを生成する
		IP なので connect するわけではない
		ただし送信先情報は保存されるため WriteTo する必要はない
	*/
	conn, err := net.DialIP("ip4:icmp", nil, remoteAddr)
	if err != nil {
		msg := fmt.Sprintf("ping error:%v", err)
		return time.Duration(0), errors.New(msg)
	}
	defer conn.Close()

	// Write, ReadFrom のタイムアウトを設定する
	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		msg := fmt.Sprintf("ping error:%v", err)
		return time.Duration(0), errors.New(msg)
	}

	// 送信
	request := time.Now()
	{
		// WriteTo でもいいけど、DialIP で送信先を設定してるので Write を使う
		_, err := conn.Write(writeBytes)
		if err != nil {
			msg := fmt.Sprintf("ping error:%v", err)
			return time.Duration(0), errors.New(msg)
		}
	}

	// 受信
	readBytes := make([]byte, 1024)
	readSize, fromIP, err := conn.ReadFrom(readBytes)
	reply := time.Now()
	if err != nil {
		msg := fmt.Sprintf("ping error:%v", err)
		return time.Duration(0), errors.New(msg)
	}
	if fromIP.String() != remoteAddr.String() {
		msg := fmt.Sprintf("ping error:パケット送信元 %v が EchoRequest 先 %v と異なります。", fromIP, remoteAddr)
		return time.Duration(0), errors.New(msg)
	}

	// バッファをリサイズする
	readBytes = readBytes[:readSize]

	// 受信データを構造体にする
	echoReply := &ping.ICMPEchoMessage{}
	if err = ping.UnmarshalEcho(readBytes, echoReply); err != nil {
		msg := fmt.Sprintf("ping error:%v", err)
		return time.Duration(0), errors.New(msg)
	}

	if !ping.IsSameEchoField(echoRequest, echoReply) {
		return time.Duration(0), errors.New("ping error:EchoReply のフィールドが一致しません。")
	}

	return reply.Sub(request), nil
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("引数に宛先 ipv4 アドレスを入力してください。")
		return
	}

	fmt.Printf("Pinging %v\n", args[0])

	// FIXME:i == 0 のとき、宛先に関係なく "ping error:UnmarshalEcho error:チェックサム 65535 は再計算で 0x0000 になりません。" が発生する。
	for i := 1; i < 5; i++ {
		if duration, err := do(args[0], time.Second*5, uint16(i), uint16(i), 32); err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("返答を受信:RTT=%v ms\n", duration.Milliseconds())
		}

		time.Sleep(time.Second)
	}
}
