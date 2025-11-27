package tcp_packet_reader

// IP version constants (match p0f definitions)
const (
	IP_VER4 int8 = 0x04
	IP_VER6 int8 = 0x06
)

// TCP flag bitmasks (uint8), matching p0f C definitions
const (
	TCP_FIN  uint8 = 0x01
	TCP_SYN  uint8 = 0x02
	TCP_RST  uint8 = 0x04
	TCP_PUSH uint8 = 0x08
	TCP_ACK  uint8 = 0x10
	TCP_URG  uint8 = 0x20
)

// TCP option kinds (match p0f definitions)
const (
	TCPOPT_EOL    uint8 = 0 // End of options (1 byte)
	TCPOPT_NOP    uint8 = 1 // No-op (1 byte)
	TCPOPT_MAXSEG uint8 = 2 // Maximum segment size (4 bytes)
	TCPOPT_WSCALE uint8 = 3 // Window scaling (3 bytes)
	TCPOPT_SACKOK uint8 = 4 // SACK permitted (2 bytes)
	TCPOPT_SACK   uint8 = 5 // SACK option (10–34 bytes)
	TCPOPT_TSTAMP uint8 = 8 // TCP timestamp (10 bytes)
)
