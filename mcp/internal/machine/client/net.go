package client

import (
	"bufio"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"shanks/internal/machine/constants"
)

func ReadNet() constants.Net {
	counters := readDevCounters()
	ifaces, _ := net.Interfaces()
	out := constants.Net{Interfaces: []constants.Iface{}, ListenPorts: []int{}}
	for _, iface := range ifaces {
		row := constants.Iface{
			Name:         iface.Name,
			MTU:          iface.MTU,
			Up:           iface.Flags&net.FlagUp != 0,
			HardwareAddr: iface.HardwareAddr.String(),
			Addresses:    []string{},
		}
		if c, ok := counters[iface.Name]; ok {
			row.RXBytes, row.TXBytes, row.RXErrors, row.TXErrors = c[0], c[1], c[2], c[3]
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			row.Addresses = append(row.Addresses, addr.String())
		}
		out.Interfaces = append(out.Interfaces, row)
	}
	out.DefaultRoute = readDefaultRoute()
	out.Sockets = readSockstat()
	ports := map[int]struct{}{}
	addTCP(constants.PathTCP, &out.TCP, ports)
	addTCP(constants.PathTCP6, &out.TCP, ports)
	for port := range ports {
		out.ListenPorts = append(out.ListenPorts, port)
	}
	sort.Ints(out.ListenPorts)
	if len(out.ListenPorts) > constants.ListenPortCap {
		out.ListenPorts = out.ListenPorts[:constants.ListenPortCap]
	}
	out.RetransSegs = snmpField(constants.PathSNMP, "Tcp:", "RetransSegs")
	out.ListenDrops = constants.ListenDrop{
		Overflows: snmpField(constants.PathNetstat, "TcpExt:", "ListenOverflows"),
		Drops:     snmpField(constants.PathNetstat, "TcpExt:", "ListenDrops"),
	}
	return out
}

func readDevCounters() map[string][4]uint64 {
	out := map[string][4]uint64{}
	f, err := os.Open(constants.PathNetDev)
	if err != nil {
		return out
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		name, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) < 11 {
			continue
		}
		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		rxErr, _ := strconv.ParseUint(fields[2], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)
		txErr, _ := strconv.ParseUint(fields[10], 10, 64)
		out[strings.TrimSpace(name)] = [4]uint64{rx, tx, rxErr, txErr}
	}
	return out
}

func readDefaultRoute() constants.Route {
	f, err := os.Open(constants.PathRoute)
	if err != nil {
		return constants.Route{}
	}
	defer f.Close()
	best := constants.Route{Metric: -1}
	sc := bufio.NewScanner(f)
	sc.Scan()
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 7 || fields[1] != "00000000" {
			continue
		}
		metric, _ := strconv.Atoi(fields[6])
		if best.Metric >= 0 && metric >= best.Metric {
			continue
		}
		best = constants.Route{Interface: fields[0], Gateway: hexLEIP(fields[2]), Metric: metric}
	}
	if best.Metric < 0 {
		return constants.Route{}
	}
	return best
}

func hexLEIP(hex string) string {
	n, err := strconv.ParseUint(hex, 16, 32)
	if err != nil || len(hex) != 8 {
		return hex
	}
	return strconv.FormatUint(n&0xff, 10) + "." +
		strconv.FormatUint((n>>8)&0xff, 10) + "." +
		strconv.FormatUint((n>>16)&0xff, 10) + "." +
		strconv.FormatUint((n>>24)&0xff, 10)
}

func readSockstat() constants.Sockets {
	f, err := os.Open(constants.PathSockstat)
	if err != nil {
		return constants.Sockets{}
	}
	defer f.Close()
	var out constants.Sockets
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		switch {
		case len(fields) >= 3 && fields[0] == "sockets:":
			out.Used, _ = strconv.ParseUint(fields[2], 10, 64)
		case len(fields) >= 3 && fields[0] == "TCP:":
			out.TCPInuse = fieldAfter(fields, "inuse")
			out.TCPTimeWait = fieldAfter(fields, "tw")
		case len(fields) >= 3 && fields[0] == "UDP:":
			out.UDPInuse = fieldAfter(fields, "inuse")
		}
	}
	return out
}

func fieldAfter(fields []string, key string) uint64 {
	for i := 0; i+1 < len(fields); i++ {
		if fields[i] == key {
			n, _ := strconv.ParseUint(fields[i+1], 10, 64)
			return n
		}
	}
	return 0
}

func addTCP(path string, states *constants.TCPStates, ports map[int]struct{}) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Scan()
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 4 {
			continue
		}
		switch fields[3] {
		case "01":
			states.Established++
		case "02":
			states.SynSent++
		case "03":
			states.SynRecv++
		case "04":
			states.FinWait1++
		case "05":
			states.FinWait2++
		case "06":
			states.TimeWait++
		case "07":
			states.Close++
		case "08":
			states.CloseWait++
		case "09":
			states.LastAck++
		case "0A":
			states.Listen++
			if port := hexPort(fields[1]); port > 0 {
				ports[port] = struct{}{}
			}
		case "0B":
			states.Closing++
		default:
			states.Other++
		}
	}
}

func hexPort(addr string) int {
	_, port, ok := strings.Cut(addr, ":")
	if !ok {
		return 0
	}
	n, err := strconv.ParseInt(port, 16, 32)
	if err != nil {
		return 0
	}
	return int(n)
}

func snmpField(path, section, key string) uint64 {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var header []string
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != section {
			continue
		}
		if header == nil {
			header = fields
			continue
		}
		for i := 1; i < len(header) && i < len(fields); i++ {
			if header[i] == key {
				n, _ := strconv.ParseUint(fields[i], 10, 64)
				return n
			}
		}
		return 0
	}
	return 0
}
