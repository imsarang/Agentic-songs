package client

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"shanks/internal/machine/constants"
	"shanks/internal/machine/proc"
)

const cpuSample = 200 * time.Millisecond

func ReadCapacity() (constants.Capacity, error) {
	first, err := readCPUTimes()
	if err != nil {
		return constants.Capacity{}, err
	}

	time.Sleep(cpuSample)

	second, err := readCPUTimes()
	if err != nil {
		return constants.Capacity{}, err
	}

	load, threads := readLoad()

	return constants.Capacity{
		Load:        load,
		CPUPercent:  cpuPercent(first, second, int(cpuSample/time.Millisecond)),
		Pressure:    readPressure(),
		MemoryBytes: ramSwap(),
		Files:       readFileCount(),
		PIDs:        constants.CountMax{Count: uint64(len(proc.PIDs())), Max: proc.Uint(constants.PathPIDMax)},
		Threads:     constants.CountMax{Count: threads, Max: proc.Uint(constants.PathThreadsMax)},
	}, nil
}

type cpuTimes struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

func (c cpuTimes) total() uint64 {
	return c.user + c.nice + c.system + c.idle + c.iowait + c.irq + c.softirq + c.steal
}

func readCPUTimes() (cpuTimes, error) {
	f, err := os.Open(constants.PathStat)
	if err != nil {
		return cpuTimes{}, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		return cpuTimes{}, sc.Err()
	}
	fields := strings.Fields(sc.Text())
	nums := make([]uint64, 8)
	for i := 0; i < len(nums) && i+1 < len(fields); i++ {
		nums[i], _ = strconv.ParseUint(fields[i+1], 10, 64)
	}
	return cpuTimes{nums[0], nums[1], nums[2], nums[3], nums[4], nums[5], nums[6], nums[7]}, nil
}

func cpuPercent(a, b cpuTimes, sampleMS int) constants.CPUPercent {
	d := cpuTimes{
		user: b.user - a.user, nice: b.nice - a.nice, system: b.system - a.system,
		idle: b.idle - a.idle, iowait: b.iowait - a.iowait, irq: b.irq - a.irq,
		softirq: b.softirq - a.softirq, steal: b.steal - a.steal,
	}
	total := d.total()
	pct := func(v uint64) float64 {
		if total == 0 {
			return 0
		}
		return float64(v) * 100 / float64(total)
	}
	return constants.CPUPercent{
		SampleMS: sampleMS,
		User:     pct(d.user),
		Nice:     pct(d.nice),
		System:   pct(d.system),
		Idle:     pct(d.idle),
		IOWait:   pct(d.iowait),
		IRQ:      pct(d.irq),
		SoftIRQ:  pct(d.softirq),
		Steal:    pct(d.steal),
	}
}

func readLoad() (constants.Load, uint64) {
	b, err := os.ReadFile(constants.PathLoadavg)
	if err != nil {
		return constants.Load{}, 0
	}
	fields := strings.Fields(string(b))
	var load constants.Load
	if len(fields) > 0 {
		load.One, _ = strconv.ParseFloat(fields[0], 64)
	}
	if len(fields) > 1 {
		load.Five, _ = strconv.ParseFloat(fields[1], 64)
	}
	if len(fields) > 2 {
		load.Fifteen, _ = strconv.ParseFloat(fields[2], 64)
	}
	var threads uint64
	if len(fields) > 3 {
		_, after, ok := strings.Cut(fields[3], "/")
		if ok {
			threads, _ = strconv.ParseUint(after, 10, 64)
		}
	}
	return load, threads
}

func ramSwap() constants.CapacityMem {
	mem := proc.MeminfoKB()
	out := constants.CapacityMem{
		RAMTotal:     mem["MemTotal"] * 1024,
		RAMAvailable: mem["MemAvailable"] * 1024,
		SwapTotal:    mem["SwapTotal"] * 1024,
		SwapFree:     mem["SwapFree"] * 1024,
	}
	if out.RAMTotal > out.RAMAvailable {
		out.RAMUsed = out.RAMTotal - out.RAMAvailable
	}
	if out.SwapTotal > out.SwapFree {
		out.SwapUsed = out.SwapTotal - out.SwapFree
	}
	return out
}

func readPressure() constants.Pressure {
	return constants.Pressure{
		CPU:    readPressureFile(constants.PathPressureCPU),
		Memory: readPressureFile(constants.PathPressureMemory),
		IO:     readPressureFile(constants.PathPressureIO),
	}
}

func readPressureFile(path string) constants.PressureLevel {
	f, err := os.Open(path)
	if err != nil {
		return constants.PressureLevel{}
	}
	defer f.Close()
	var out constants.PressureLevel
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		vals := map[string]float64{}
		for _, field := range fields[1:] {
			key, raw, ok := strings.Cut(field, "=")
			if !ok || key == "total" {
				continue
			}
			vals[key], _ = strconv.ParseFloat(raw, 64)
		}
		switch fields[0] {
		case "some":
			out.SomeAvg10, out.SomeAvg60, out.SomeAvg300 = vals["avg10"], vals["avg60"], vals["avg300"]
		case "full":
			out.FullAvg10, out.FullAvg60, out.FullAvg300 = vals["avg10"], vals["avg60"], vals["avg300"]
		}
	}
	return out
}

func readFileCount() constants.CountMax {
	b, err := os.ReadFile(constants.PathFileNr)
	if err != nil {
		return constants.CountMax{}
	}
	fields := strings.Fields(string(b))
	var out constants.CountMax
	if len(fields) > 0 {
		out.Count, _ = strconv.ParseUint(fields[0], 10, 64)
	}
	if len(fields) > 2 {
		out.Max, _ = strconv.ParseUint(fields[2], 10, 64)
	}
	return out
}
