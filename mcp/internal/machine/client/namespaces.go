package client

import (
	"os"
	"path/filepath"
	"strconv"

	"shanks/internal/machine/constants"
	"shanks/internal/machine/proc"
)

func ReadNamespaces() constants.Namespaces {
	sets := map[string]map[string]struct{}{}
	for _, name := range constants.NamespaceFiles {
		sets[name] = map[string]struct{}{}
	}
	for _, pid := range proc.PIDs() {
		dir := filepath.Join(proc.PathProc, strconv.Itoa(pid), "ns")
		for _, name := range constants.NamespaceFiles {
			target, err := os.Readlink(filepath.Join(dir, name))
			if err != nil || target == "" {
				continue
			}
			sets[name][target] = struct{}{}
		}
	}
	return constants.Namespaces{
		Cgroup: len(sets["cgroup"]),
		IPC:    len(sets["ipc"]),
		Mnt:    len(sets["mnt"]),
		Net:    len(sets["net"]),
		PID:    len(sets["pid"]),
		Time:   len(sets["time"]),
		User:   len(sets["user"]),
		UTS:    len(sets["uts"]),
	}
}
