package tcp_packet_reader

import (
	"encoding/binary"
	"github.com/google/gopacket/layers"
)

// https://github.com/p0f/p0f/blob/4abbd20ecd7421461e360c77a98dff98e08a8b10/process.h#L22
//
//	struct packet_data {
//		u8  ip_ver;                           /* IP_VER4, IP_VER6                   */
//		u8  tcp_type;                         /* TCP_SYN, ACK, FIN, RST             */
//		u8  src[16];                          /* Source address (left-aligned)      */
//		u8  dst[16];                          /* Destination address (left-aligned  */
//		u16 sport;                            /* Source port                        */
//		u16 dport;                            /* Destination port                   */
//		u8  ttl;                              /* Observed TTL                       */
//		u8  tos;                              /* IP ToS value                       */
//		u16 mss;                              /* Maximum segment size               */
//		u16 win;                              /* Window size                        */
//		u8  wscale;                           /* Window scaling                     */
//		u16 tot_hdr;                          /* Total headers (for MTU calc)       */
//		u8  opt_layout[MAX_TCP_OPT];          /* Ordering of TCP options            */
//		u8  opt_cnt;                          /* Count of TCP options               */
//		u8  opt_eol_pad;                      /* Amount of padding past EOL         */
//		u32 ts1;                              /* Own timestamp                      */
//		u32 quirks;                           /* QUIRK_*                            */
//		u8  ip_opt_len;                       /* Length of IP options               */
//		u8* payload;                          /* TCP payload                        */
//		u16 pay_len;                          /* Length of TCP payload              */
//		u32 seq;                              /* seq value seen                     */
//	};

type packetData struct {
	IPVer     uint8    // ip_ver
	TCPType   uint8    // tcp_type (SYN/ACK/FIN/RST flags masked)
	Src       [16]byte // left-aligned copy of IPv4 or IPv6
	Dst       [16]byte
	Sport     uint16
	Dport     uint16
	TTL       uint8
	TOS       uint8
	MSS       uint16
	Win       uint16
	WScale    uint8
	TotHdr    uint16
	OptLayout [40]uint8 // MAX_TCP_OPT = 40 in p0f
	OptCnt    uint8
	OptEOLPad uint8
	TS1       uint32
	Quirks    uint32
	IPOptLen  uint8
	Payload   *byte
	PayLen    uint16
	Seq       uint32
}

// IP-level quirks:
const (
	QUIRK_ECN     uint32 = 0x00000001 // ECN supported
	QUIRK_DF      uint32 = 0x00000002 // DF used (probably PMTUD)
	QUIRK_NZ_ID   uint32 = 0x00000004 // Non-zero IDs when DF set
	QUIRK_ZERO_ID uint32 = 0x00000008 // Zero IDs when DF not set
	QUIRK_NZ_MBZ  uint32 = 0x00000010 // IP "must be zero" field isn't
	QUIRK_FLOW    uint32 = 0x00000020 // IPv6 flows used
)

// Core TCP quirks:
const (
	QUIRK_ZERO_SEQ uint32 = 0x00001000 // SEQ is zero
	QUIRK_NZ_ACK   uint32 = 0x00002000 // ACK non-zero when ACK flag not set
	QUIRK_ZERO_ACK uint32 = 0x00004000 // ACK is zero when ACK flag set
	QUIRK_NZ_URG   uint32 = 0x00008000 // URG non-zero when URG flag not set
	QUIRK_URG      uint32 = 0x00010000 // URG flag set
	QUIRK_PUSH     uint32 = 0x00020000 // PUSH flag on a control packet
)

// TCP option quirks:
const (
	QUIRK_OPT_ZERO_TS1 uint32 = 0x01000000 // Own timestamp set to zero
	QUIRK_OPT_NZ_TS2   uint32 = 0x02000000 // Peer timestamp non-zero on SYN
	QUIRK_OPT_EOL_NZ   uint32 = 0x04000000 // Non-zero padding past EOL
	QUIRK_OPT_EXWS     uint32 = 0x08000000 // Excessive window scaling
	QUIRK_OPT_BAD      uint32 = 0x10000000 // Problem parsing TCP options
)

func buildPacketDataIpv4AndTcpPacket(ipv4 *layers.IPv4, tcp *layers.TCP, pk *packetData) *packetData {
	//
	// --------------------------
	//  IP Parsing (IPv4 only)
	// --------------------------
	//

	pk.IPVer = 4

	// Copy IPv4 into left-aligned 16-byte array
	copy(pk.Src[:], ipv4.SrcIP.To4())
	copy(pk.Dst[:], ipv4.DstIP.To4())

	pk.TOS = ipv4.TOS >> 2 // p0f drops ECN bits
	pk.TTL = ipv4.TTL
	pk.IPOptLen = uint8(ipv4.IHL*4 - 20) // IP options length

	// tot_hdr = IPv4 header length
	pk.TotHdr = uint16(ipv4.IHL * 4)

	//
	// quirks: IP-level
	//

	if ipv4.Flags&layers.IPv4DontFragment != 0 {
		pk.Quirks |= QUIRK_DF

		ipid := uint16(ipv4.Id)
		if ipid != 0 {
			pk.Quirks |= QUIRK_NZ_ID
		}
	} else {
		ipid := uint16(ipv4.Id)
		if ipid == 0 {
			pk.Quirks |= QUIRK_ZERO_ID
		}
	}

	// ECN
	if ipv4.TOS&(0x03) != 0 {
		pk.Quirks |= QUIRK_ECN
	}

	//
	// --------------------------
	// TCP Parsing
	// --------------------------
	//

	pk.Sport = uint16(tcp.SrcPort)
	pk.Dport = uint16(tcp.DstPort)
	pk.Seq = tcp.Seq

	// Window
	pk.Win = tcp.Window

	// P0f tcp_type = flags & (SYN|ACK|FIN|RST)
	pk.TCPType = 0
	if tcp.SYN {
		pk.TCPType |= 0x02
	}
	if tcp.ACK {
		pk.TCPType |= 0x10
	}
	if tcp.FIN {
		pk.TCPType |= 0x01
	}
	if tcp.RST {
		pk.TCPType |= 0x04
	}

	//
	// Quirks: TCP flags
	//

	if pk.Seq == 0 {
		pk.Quirks |= QUIRK_ZERO_SEQ
	}

	if tcp.ACK {
		if tcp.Ack == 0 {
			pk.Quirks |= QUIRK_ZERO_ACK
		}
	} else {
		// Non-zero ACK on non-ACK (except RST)
		if tcp.Ack != 0 && !tcp.RST {
			pk.Quirks |= QUIRK_NZ_ACK
		}
	}

	if tcp.URG {
		pk.Quirks |= QUIRK_URG
	} else if tcp.Urgent != 0 {
		pk.Quirks |= QUIRK_NZ_URG
	}

	if tcp.PSH {
		pk.Quirks |= QUIRK_PUSH
	}

	//
	// TCP header length → contributes to tot_hdr
	//
	tcpHdrLen := int(tcp.DataOffset) * 4
	pk.TotHdr += uint16(tcpHdrLen)

	//
	// Payload
	//
	pk.Payload = nil
	pk.PayLen = 0

	//
	// --------------------------
	// TCP Options
	// --------------------------
	//

	// MSS, WScale, TS1 – possible from gopacket
	for _, opt := range tcp.Options {

		switch opt.OptionType {

		case layers.TCPOptionKindMSS:
			if len(opt.OptionData) >= 2 {
				pk.MSS = binary.BigEndian.Uint16(opt.OptionData[:2])
			}

		case layers.TCPOptionKindWindowScale:
			if len(opt.OptionData) >= 1 {
				pk.WScale = opt.OptionData[0]
				if pk.WScale > 14 {
					pk.Quirks |= QUIRK_OPT_EXWS
				}
			}

		case layers.TCPOptionKindTimestamps:
			if len(opt.OptionData) >= 8 {
				pk.TS1 = binary.BigEndian.Uint32(opt.OptionData[:4])
				if pk.TS1 == 0 {
					pk.Quirks |= QUIRK_OPT_ZERO_TS1
				}
			}
		}
	}

	//
	// Missing p0f raw-option-level details
	//
	// pk.OptLayout = doesn't have information (requires raw TCP header bytes)
	//    p0f requires raw byte-by-byte option stream; gopacket only provides structured decoded options.
	//
	// pk.OptCnt = doesn't have information
	//   p0f counts options based on raw header walk. Cannot emulate without raw bytes.
	//
	// pk.OptEOLPad = doesn't have information
	//   requires raw bytes after EOL marker; gopacket discards padding.
	//
	// pk.Quirks |= QUIRK_OPT_* (those based on raw layout) = cannot determine
	//  QUIRK_OPT_EOL_NZ, QUIRK_OPT_BAD, QUIRK_OPT_NZ_TS2
	//  depend on raw option parsing errors or padding.
	//
	// Payload pointer exact semantics: gopacket gives payload, but not pointer arithmetic;
	//  depending on p0f’s usage this is “good enough.”
	//

	return pk
}
