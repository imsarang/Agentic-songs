package client

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"shanks/internal/machine/constants"
	"shanks/internal/machine/proc"
)

func ReadBlockers() constants.Blockers {
	out := constants.Blockers{
		DiskSleep:   []constants.ProcBrief{},
		KernelLines: []string{},
	}
	for _, pid := range proc.PIDs() {
		brief, state, ok := readBrief(pid)
		if !ok || state != "D" || len(out.DiskSleep) >= constants.ProcessListCap {
			continue
		}
		out.DiskSleep = append(out.DiskSleep, brief)
	}
	out.Locks = readLocks()
	out.KernelLines = kernelLines()
	return out
}

func readBrief(pid int) (constants.ProcBrief, string, bool) {
	raw := proc.Text(proc.StatPath(pid))
	open := strings.IndexByte(raw, '(')
	close := strings.LastIndexByte(raw, ')')
	if open < 0 || close < open+1 {
		return constants.ProcBrief{}, "", false
	}
	rest := strings.Fields(raw[close+1:])
	if len(rest) < 2 {
		return constants.ProcBrief{}, "", false
	}
	ppid := 0
	for i := 0; i < len(rest[1]); i++ {
		if rest[1][i] < '0' || rest[1][i] > '9' {
			return constants.ProcBrief{}, "", false
		}
		ppid = ppid*10 + int(rest[1][i]-'0')
	}
	return constants.ProcBrief{PID: pid, PPID: ppid, Command: raw[open+1 : close]}, rest[0], true
}

func readLocks() constants.LockSummary {
	f, err := os.Open(constants.PathLocks)
	if err != nil {
		return constants.LockSummary{}
	}
	defer f.Close()
	var out constants.LockSummary
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		out.Total++
		kind := fields[1]
		if fields[1] == "->" && len(fields) > 2 {
			out.Waiters++
			kind = fields[2]
		}
		switch kind {
		case "POSIX":
			out.POSIX++
		case "FLOCK":
			out.Flock++
		case "LEASE":
			out.Lease++
		default:
			out.Other++
		}
	}
	return out
}

func kernelLines() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, constants.Journalctl, "--no-pager", "-o", "cat", "-k", "-n", "80", "-g", "lockdep|Out of memory|oom-kill")
	raw, err := cmd.Output()
	if err != nil && len(raw) == 0 {
		return []string{}
	}
	var lines []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) >= constants.KernelLineCap {
			break
		}
	}
	if lines == nil {
		return []string{}
	}
	return lines
}
