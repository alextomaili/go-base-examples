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


Парсер пакетов протокола на Go
https://github.com/google/gopacket


Как парсить протокол

 1. Парсим байтстрим в layers.IPv4, layers.TCP

 2. Смотрим https://github.com/p0f/p0f как пакеты парсятся в промежуточную структуру
   /* Parse PCAP input, with plenty of sanity checking. Store interesting details
     in a protocol-agnostic buffer that will be then examined upstream. */
  void parse_packet(void* junk, const struct pcap_pkthdr* hdr, const u8* data) {...}

  как вызыватся:
      if (pcap_dispatch(pt, -1, (pcap_handler)parse_packet, 0) < 0)


  void parse_packet(void* junk, const struct pcap_pkthdr* hdr, const u8* data) {
    struct packet_data pk;
     .....
     .... заполняем из hdr и data
     .....
    flow_dispatch(&pk); << обработка, тут делаются фингеприныт из данных в "struct packet_data"
  }

## ################################################

Либа для работы с дескрипторами сокетов на низком уровне
https://github.com/mikioh/tcpopt - The tcpopt library provides encoding/decoding of TCP-level socket options for Go.
https://github.com/mikioh/tcpinfo - The tcpinfo library provides encoding/decoding of TCP connection state information

cgo шаблон
https://github.com/mikioh/tcpinfo/blob/c87206fb4c9e77563969f5ed106798b287cc9049/defs_linux.go#L16

#include <linux/inet_diag.h>
#include <linux/sockios.h>
#include <linux/tcp.h>

type tcpInfo C.struct_tcp_info
type tcpCCInfo C.union_tcp_cc_info


type tcpInfo struct {
	State           uint8
	Ca_state        uint8
	Retransmits     uint8
	Probes          uint8
	Backoff         uint8
	Options         uint8
	Pad_cgo_0       [1]byte
	Pad_cgo_1       [1]byte
	Rto             uint32


from /include/uapi/linux/tcp.h

struct tcp_info {
    __u8    tcpi_state;
    __u8    tcpi_ca_state;
    __u8    tcpi_retransmits;
    __u8    tcpi_probes;
    __u8    tcpi_backoff;
    __u8    tcpi_options;


struct tcp_info is a kernel→userspace diagnostics structure used to retrieve detailed
runtime information about a TCP connection.
It does NOT control TCP behavior — it reports the internal state of the TCP socket.

It is used by:
getsockopt(fd, SOL_TCP, TCP_INFO, ...)
tools like ss -ti, ip tcp_metrics, netstat -s, tcpdump -v
performance monitoring systems
congestion-control analysis tools
socket debugging
retransmission and latency diagnostics
connection profiling

