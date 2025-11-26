package main

import (
	"flag"
	"fmt"
	"github.com/alextomaili/go-base-examples/pkg/tcp_packet_reader"
	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"
	"log"
	"net"
)

var (
	fIp = flag.String("ip", "127.0.0.1", "target ip")
	fNp = flag.Int("n", 20, "packet count")
)

func initArgs() (targetIP net.IP, maxPackets int64) {
	log.SetFlags(log.Ldate)
	log.SetFlags(log.Ltime)
	log.SetFlags(log.Lmicroseconds)
	log.SetFlags(log.Lshortfile)

	flag.Parse()

	maxPackets = int64(*fNp)

	targetIP = net.ParseIP(*fIp).To4()
	if targetIP == nil {
		log.Fatalf("invalid argument: bad ipv4")
	}
	fmt.Printf("Use ip: %s, packet count: %d\n", *fIp, *fNp)

	return targetIP, maxPackets
}

func main() {
	ipv4, maxPackets := initArgs()

	// unix.AF_INET - for ipv4
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_TCP)
	if err != nil {
		log.Fatalf("socket error: %v", err)
	}
	defer unix.Close(fd)

	filter, bpfErr := bpfSynPacketOnly()
	if bpfErr != nil {
		log.Fatalf("BPF assemble error: %v", bpfErr)
	}

	sockFilters := toSockFilter(filter)
	fprog := unix.SockFprog{
		Len:    uint16(len(sockFilters)),
		Filter: &sockFilters[0],
	}
	if err := unix.SetsockoptSockFprog(fd, unix.SOL_SOCKET, unix.SO_ATTACH_FILTER, &fprog); err != nil {
		log.Fatalf("attach BPF error: %v", err)
	}
	log.Println("BPF filter attached (TCP SYN only)")

	addr := &unix.SockaddrInet4{
		Port: 0,
	}
	copy(addr.Addr[:], ipv4)

	if err := unix.Bind(fd, addr); err != nil {
		log.Fatalf("bind error: %v", err)
	}
	log.Println("Raw socket created and bound:", fd)

	recvFromLoop(fd, maxPackets)
	//recvMsgLoop(fd, maxPackets)
}

func bpfSynPacketOnly() ([]bpf.RawInstruction, error) {
	return bpf.Assemble([]bpf.Instruction{

		// load into X the length of an IPv4 packet header
		// (Avoids the old TXA + LoadIndirect pattern which is unreliable in Go’s bpf package for cBPF.)
		// first byte of the packet in our case is first byte of IP packet
		// layout is: [IP][TCP]
		bpf.LoadMemShift{Off: 0},

		// load TCP flag byte
		bpf.LoadIndirect{Off: 13, Size: 1},

		/*  -- this is a hack to debug only, in most cases len of IP packet is 20 bytes
		bpf.LoadAbsolute{Off: 33, Size: 1},
		*/

		// Accept only packets where flags == SYN (0x02)
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 0x02, SkipTrue: 1},

		// Reject
		bpf.RetConstant{Val: 0},

		// Accept full packet
		bpf.RetConstant{Val: 0xFFFF},
	})

}

func recvFromLoop(fd int, maxPackets int64) {
	r := tcp_packet_reader.NewReader()

	rf := func(a []byte, an *int, b []byte, n *int) error {
		var (
			err  error
			from unix.Sockaddr
		)

		*n, from, err = unix.Recvfrom(fd, b, 0)
		if err == nil {
			adr := from.(*unix.SockaddrInet4).Addr[:] // because fd, err := unix.Socket(unix.AF_INET ...
			copy(a, adr)
			*an = len(adr)
		}

		return err
	}

	var pc int64
	for {
		if maxPackets > -1 {
			if pc >= maxPackets {
				return
			}
			pc = pc + 1
		}

		r.Reset()
		err := r.Read(rf)
		if err != nil {
			log.Fatalf("recvfrom error: %v", err)
		}

		err = r.Process()
		if err != nil {
			log.Fatalf("process packet error: %v", err)
		}

		log.Printf("Captured %d bytes from: %+v packet: [%s] \n",
			r.BufLen(), r.IpAddr(), r.PacketStr())
	}
}

func recvMsgLoop(fd int, maxPackets int64) {
	r := tcp_packet_reader.NewReaderExt()

	rf := func(a []byte, an *int, b []byte, n *int, oob []byte, oobn *int, flags *int) error {
		var (
			err  error
			from unix.Sockaddr
		)

		*n, *oobn, *flags, from, err = unix.Recvmsg(fd, b, oob, 0)
		if err == nil {
			adr := from.(*unix.SockaddrInet4).Addr[:] // because fd, err := unix.Socket(unix.AF_INET ...
			copy(a, adr)
			*an = len(adr)
		}

		return err
	}

	var pc int64
	for {
		if maxPackets > -1 {
			if pc >= maxPackets {
				return
			}
			pc = pc + 1
		}

		r.Reset()
		err := r.ReadExt(rf)
		if err != nil {
			log.Fatalf("recvfrom error: %v", err)
		}

		log.Printf("Received %d bytes from %+v (flags: %v, oob bytes: %d)\n",
			r.BufLen(), r.IpAddr(), r.Flags(), r.OobBufLen())
	}
}

func toSockFilter(ins []bpf.RawInstruction) []unix.SockFilter {
	out := make([]unix.SockFilter, len(ins))
	for i, in := range ins {
		out[i] = unix.SockFilter{
			Code: in.Op,
			Jt:   in.Jt,
			Jf:   in.Jf,
			K:    in.K,
		}
	}
	return out
}

func htons(i uint16) uint16 { return (i<<8)&0xff00 | i>>8 }
