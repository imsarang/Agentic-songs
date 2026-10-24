package constants

const (
	// Cgroup v2 hierarchy. Every directory is a cgroup.
	PathCgroup = "/sys/fs/cgroup"
	// Controllers enabled on the root cgroup, such as cpu and memory.
	PathCgroupControllers = "/sys/fs/cgroup/cgroup.controllers"
	// This process's cgroup path. The line looks like 0::/user.slice/...
	PathCgroupSelf = "/proc/self/cgroup"
	// Current memory charge of one cgroup, in bytes.
	CgroupMemoryFile = "memory.current"
	// How many cgroups to return, largest memory first.
	CgroupTopN = 10
)

type Cgroups struct {
	Version     string      `json:"version"`
	Controllers []string    `json:"controllers"`
	Count       int         `json:"count"`
	Self        string      `json:"self"`
	TopMemory   []CgroupMem `json:"top_memory"`
}

type CgroupMem struct {
	Path  string `json:"path"`
	Bytes uint64 `json:"bytes"`
}
