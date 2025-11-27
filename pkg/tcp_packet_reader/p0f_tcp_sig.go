package tcp_packet_reader

import (
	"fmt"
	"strconv"
	"time"
)

const (
	WIN_TYPE_NORMAL = 0
)

// Placeholder for compatibility with original p0f code
type tcpSigRecord struct{}

type tcpSig struct {
	OptHash   uint32 // Hash of opt_layout & opt_cnt
	Quirks    uint32
	OptEOLPad uint8         // Amount of padding past EOL
	IPOptLen  uint8         // Length of IP options
	IPVer     int8          // -1 = any, 4 = IPv4, 6 = IPv6
	TTL       uint8         // Actual TTL
	MSS       int32         // -1 = any, otherwise MSS
	Win       uint16        // Window size
	WinType   uint8         // WIN_TYPE_*
	WScale    int16         // -1 = any
	PayClass  int8          // -1 = any, 0 = zero, 1 = non-zero
	TotHdr    uint16        // Total header length
	TS1       uint32        // Own timestamp
	RecvMs    uint64        // unix time in ms
	Matched   *tcpSigRecord // placeholder for record pointer
	Fuzzy     uint8
	Dist      uint8
}

func hash32(data []byte, seed uint32) uint32 {
	var h = seed
	for _, b := range data {
		h += uint32(b)
		h += h << 10
		h ^= h >> 6
	}
	h += h << 3
	h ^= h >> 11
	h += h << 15
	return h
}

func unixTimeMs() uint64 {
	return uint64(time.Now().UnixNano() / 1e6)
}

func packetToSig(pk *packetData, seed uint32, ts *tcpSig) {
	// --- Hash TCP option layout  ---
	optData := pk.OptLayout[:pk.OptCnt]
	ts.OptHash = hash32(optData, seed)

	// --- Copy simple fields ---
	ts.Quirks = pk.Quirks
	ts.OptEOLPad = pk.OptEOLPad
	ts.IPOptLen = pk.IPOptLen
	ts.IPVer = int8(pk.IPVer)
	ts.TTL = pk.TTL
	ts.MSS = int32(pk.MSS)
	ts.Win = pk.Win

	// Same as in p0f: normal window type
	ts.WinType = WIN_TYPE_NORMAL

	ts.WScale = int16(pk.WScale)

	// 0 = zero payload, 1 = non-zero payload
	if pk.PayLen > 0 {
		ts.PayClass = 1
	} else {
		ts.PayClass = 0
	}

	ts.TotHdr = pk.TotHdr
	ts.TS1 = pk.TS1
	ts.RecvMs = unixTimeMs()

	ts.Matched = nil
	ts.Fuzzy = 0
	ts.Dist = 0
}

// dumpSig renders a TCP signature into a caller-provided buffer.
// ret must have enough capacity; otherwise returns an error.
// Returns number of bytes written through returned length.
func dumpSig(pk *packetData, ts *tcpSig, ret []byte, synMSS uint16) (int, error) {
	if len(ret) == 0 {
		return 0, fmt.Errorf("ret buffer too small")
	}

	i := 0
	write := func(s string) error {
		if i+len(s) > len(ret) {
			return fmt.Errorf("buffer too small")
		}
		copy(ret[i:], s)
		i += len(s)
		return nil
	}

	writeInt := func(v int) error {
		// max 10 digits inplace
		buf := strconv.AppendInt(ret[:0], int64(v), 10)
		if i+len(buf) > len(ret) {
			return fmt.Errorf("buffer too small")
		}
		copy(ret[i:], buf)
		i += len(buf)
		return nil
	}

	// ----------------------------------------------------------
	// Part 1: IP version, TTL+dist, IP option len
	//   C: "%u:%u+%u:%u:"
	// ----------------------------------------------------------
	dist := guessDist(pk.TTL)

	if dist > MAX_DIST {
		// "%u:%u+?:%u:"
		if err := write(fmt.Sprintf("%d:%d+?:%d:", pk.IPVer, pk.TTL, pk.IPOptLen)); err != nil {
			return 0, err
		}
	} else {
		// "%u:%u+%u:%u:"
		if err := write(fmt.Sprintf("%d:%d+%d:%d:", pk.IPVer, pk.TTL, dist, pk.IPOptLen)); err != nil {
			return 0, err
		}
	}

	// ----------------------------------------------------------
	// Part 2: MSS
	// ----------------------------------------------------------
	if pk.MSS == SPECIAL_MSS && pk.TCPType == (TCP_SYN|TCP_ACK) {
		if err := write("*:"); err != nil {
			return 0, err
		}
	} else {
		if err := writeInt(int(pk.MSS)); err != nil {
			return 0, err
		}
		if err := write(":"); err != nil {
			return 0, err
		}
	}

	// ----------------------------------------------------------
	// Part 3: detect window multiplier (MSS or MTU)
	// ----------------------------------------------------------
	winM, useMTU := detectWinMulti(ts, synMSS)

	if winM > 0 {
		if useMTU {
			if err := write("mtu"); err != nil {
				return 0, err
			}
		} else {
			if err := write("mss"); err != nil {
				return 0, err
			}
		}
		if err := write("*"); err != nil {
			return 0, err
		}
		if err := writeInt(int(winM)); err != nil {
			return 0, err
		}
	} else {
		if err := writeInt(int(pk.Win)); err != nil {
			return 0, err
		}
	}

	if err := write(","); err != nil {
		return 0, err
	}
	if err := writeInt(int(pk.WScale)); err != nil {
		return 0, err
	}
	if err := write(":"); err != nil {
		return 0, err
	}

	// ----------------------------------------------------------
	// Part 4: TCP option layout
	// ----------------------------------------------------------
	for idx := uint8(0); idx < pk.OptCnt; idx++ {
		opt := pk.OptLayout[idx]

		if idx > 0 {
			if err := write(","); err != nil {
				return 0, err
			}
		}

		switch opt {
		case TCPOPT_EOL:
			if err := write(fmt.Sprintf("eol+%d", pk.OptEOLPad)); err != nil {
				return 0, err
			}
		case TCPOPT_NOP:
			if err := write("nop"); err != nil {
				return 0, err
			}
		case TCPOPT_MAXSEG:
			if err := write("mss"); err != nil {
				return 0, err
			}
		case TCPOPT_WSCALE:
			if err := write("ws"); err != nil {
				return 0, err
			}
		case TCPOPT_SACKOK:
			if err := write("sok"); err != nil {
				return 0, err
			}
		case TCPOPT_SACK:
			if err := write("sack"); err != nil {
				return 0, err
			}
		case TCPOPT_TSTAMP:
			if err := write("ts"); err != nil {
				return 0, err
			}
		default:
			if err := write(fmt.Sprintf("?%d", opt)); err != nil {
				return 0, err
			}
		}
	}

	if err := write(":"); err != nil {
		return 0, err
	}

	// ----------------------------------------------------------
	// Part 5: quirks
	// ----------------------------------------------------------
	if pk.Quirks != 0 {
		first := true

		add := func(s string) error {
			if !first {
				if err := write(","); err != nil {
					return err
				}
			}
			first = false
			return write(s)
		}

		if pk.Quirks&QUIRK_DF != 0 {
			if err := add("df"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_NZ_ID != 0 {
			if err := add("id+"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_ZERO_ID != 0 {
			if err := add("id-"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_ECN != 0 {
			if err := add("ecn"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_NZ_MBZ != 0 {
			if err := add("0+"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_FLOW != 0 {
			if err := add("flow"); err != nil {
				return 0, err
			}
		}

		if pk.Quirks&QUIRK_ZERO_SEQ != 0 {
			if err := add("seq-"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_NZ_ACK != 0 {
			if err := add("ack+"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_ZERO_ACK != 0 {
			if err := add("ack-"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_NZ_URG != 0 {
			if err := add("uptr+"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_URG != 0 {
			if err := add("urgf+"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_PUSH != 0 {
			if err := add("pushf+"); err != nil {
				return 0, err
			}
		}

		if pk.Quirks&QUIRK_OPT_ZERO_TS1 != 0 {
			if err := add("ts1-"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_OPT_NZ_TS2 != 0 {
			if err := add("ts2+"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_OPT_EOL_NZ != 0 {
			if err := add("opt+"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_OPT_EXWS != 0 {
			if err := add("exws"); err != nil {
				return 0, err
			}
		}
		if pk.Quirks&QUIRK_OPT_BAD != 0 {
			if err := add("bad"); err != nil {
				return 0, err
			}
		}
	}

	// ----------------------------------------------------------
	// Part 6: payload size
	// ----------------------------------------------------------
	if pk.PayLen > 0 {
		if err := write(":+"); err != nil {
			return 0, err
		}
	} else {
		if err := write(":0"); err != nil {
			return 0, err
		}
	}

	return i, nil
}

func guessDist(ttl uint8) uint8 {
	if ttl <= 32 {
		return 32 - ttl
	}
	if ttl <= 64 {
		return 64 - ttl
	}
	if ttl <= 128 {
		return 128 - ttl
	}
	return 255 - ttl
}

func detectWinMulti(ts *tcpSig, synMSS uint16) (multiplier int16, useMTU bool) {
	win := ts.Win
	mss := ts.MSS
	mss12 := mss - 12

	if win == 0 || mss < 100 || ts.WinType != WIN_TYPE_NORMAL {
		return -1, false
	}

	try := func(div int32, mtu bool) (ok bool, m int16, use bool) {
		if div > 0 && win%uint16(div) == 0 {
			return true, int16(win / uint16(div)), mtu
		}
		return false, 0, false
	}

	// MSS
	if ok, m, mtu := try(mss, false); ok {
		return m, mtu
	}

	// MSS-12 (timestamp influence)
	if ts.TS1 != 0 {
		if ok, m, mtu := try(mss12, false); ok {
			return m, mtu
		}
	}

	// MTU=1500 common MSS patterns
	if ok, m, mtu := try(1500-MIN_TCP4, false); ok {
		return m, mtu
	}
	if ok, m, mtu := try(1500-MIN_TCP4-12, false); ok {
		return m, mtu
	}

	// IPv6 MTU patterns
	if ts.IPVer == IP_VER6 {
		if ok, m, mtu := try(1500-MIN_TCP6, false); ok {
			return m, mtu
		}
		if ok, m, mtu := try(1500-MIN_TCP6-12, false); ok {
			return m, mtu
		}
	}

	// MTU-based cases
	if ok, m, mtu := try(mss+MIN_TCP4, true); ok {
		return m, mtu
	}
	if ok, m, mtu := try(int32(mss)+int32(ts.TotHdr), true); ok {
		return m, mtu
	}
	if ts.IPVer == IP_VER6 {
		if ok, m, mtu := try(mss+MIN_TCP6, true); ok {
			return m, mtu
		}
	}
	if ok, m, mtu := try(1500, true); ok {
		return m, mtu
	}

	// SYN+ACK peer-MSS clues
	if synMSS != 0 {
		if ok, m, mtu := try(int32(synMSS), false); ok {
			return m, mtu
		}
		if ok, m, mtu := try(int32(synMSS)-12, false); ok {
			return m, mtu
		}
	}

	return -1, false
}
