package tcp_packet_reader

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"time"
)

type PrintFormat int

const (
	PF_DEBUG PrintFormat = iota
	PF_CSV
)

type (
	Reader struct {
		addr []byte
		an   int
		buf  []byte
		n    int

		// p0f compatible packet representation
		pd     packetData
		tcpSig tcpSig

		tcpSynSig    []byte
		tcpSynSigN   int
		tcpSynSplits []int

		// for debug
		strBuf *bytes.Buffer
	}

	ProcessOpts struct {
		PrintFormat PrintFormat
	}
)

func NewReader() *Reader {
	r := &Reader{
		addr:         make([]byte, 16),
		buf:          make([]byte, 65535),
		tcpSynSig:    make([]byte, 1024),
		tcpSynSplits: make([]int, 0, 32),
		strBuf:       bytes.NewBuffer(make([]byte, 0, 2048)),
	}

	return r
}

func (r *Reader) Read(rf func(a []byte, an *int, b []byte, n *int) error) error {
	return rf(r.addr, &r.an, r.buf, &r.n)
}

func (r *Reader) resetPacketData() {
	r.pd = packetData{}
	r.tcpSig = tcpSig{}
	r.tcpSynSigN = 0
	r.tcpSynSplits = r.tcpSynSplits[:0]
}

func (r *Reader) Reset() {
	r.resetPacketData()
	r.an = 0
	r.n = 0
	r.strBuf.Reset()
}

func (r *Reader) BufLen() int {
	return r.n
}

func (r *Reader) IpAddr() []byte {
	return r.addr[:r.an]
}

func (r *Reader) TcpSynSig() []byte {
	return r.tcpSynSig[:r.tcpSynSigN]
}

func (r *Reader) TcpSynSigPartsLen() int {
	return len(r.tcpSynSplits) / 2
}

func (r *Reader) TcpSynSigPart(n int) []byte {
	return r.tcpSynSig[r.tcpSynSplits[n*2]:r.tcpSynSplits[n*2+1]]
}

func (r *Reader) IpAddrStr() string {
	return net.IP(r.IpAddr()).String()
}

func (r *Reader) Process(po *ProcessOpts) error {
	p := gopacket.NewPacket(r.buf[:r.n], layers.LayerTypeIPv4, gopacket.NoCopy)

	ip4Layer := p.Layer(layers.LayerTypeIPv4)
	if ip4Layer == nil {
		return errors.New("packet doesn't contain LayerTypeIPv4")
	}
	ipv4 := ip4Layer.(*layers.IPv4)

	tcpLayer := p.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return errors.New("packet doesn't contain LayerTypeTCP")
	}
	tcp := tcpLayer.(*layers.TCP)

	if err := r.buildSignature(ipv4, tcp); err != nil {
		return err
	}

	r.strDumpIpv4AndTcpPacket(ipv4, tcp, po.PrintFormat)
	return nil
}

func (r *Reader) CsvHeader() string {
	return fmt.Sprintf("date_time\tip\tport\tsig\tip_version\tttl_dist\tip_opt_len\tmss\tmss_mtu\ttcp_opt\tquirks\tpayload_size")
}

func (r *Reader) strDumpIpv4AndTcpPacket(ipv4 *layers.IPv4, tcp *layers.TCP, pf PrintFormat) {
	b := r.strBuf
	b.Reset()

	now := time.Now()

	fmt.Fprintf(b, "%04d-%02d-%02dT%02d:%02d:%02d.%03d",
		now.Year(),
		int(now.Month()),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond()/1e6,
	)

	if pf == PF_CSV {
		fmt.Fprintf(b, "\t%s\t%d", ipv4.SrcIP.String(), tcp.SrcPort)

		fmt.Fprintf(b, "\t%s\t", string(r.TcpSynSig()))
		r.strDumpSivAsTabDel(b)

		return
	}

	fmt.Fprintf(b, " addr: [%s:%d] ", ipv4.SrcIP.String(), tcp.SrcPort)
	fmt.Fprintf(b, "ipv4: [IHL: %d, Length: %d, TTL: %d] ", ipv4.IHL, ipv4.Length, ipv4.TTL)
	fmt.Fprintf(b, "tcp: [ Flags {SYN: %t, ACK: %t, RST: %t, FIN: %t, PSH: %t, URG: %t}] ",
		tcp.SYN, tcp.ACK, tcp.RST, tcp.FIN, tcp.PSH, tcp.URG)
	fmt.Fprintf(b, "raw captured: [ tcp flags: %v, rawb: %v ] ", r.buf[33:34], r.buf[:40])

	fmt.Fprintf(b, "syn_sig: [ %s # %d <", string(r.TcpSynSig()), len(r.tcpSynSplits))
	r.strDumpSivAsTabDel(b)
	fmt.Fprintf(b, ">]")
}

func (r *Reader) strDumpSivAsTabDel(b *bytes.Buffer) {
	for i := 0; i < r.TcpSynSigPartsLen(); i++ {
		if i != 0 {
			fmt.Fprintf(b, "\t")
		}
		fmt.Fprintf(b, "%s", string(r.TcpSynSigPart(i)))
	}
}

func (r *Reader) buildSignature(ipv4 *layers.IPv4, tcp *layers.TCP) error {
	r.resetPacketData()

	buildPacketDataIpv4AndTcpPacket(ipv4, tcp, &r.pd)
	packetToSig(&r.pd, 0, &r.tcpSig)

	// synMSS = 0 - looks like OK for syn packet because we have this in fp_tcp.c:
	// struct tcp_sig* fingerprint_tcp( ........
	//    ....
	//    add_observation_field("raw_sig", dump_sig(pk, sig, f->syn_mss));
	//    if (pk->tcp_type == TCP_SYN) f->syn_mss = pk->mss;
	//
	if sl, err := dumpSig(&r.pd, &r.tcpSig, r.tcpSynSig, &r.tcpSynSplits, 0); err == nil {
		r.tcpSynSigN = sl
	} else {
		return err
	}

	return nil
}

func (r *Reader) PacketStr() string {
	return r.strBuf.String()
}

type ReaderExt struct {
	Reader
	oob   []byte
	oobn  int
	flags int
}

func NewReaderExt() *ReaderExt {
	r := &ReaderExt{
		Reader: Reader{
			addr: make([]byte, 16),
			buf:  make([]byte, 65535),
		},
		oob: make([]byte, 1024),
	}
	return r
}

func (r *ReaderExt) ReadExt(rf func(a []byte, an *int, b []byte, n *int,
	oob []byte, oobn *int, flags *int) error) error {
	return rf(r.addr, &r.an, r.buf, &r.n, r.oob, &r.oobn, &r.flags)
}

func (r *ReaderExt) Reset() {
	r.Reader.Reset()
	r.oobn = 0
	r.flags = 0
}

func (r *ReaderExt) OobBufLen() int {
	return r.oobn
}

func (r *ReaderExt) Flags() int {
	return r.flags
}
