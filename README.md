# goping

## 使用方法
```
$ sudo ./goping -h
Usage of ./goping:
  -c uint
        実行回数を指定します (default 5)
  -d string
        IPv4送信先を指定します
  -l uint
        ペイロードのサイズをバイトで指定します (default 32)
  -w duration
        タイムアウト時間を指定します (default 1s)
```

## 実行例
```
$ sudo ./goping -d 8.8.8.8
PING 8.8.8.8
recv reply: rtt=5 ms
recv reply: rtt=4 ms
recv reply: rtt=4 ms
recv reply: rtt=4 ms
recv reply: rtt=5 ms
```
