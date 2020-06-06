package main

import (
	"errors"
	"fmt"
	"goping/internal"
	"net"
	"time"
)

func readn(conn net.Conn, timeout time.Duration, n int) ([]byte, error) {
	if n < 1 {
		msg := fmt.Sprintf("readn error:%d は不正です。", n)
		return nil, errors.New(msg)
	}

	var results []byte
	start := time.Now()
	for {
		var buff []byte
		readLen, err := conn.Read(buff)
		if err != nil {
			msg := fmt.Sprintf("readn error:%v", err)
			return nil, errors.New(msg)
		}

		if time.Since(start) >= timeout {
			return nil, errors.New("readn error:timeout")
		}

		results = append(results, buff...)
		n -= readLen
		if n == 0 {
			return results, nil
		}
	}
}

func ping(destIP string, timeout time.Duration, identifier uint16, sequenceNumber uint16, dataBytes uint16) (time.Duration, error) {
	// ペイロードはすべて 0 になる
	echoRequest := internal.NewEchoRequest(identifier, sequenceNumber, make([]byte, dataBytes))
	writeBytes, err := internal.MarshalEcho(echoRequest)
	if err != nil {
		msg := fmt.Sprintf("ping error:%v", err)
		return time.Duration(0), errors.New(msg)
	}

	// 接続
	conn, err := net.DialTimeout("ip4:1", destIP, timeout)
	if err != nil {
		msg := fmt.Sprintf("ping error:%v", err)
		return time.Duration(0), errors.New(msg)
	}

	// Write, Read のタイムアウトを設定する
	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		msg := fmt.Sprintf("ping error:%v", err)
		return time.Duration(0), errors.New(msg)
	}

	// 送信
	writeStart := time.Now()
	{
		wroteLen, err := conn.Write(writeBytes)
		if err != nil {
			msg := fmt.Sprintf("ping error:%v", err)
			return time.Duration(0), errors.New(msg)
		}
		if wroteLen != len(writeBytes) {
			return time.Duration(0), errors.New("ping error:1 回の Write で全てを送信できませんでした。")
		}
	}

	// 受信
	readBytes, err := readn(conn, timeout, len(writeBytes))
	if err != nil {
		msg := fmt.Sprintf("ping error:%v", err)
		return time.Duration(0), errors.New(msg)
	}
	readEnd := time.Now()

	// 受信データを構造体にする
	echoReply := &internal.ICMPEchoMessage{}
	if err = internal.UnmarshalEcho(readBytes, echoReply); err != nil {
		msg := fmt.Sprintf("ping error:%v", err)
		return time.Duration(0), errors.New(msg)
	}

	if !internal.IsSameEchoField(echoRequest, echoReply) {
		return time.Duration(0), errors.New("ping error:EchoReply のフィールドが一致しません。")
	}

	return readEnd.Sub(writeStart), nil
}

func main() {
	for i := 0; i < 5; i++ {
		if duration, err := ping("localhost", time.Second*5, uint16(i), uint16(i), 32); err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("TimeDuration:%v\n", duration)
		}
	}
}
