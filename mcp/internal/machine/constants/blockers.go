package constants

const (
	// File locks held on this machine. A line starting with "->" is a waiter.
	PathLocks = "/proc/locks"
	// How many kernel log lines to keep.
	KernelLineCap = 20
)

type Blockers struct {
	DiskSleep   []ProcBrief `json:"disk_sleep"`
	Locks       LockSummary `json:"locks"`
	KernelLines []string    `json:"kernel_lines"`
}

type LockSummary struct {
	Total   int `json:"total"`
	Waiters int `json:"waiters"`
	POSIX   int `json:"posix"`
	Flock   int `json:"flock"`
	Lease   int `json:"lease"`
	Other   int `json:"other"`
}
