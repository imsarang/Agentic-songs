package client

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"syscall"

	"shanks/internal/machine/constants"
)

func ReadSpec() (constants.Spec, error) {
	var u syscall.Utsname
	if err := syscall.Uname(&u); err != nil {
		return constants.Spec{}, err
	}
	host, err := os.Hostname()
	if err != nil {
		return constants.Spec{}, err
	}

	cpu := readCPU()
	ram, swap := readMem()
	uptime := readUptime()
	vendor := readDMI("sys_vendor")
	product := readDMI("product_name")
	version := readDMI("product_version")

	return constants.Spec{
		Hostname:      host,
		Vendor:        vendor,
		Product:       product,
		Version:       version,
		Chassis:       chassisName(readDMI("chassis_type")),
		Virtual:       virtualKind(cpu.hypervisor, product, vendor),
		OS:            cString(u.Sysname[:]),
		Kernel:        cString(u.Release[:]),
		Arch:          cString(u.Machine[:]),
		CPU:           cpu.info,
		MemoryBytes:   constants.MemoryBytes{RAM: ram, Swap: swap},
		UptimeSeconds: uptime,
		Filesystems:   readFilesystems(),
	}, nil
}

type cpuRead struct {
	info       constants.CPU
	hypervisor bool
}

func readCPU() cpuRead {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return cpuRead{}
	}
	defer f.Close()

	var out cpuRead
	sockets := map[string]struct{}{}
	coresLine := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		key, val, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "processor":
			out.info.Threads++
		case "model name":
			if out.info.Model == "" {
				out.info.Model = val
			}
		case "physical id":
			sockets[val] = struct{}{}
		case "cpu cores":
			if coresLine == 0 {
				coresLine, _ = strconv.Atoi(val)
			}
		case "flags":
			if !out.hypervisor && strings.Contains(" "+val+" ", " hypervisor ") {
				out.hypervisor = true
			}
		}
	}
	out.info.Sockets = len(sockets)
	if out.info.Sockets == 0 {
		out.info.Sockets = 1
	}
	if coresLine > 0 {
		out.info.Cores = coresLine * out.info.Sockets
	} else {
		out.info.Cores = out.info.Threads
	}
	return out
}

func readMem() (ram, swap uint64) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		n, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			ram = n * 1024
		case "SwapTotal:":
			swap = n * 1024
		}
	}
	return ram, swap
}

func readUptime() float64 {
	b, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(b))
	if len(fields) == 0 {
		return 0
	}
	v, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0
	}
	return v
}

func readFilesystems() []constants.Filesystem {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil
	}
	defer f.Close()
	seen := map[string]struct{}{}
	var out []constants.Filesystem
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 3 {
			continue
		}
		source, mount, fstype := fields[0], fields[1], fields[2]
		if _, ok := constants.RealFS[fstype]; !ok {
			continue
		}
		if _, ok := seen[mount]; ok {
			continue
		}
		seen[mount] = struct{}{}
		var st syscall.Statfs_t
		if err := syscall.Statfs(mount, &st); err != nil || st.Bsize <= 0 {
			continue
		}
		bsize := uint64(st.Bsize)
		total := st.Blocks * bsize
		free := st.Bavail * bsize
		used := uint64(0)
		if total > free {
			used = total - free
		}
		out = append(out, constants.Filesystem{
			Mount:      unescapeMount(mount),
			Source:     unescapeMount(source),
			FSType:     fstype,
			TotalBytes: total,
			UsedBytes:  used,
			FreeBytes:  free,
		})
	}
	return out
}

func virtualKind(hypervisor bool, product, vendor string) string {
	blob := strings.ToLower(product + " " + vendor)
	for _, name := range constants.Hypervisors {
		if strings.Contains(blob, name) {
			return name
		}
	}
	if hypervisor {
		return "virtual"
	}
	return "physical"
}

func chassisName(code string) string {
	switch strings.TrimSpace(code) {
	case "3":
		return "desktop"
	case "8":
		return "portable"
	case "9":
		return "laptop"
	case "10":
		return "notebook"
	case "30":
		return "tablet"
	case "31", "32":
		return "convertible"
	default:
		if code == "" {
			return ""
		}
		return "other"
	}
}

func readDMI(name string) string {
	b, err := os.ReadFile("/sys/class/dmi/id/" + name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func unescapeMount(s string) string {
	r := strings.NewReplacer(`\040`, " ", `\011`, "\t", `\012`, "\n", `\134`, `\`)
	return r.Replace(s)
}

func cString(b []int8) string {
	out := make([]byte, 0, len(b))
	for _, c := range b {
		if c == 0 {
			break
		}
		out = append(out, byte(c))
	}
	return string(out)
}
