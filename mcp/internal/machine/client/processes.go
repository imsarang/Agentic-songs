package client

import (
	"bufio"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"shanks/internal/machine/constants"
	"shanks/internal/machine/proc"
)

func ReadProcesses() constants.Processes {
	first := readProcStats()
	start := time.Now()
	time.Sleep(cpuSample)
	elapsed := time.Since(start).Seconds()
	second := readProcStats()
	page := uint64(os.Getpagesize())
	now := time.Now().Unix()
	btime := bootTimeSec()

	var out constants.Processes
	out.Zombies = []constants.ProcBrief{}
	out.DiskSleep = []constants.ProcBrief{}
	var cpuRows []constants.ProcUse
	var rssRows []constants.ProcUse
	var etimeRows []constants.ProcUse

	for pid, nowStat := range second {
		switch nowStat.state {
		case "R":
			out.States.Running++
		case "S":
			out.States.Sleeping++
		case "D":
			out.States.DiskSleep++
		case "Z":
			out.States.Zombie++
		case "T", "t":
			out.States.Stopped++
		case "I":
			out.States.Idle++
		default:
			out.States.Other++
		}
		if nowStat.kernel {
			out.KernelThreads++
		}
		if nowStat.state == "Z" && len(out.Zombies) < constants.ProcessListCap {
			out.Zombies = append(out.Zombies, brief(pid, nowStat))
		}
		if nowStat.state == "D" && len(out.DiskSleep) < constants.ProcessListCap {
			out.DiskSleep = append(out.DiskSleep, brief(pid, nowStat))
		}
		prev, ok := first[pid]
		ticks := nowStat.ticks
		if ok && ticks >= prev.ticks {
			ticks = nowStat.ticks - prev.ticks
		} else {
			ticks = 0
		}
		cpu := 0.0
		if elapsed > 0 {
			cpu = float64(ticks) / float64(constants.ClockTicksPerSec) / elapsed * 100
		}
		etime := int64(0)
		if btime > 0 && nowStat.startTicks > 0 {
			started := btime + int64(nowStat.startTicks)/constants.ClockTicksPerSec
			if now > started {
				etime = now - started
			}
		}
		row := constants.ProcUse{
			PID:        pid,
			Command:    nowStat.command,
			CPUPercent: cpu,
			RSSBytes:   nowStat.rssPages * page,
			EtimeSec:   etime,
		}
		cpuRows = append(cpuRows, row)
		rssRows = append(rssRows, row)
		etimeRows = append(etimeRows, row)
	}

	sort.Slice(cpuRows, func(i, j int) bool { return cpuRows[i].CPUPercent > cpuRows[j].CPUPercent })
	sort.Slice(rssRows, func(i, j int) bool { return rssRows[i].RSSBytes > rssRows[j].RSSBytes })
	sort.Slice(etimeRows, func(i, j int) bool { return etimeRows[i].EtimeSec > etimeRows[j].EtimeSec })
	out.TopCPU = capUse(cpuRows)
	out.TopRSS = capUse(rssRows)
	out.TopEtime = capUse(etimeRows)
	return out
}

type procStat struct {
	command    string
	state      string
	ppid       int
	ticks      uint64
	startTicks uint64
	rssPages   uint64
	kernel     bool
}

func readProcStats() map[int]procStat {
	out := map[int]procStat{}
	for _, pid := range proc.PIDs() {
		raw, err := os.ReadFile(proc.StatPath(pid))
		if err != nil {
			continue
		}
		stat, ok := parseProcStat(string(raw))
		if !ok {
			continue
		}
		out[pid] = stat
	}
	return out
}

func parseProcStat(line string) (procStat, bool) {
	open := strings.IndexByte(line, '(')
	close := strings.LastIndexByte(line, ')')
	if open < 0 || close < open+1 {
		return procStat{}, false
	}
	rest := strings.Fields(line[close+1:])
	// rest[0] state, [1] ppid, [6] flags, [11] utime, [12] stime, [19] starttime, [21] rss pages.
	if len(rest) < 22 {
		return procStat{}, false
	}
	ppid, _ := strconv.Atoi(rest[1])
	flags, _ := strconv.ParseUint(rest[6], 10, 64)
	utime, _ := strconv.ParseUint(rest[11], 10, 64)
	stime, _ := strconv.ParseUint(rest[12], 10, 64)
	start, _ := strconv.ParseUint(rest[19], 10, 64)
	rss, _ := strconv.ParseUint(rest[21], 10, 64)
	return procStat{
		command:    line[open+1 : close],
		state:      rest[0],
		ppid:       ppid,
		ticks:      utime + stime,
		startTicks: start,
		rssPages:   rss,
		kernel:     flags&constants.FlagKernelThread != 0,
	}, true
}

func bootTimeSec() int64 {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "btime ") {
			continue
		}
		n, err := strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(line, "btime ")), 10, 64)
		if err != nil {
			return 0
		}
		return n
	}
	return 0
}

func brief(pid int, stat procStat) constants.ProcBrief {
	return constants.ProcBrief{PID: pid, PPID: stat.ppid, Command: stat.command}
}

func capUse(rows []constants.ProcUse) []constants.ProcUse {
	if len(rows) > constants.ProcessTopN {
		rows = rows[:constants.ProcessTopN]
	}
	if rows == nil {
		return []constants.ProcUse{}
	}
	return rows
}
