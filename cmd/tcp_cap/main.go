package main

import (
	"encoding/hex"
	"flag"
	"fmt"
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

	addr := &unix.SockaddrInet4{
		Port: 0,
	}
	copy(addr.Addr[:], ipv4)

	if err := unix.Bind(fd, addr); err != nil {
		log.Fatalf("bind error: %v", err)
	}
	log.Println("Raw socket created and bound:", fd)

	// BPF filter: match TCP SYN packets (tcp[13] & 0x02 != 0)
	// IP header is included because AF_INET raw sockets pass IP+TCP
	filter, err := bpf.Assemble([]bpf.Instruction{
		// Load TCP flags byte: offset 13 from TCP header beginning.
		// But TCP header offset depends on IP header length.
		// So we must calculate offset: IP header length = (ip[0] & 0x0F) * 4
		// TCP flags = TCP offset + 13

		// Step 1: Load first byte of IP header (version + IHL)
		bpf.LoadAbsolute{Off: 0, Size: 1},

		// Step 2: Extract IHL (lower 4 bits) and multiply by 4
		bpf.ALUOpConstant{Op: bpf.ALUOpAnd, Val: 0x0F},
		bpf.ALUOpConstant{Op: bpf.ALUOpShiftLeft, Val: 2}, // x4

		// Now A contains IP header length.
		// We'll save it in X register.
		bpf.TXA{}, // transfer A -> X

		// Step 3: Load TCP flags: load byte at offset X + 13
		bpf.LoadIndirect{Off: 13, Size: 1},

		// Step 4: Check SYN bit
		bpf.ALUOpConstant{Op: bpf.ALUOpAnd, Val: 0x02},
		bpf.JumpIf{Cond: bpf.JumpNotEqual, Val: 0, SkipFalse: 1},

		// ACCEPT
		bpf.RetConstant{Val: 65535},

		// REJECT
		bpf.RetConstant{Val: 0},
	})
	if err != nil {
		log.Fatalf("BPF assemble error: %v", err)
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

	//recvFromLoop(fd, maxPackets)
	recvMsgLoop(fd, maxPackets)
}

func recvFromLoop(fd int, maxPackets int64) {
	buf := make([]byte, 65535)

	var pc int64
	for {
		if maxPackets > -1 {
			if pc >= maxPackets {
				return
			}
			pc = pc + 1
		}

		n, from, err := unix.Recvfrom(fd, buf, 0)
		if err != nil {
			log.Fatalf("recvfrom error: %v", err)
		}

		log.Printf("Captured %d bytes from %+v\n", n, from)
	}
}

func recvMsgLoop(fd int, maxPackets int64) {
	buf := make([]byte, 65535)
	oob := make([]byte, 512) // for ancillary data

	var pc int64
	for {
		if maxPackets > -1 {
			if pc >= maxPackets {
				return
			}
			pc = pc + 1
		}

		n, oobn, flags, from, err := unix.Recvmsg(fd, buf, oob, 0)
		if err != nil {
			log.Fatalf("recvmsg error: %v", err)
		}

		log.Printf("Received %d bytes from %+v (flags: %v, oob bytes: %d) [%s ...] \n",
			n, from, flags, oobn, hex.EncodeToString(buf[:min(n, 16)]))
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
