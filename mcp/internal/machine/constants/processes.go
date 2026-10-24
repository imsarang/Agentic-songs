package constants

const (
	// Per-process line under /proc/<pid>: state, parent, kernel-thread flag, CPU ticks, RSS.
	ProcStatFile = "stat"
	// Bit in the flags field of ProcStatFile. Set for kernel threads such as kthreadd.
	FlagKernelThread = 0x00200000
	// Linux user clock. CPU ticks in ProcStatFile advance this many times per second.
	ClockTicksPerSec = 100
	// Rows returned for the hottest CPU and largest RSS lists.
	ProcessTopN = 10
	// Rows returned for the zombie and disk-sleep lists. Counts elsewhere are the full totals.
	ProcessListCap = 20
)

type Processes struct {
	States        StateCounts `json:"states"`
	Zombies       []ProcBrief `json:"zombies"`
	DiskSleep     []ProcBrief `json:"disk_sleep"`
	KernelThreads int         `json:"kernel_threads"`
	TopCPU        []ProcUse   `json:"top_cpu"`
	TopRSS        []ProcUse   `json:"top_rss"`
}

type StateCounts struct {
	Running   int `json:"running"`
	Sleeping  int `json:"sleeping"`
	DiskSleep int `json:"disk_sleep"`
	Zombie    int `json:"zombie"`
	Stopped   int `json:"stopped"`
	Idle      int `json:"idle"`
	Other     int `json:"other"`
}

type ProcBrief struct {
	PID     int    `json:"pid"`
	PPID    int    `json:"ppid"`
	Command string `json:"command"`
}

type ProcUse struct {
	PID        int     `json:"pid"`
	Command    string  `json:"command"`
	CPUPercent float64 `json:"cpu_percent"`
	RSSBytes   uint64  `json:"rss_bytes"`
}
