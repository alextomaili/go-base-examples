package tcp_packet_reader

// ********************************************
// * Constants extracted from p0f configuration *
// ********************************************

// Maximum number of TCP options we will process (< 256):
const MAX_TCP_OPT = 24 // #define MAX_TCP_OPT 24

// Maximum TTL distance for non-fuzzy signature matching:
const MAX_DIST = 35 // #define MAX_DIST 35

// Special MSS and window value (used by p0f-sendsyn):
const (
	SPECIAL_MSS = 1331 // #define SPECIAL_MSS 1331
	SPECIAL_WIN = 1337 // #define SPECIAL_WIN 1337
)

// PCAP snapshot length:
const SNAPLEN = 65535 // #define SNAPLEN 65535

// Minimum TCP header lengths (from earlier p0f code)
const (
	MIN_TCP4 = 40 // IPv4 minimum (20 bytes IP + 20 bytes TCP)
	MIN_TCP6 = 60 // IPv6 minimum (40 bytes IP + 20 bytes TCP)
)
