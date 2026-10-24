package client

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"shanks/internal/machine/constants"
	"shanks/internal/machine/proc"
)

func ReadCgroups() constants.Cgroups {
	out := constants.Cgroups{
		Version:     "unknown",
		Controllers: []string{},
		TopMemory:   []constants.CgroupMem{},
	}
	if text := proc.Text(constants.PathCgroupControllers); text != "" {
		out.Version = "v2"
		out.Controllers = strings.Fields(text)
	}
	self := proc.Text(constants.PathCgroupSelf)
	if _, path, ok := strings.Cut(self, "::"); ok {
		out.Self = path
	}
	var rows []constants.CgroupMem
	_ = filepath.WalkDir(constants.PathCgroup, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || !d.IsDir() {
			return nil
		}
		out.Count++
		raw := proc.Text(filepath.Join(path, constants.CgroupMemoryFile))
		if raw == "" {
			return nil
		}
		rows = append(rows, constants.CgroupMem{Path: path, Bytes: proc.Uint(filepath.Join(path, constants.CgroupMemoryFile))})
		return nil
	})
	sort.Slice(rows, func(i, j int) bool { return rows[i].Bytes > rows[j].Bytes })
	if len(rows) > constants.CgroupTopN {
		rows = rows[:constants.CgroupTopN]
	}
	if rows != nil {
		out.TopMemory = rows
	}
	return out
}
