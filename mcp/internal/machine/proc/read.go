package proc

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"shanks/internal/machine/constants"
)

const (
	// Memory and swap, in kB. Spec reads the totals. Capacity reads what is still available.
	PathMeminfo = "/proc/meminfo"
	// One directory per live PID. Processes reads <pid>/stat from here. Capacity counts the directories.
	PathProc = "/proc"
)

// MeminfoKB returns every numeric field in PathMeminfo, in kB.
func MeminfoKB() map[string]uint64 {
	out := map[string]uint64{}
	f, err := os.Open(PathMeminfo)
	if err != nil {
		return out
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
		out[strings.TrimSuffix(fields[0], ":")] = n
	}
	return out
}

// PIDs lists the numeric directory names in PathProc.
func PIDs() []int {
	entries, err := os.ReadDir(PathProc)
	if err != nil {
		return nil
	}
	out := make([]int, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if name == "" || name[0] < '0' || name[0] > '9' {
			continue
		}
		pid, err := strconv.Atoi(name)
		if err != nil {
			continue
		}
		out = append(out, pid)
	}
	return out
}

// StatPath is /proc/<pid>/stat.
func StatPath(pid int) string {
	return PathProc + "/" + strconv.Itoa(pid) + "/" + constants.ProcStatFile
}

// Text reads a file and trims whitespace. Missing files return an empty string.
func Text(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// Uint reads a file whose only content is an integer, such as pid_max.
func Uint(path string) uint64 {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	n, _ := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
	return n
}
