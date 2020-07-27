# goping

## 概要
* Go と raw socket の勉強のために ping コマンドを作成した
* 苦労した点
  * データのシリアライズ
  * チェックサムの計算
  * raw socket 関数の扱い方
* Linux 環境だと実行に特権レベルが必要になる
  * sudo して実行する
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
  -c int
    	試行回数を指定します (default 5)
  -d string
    	送信先を指定します
  -t int
    	タイムアウト時間を秒で指定します (default 1)
```

## 実行例
```
$ sudo ./goping -d 8.8.8.8
Pinging 8.8.8.8
返答を受信:RTT=6 ms
返答を受信:RTT=5 ms
返答を受信:RTT=5 ms
返答を受信:RTT=5 ms
返答を受信:RTT=5 ms
```