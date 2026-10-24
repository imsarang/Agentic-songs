package constants

const (
	// Per-interface byte and error counters.
	PathNetDev = "/proc/net/dev"
	// Kernel routing table. Destination 00000000 is the default route.
	PathRoute = "/proc/net/route"
	// Socket totals: used, TCP inuse, UDP inuse.
	PathSockstat = "/proc/net/sockstat"
	// TCP counters. RetransSegs is on the Tcp line.
	PathSNMP = "/proc/net/snmp"
	// Extended TCP counters. ListenOverflows and ListenDrops are on the TcpExt line.
	PathNetstat = "/proc/net/netstat"
	// IPv4 TCP sockets. Column st is the connection state.
	PathTCP = "/proc/net/tcp"
	// IPv6 TCP sockets.
	PathTCP6 = "/proc/net/tcp6"
	// How many listening ports to return.
	ListenPortCap = 40
)

type Net struct {
	Interfaces   []Iface    `json:"interfaces"`
	DefaultRoute Route      `json:"default_route"`
	Sockets      Sockets    `json:"sockets"`
	TCP          TCPStates  `json:"tcp"`
	ListenPorts  []int      `json:"listen_ports"`
	RetransSegs  uint64     `json:"retrans_segs"`
	ListenDrops  ListenDrop `json:"listen_drops"`
}

type Iface struct {
	Name         string   `json:"name"`
	MTU          int      `json:"mtu"`
	Up           bool     `json:"up"`
	HardwareAddr string   `json:"hardware_addr"`
	Addresses    []string `json:"addresses"`
	RXBytes      uint64   `json:"rx_bytes"`
	TXBytes      uint64   `json:"tx_bytes"`
	RXErrors     uint64   `json:"rx_errors"`
	TXErrors     uint64   `json:"tx_errors"`
}

type Route struct {
	Interface string `json:"interface"`
	Gateway   string `json:"gateway"`
	Metric    int    `json:"metric"`
}

type Sockets struct {
	Used        uint64 `json:"used"`
	TCPInuse    uint64 `json:"tcp_inuse"`
	TCPTimeWait uint64 `json:"tcp_time_wait"`
	UDPInuse    uint64 `json:"udp_inuse"`
}

type TCPStates struct {
	Established int `json:"established"`
	SynSent     int `json:"syn_sent"`
	SynRecv     int `json:"syn_recv"`
	FinWait1    int `json:"fin_wait1"`
	FinWait2    int `json:"fin_wait2"`
	TimeWait    int `json:"time_wait"`
	Close       int `json:"close"`
	CloseWait   int `json:"close_wait"`
	LastAck     int `json:"last_ack"`
	Listen      int `json:"listen"`
	Closing     int `json:"closing"`
	Other       int `json:"other"`
}

type ListenDrop struct {
	Overflows uint64 `json:"overflows"`
	Drops     uint64 `json:"drops"`
}
