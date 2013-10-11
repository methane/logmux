# Log Multiplexer

```
+--------+
| Worker +--+
+--------+  |
            |
+--------+  |   +-----------------+   +------------------+
| Worker +--+-->| Log Multiplexer +-->| external command |
+--------+  |   +-----------------+   +------------------+
            |
+--------+  |
| Worker +--+
+--------+
```

## 目的

ログを何かのコマンドに流し混む場合に考えないといけない、次のような問題を
解決するためのツールです。

1. PIPE_BUF (Linux なら 4096バイト) 以上のログを書くと混ざるかも
2. 外部プログラムが死んだ時に、Worker が外部コマンドを再起動すると、Worker数分プロセスが立ち上がってしまう.

これらの問題を真面目に解決しようとすると、 Worker は Master にログを転送し、 Master が
外部コマンドの管理をしないといけません。

nginx のように graceful にバイナリを差し替える機能を持ったサーバーの場合、
Master プロセスの切り替えが発生するので、外部プログラムの管理をどうするかという
問題も発生します。

Master-Worker型のプログラムを作成するたびにこの仕組みを実装するのは面倒なので、
外部に Multiplexer デーモンをおいて、 Worker が気軽にログを投げられるようにします。

## 使い方

Log Multiplexer は、 Unix Domain Socket を Listen します。
そこに書き込まれたデータを、行単位で、指定した外部コマンドにパイプします。

```
$ logmux ソケットのパス "実行するコマンド"
```

sample:
```bash
$ logmux /tmp/log.sock "cat -n >> /tmp/log.txt"
```

同梱している unixcat.py を使って動作を確認できます。

```bash
$ ./unixcat.py /tmp/log.sock
foo
bar
baz
```
