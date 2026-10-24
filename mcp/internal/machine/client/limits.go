package client

import (
	"bufio"
	"os"
	"strings"

	"shanks/internal/machine/constants"
	"shanks/internal/machine/proc"
)

func ReadLimits() constants.Limits {
	out := constants.Limits{
		PIDMax:     proc.Uint(constants.PathPIDMax),
		ThreadsMax: proc.Uint(constants.PathThreadsMax),
		FileMax:    proc.Uint(constants.PathFileMax),
		Somaxconn:  proc.Uint(constants.PathSomaxconn),
	}
	f, err := os.Open(constants.PathSelfLimits)
	if err != nil {
		return out
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 3 {
			continue
		}
		switch {
		case strings.HasPrefix(sc.Text(), "Max open files"):
			out.OpenFilesSoft, out.OpenFilesHard = fields[len(fields)-3], fields[len(fields)-2]
		case strings.HasPrefix(sc.Text(), "Max processes"):
			out.ProcessesSoft, out.ProcessesHard = fields[len(fields)-3], fields[len(fields)-2]
		}
	}
	return out
}
