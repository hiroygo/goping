# goping

## 概要
* Go と raw socket の勉強のために ping コマンドを作成した
* 勉強になったところ
  * データのシリアライズ
  * チェックサムの計算
  * raw socket 関数の扱い方
* Linux 環境だと実行に特権レベルが必要になる
  * sudo して実行する
  * setuid して実行する
  * net.ipv4.ping_group_range を設定する
* 参考サイト
  * https://tools.ietf.org/html/rfc792
  * https://ja.wikipedia.org/wiki/Ping
  * https://ja.wikipedia.org/wiki/Internet_Control_Message_Protocol
  * https://www.mew.org/~kazu/doc/bsdmag/cksum.html

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