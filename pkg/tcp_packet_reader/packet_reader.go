package tcp_packet_reader

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
)

type Reader struct {
	addr []byte
	an   int
	buf  []byte
	n    int

	strBuf *bytes.Buffer
}

func NewReader() *Reader {
	r := &Reader{
		addr:   make([]byte, 16),
		buf:    make([]byte, 65535),
		strBuf: bytes.NewBuffer(make([]byte, 0, 1024)),
	}

	return r
}

func (r *Reader) Read(rf func(a []byte, an *int, b []byte, n *int) error) error {
	return rf(r.addr, &r.an, r.buf, &r.n)
}

func (r *Reader) Reset() {
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

func (r *Reader) IpAddrStr() string {
	return net.IP(r.IpAddr()).String()
}

func (r *Reader) Process() error {
	p := gopacket.NewPacket(r.buf[:r.n], layers.LayerTypeIPv4, gopacket.NoCopy)
	tcpLayer := p.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return errors.New("packet doesn't contain LayerTypeTCP")
	}

	tcp := tcpLayer.(*layers.TCP)
	r.strDumpPacket(tcp)
	return nil
}

func (r *Reader) strDumpPacket(tcp *layers.TCP) {
	b := r.strBuf
	b.Reset()

	fmt.Fprintf(b, "%s:%d", r.IpAddrStr(), tcp.SrcPort)
	fmt.Fprintf(b, " Flags {SYN: %t, ACK: %t, RST: %t, FIN: %t, PSH: %t, URG: %t}",
		tcp.SYN, tcp.ACK, tcp.RST, tcp.FIN, tcp.PSH, tcp.URG)

	fmt.Fprintf(b, " b: %v, flags byte: %v", r.buf[:40], r.buf[33:34])
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
