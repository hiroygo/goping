# goping
[![test](https://github.com/hiroygo/goping/actions/workflows/test.yml/badge.svg)](https://github.com/hiroygo/goping/actions/workflows/test.yml)

## 使用方法
```
$ sudo ./goping -h
Usage of ./goping:
  -c uint
        実行回数を指定します (default 5)
  -d string
        送信先をホスト名または IPv4 アドレスで指定します
  -l uint
        ペイロードのサイズをバイトで指定します (default 32)
  -w duration
        タイムアウト時間を指定します (default 5s)
```

## 実行例
```
$ sudo ./goping -d www.google.com
PING www.google.com (172.217.31.164) 32(60) bytes of data.
reply recv: seq=0 rtt=5 ms
reply recv: seq=1 rtt=4 ms
reply recv: seq=2 rtt=4 ms
reply recv: seq=3 rtt=32 ms
reply recv: seq=4 rtt=13 ms

$ sudo ./goping -d 8.8.8.8 -l 56
PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.
reply recv: seq=0 rtt=8 ms
reply recv: seq=1 rtt=4 ms
reply recv: seq=2 rtt=4 ms
reply recv: seq=3 rtt=5 ms
reply recv: seq=4 rtt=13 ms
```
