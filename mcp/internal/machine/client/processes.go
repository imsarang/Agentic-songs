package client

import (
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

	var out constants.Processes
	out.Zombies = []constants.ProcBrief{}
	out.DiskSleep = []constants.ProcBrief{}
	var cpuRows []constants.ProcUse
	var rssRows []constants.ProcUse

	for pid, now := range second {
		switch now.state {
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
		if now.kernel {
			out.KernelThreads++
		}
		if now.state == "Z" && len(out.Zombies) < constants.ProcessListCap {
			out.Zombies = append(out.Zombies, brief(pid, now))
		}
		if now.state == "D" && len(out.DiskSleep) < constants.ProcessListCap {
			out.DiskSleep = append(out.DiskSleep, brief(pid, now))
		}
		prev, ok := first[pid]
		ticks := now.ticks
		if ok && ticks >= prev.ticks {
			ticks = now.ticks - prev.ticks
		} else {
			ticks = 0
		}
		cpu := 0.0
		if elapsed > 0 {
			cpu = float64(ticks) / float64(constants.ClockTicksPerSec) / elapsed * 100
		}
		row := constants.ProcUse{
			PID:        pid,
			Command:    now.command,
			CPUPercent: cpu,
			RSSBytes:   now.rssPages * page,
		}
		cpuRows = append(cpuRows, row)
		rssRows = append(rssRows, row)
	}

	sort.Slice(cpuRows, func(i, j int) bool { return cpuRows[i].CPUPercent > cpuRows[j].CPUPercent })
	sort.Slice(rssRows, func(i, j int) bool { return rssRows[i].RSSBytes > rssRows[j].RSSBytes })
	out.TopCPU = capUse(cpuRows)
	out.TopRSS = capUse(rssRows)
	return out
}

type procStat struct {
	command  string
	state    string
	ppid     int
	ticks    uint64
	rssPages uint64
	kernel   bool
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
	// rest[0] state, [1] ppid, [6] flags, [11] utime, [12] stime, [21] rss pages.
	if len(rest) < 22 {
		return procStat{}, false
	}
	ppid, _ := strconv.Atoi(rest[1])
	flags, _ := strconv.ParseUint(rest[6], 10, 64)
	utime, _ := strconv.ParseUint(rest[11], 10, 64)
	stime, _ := strconv.ParseUint(rest[12], 10, 64)
	rss, _ := strconv.ParseUint(rest[21], 10, 64)
	return procStat{
		command:  line[open+1 : close],
		state:    rest[0],
		ppid:     ppid,
		ticks:    utime + stime,
		rssPages: rss,
		kernel:   flags&constants.FlagKernelThread != 0,
	}, true
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
