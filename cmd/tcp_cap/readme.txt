p0f демон
  # P0f is a tool that utilizes an array of sophisticated, purely passive traffic fingerprinting mechanisms to identify
  # the players behind any incidental TCP/IP communications (often as little as a single normal SYN)
  # without interfering in any way
https://lcamtuf.coredump.cx/p0f3/

.-[ 1.2.3.4/1524 -> 4.3.2.1/80 (syn) ]-
|
| client   = 1.2.3.4
| os       = Windows XP
| dist     = 8
| params   = none
| raw_sig  = 4:120+8:0:1452:65535,0:mss,nop,nop,sok:df,id+:0
|
`----

.-[ 1.2.3.4/1524 -> 4.3.2.1/80 (mtu) ]-
|
| client   = 1.2.3.4
| link     = DSL
| raw_mtu  = 1492
|
`----

.-[ 1.2.3.4/1524 -> 4.3.2.1/80 (uptime) ]-
|
| client   = 1.2.3.4
| uptime   = 0 days 11 hrs 16 min (modulo 198 days)
| raw_freq = 250.00 Hz
|
|
`----

.-[ 1.2.3.4/1524 -> 4.3.2.1/80 (http request) ]-
|
| client   = 1.2.3.4/1524
| app      = Firefox 5.x or newer
| lang     = English
| params   = none
| raw_sig  = 1:Host,User-Agent,Accept=[text/html,application/xhtml+xml...
|
`----


файл правил p0f.fp из дистра дебиана
https://sources.debian.org/src/p0f/3.09b-2/p0f.fp


Библиотека для работы с p0f демоном (p0f passive fingerprinting)
https://github.com/gurre/gop0f


Теория фингерпринтов
https://www.cs.princeton.edu/techreports/2019/010.pdf


Исходники алгоритмов of p0f version 2
https://tools.netsa.cert.org/p0f/libp0f.html
 # libp0f is a library implementation of p0f version 2 retrieved from http://lcamtuf.coredump.cx/p0f3/.
тут можно скачать архив


Неофициальная репа в git для p0f
https://github.com/p0f/p0f
 # https://github.com/p0f/p0f/blob/master/p0f.fp - фал правил



Парсер пакетов протокола
https://github.com/google/gopacket

