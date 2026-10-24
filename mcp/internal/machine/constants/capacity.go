package constants

const (
	// Aggregate CPU time. Two reads, 200ms apart, become cpu_percent.
	PathStat = "/proc/stat"
	// Load averages plus the running/total task count in the n/m field.
	PathLoadavg = "/proc/loadavg"
	// Tasks stalled on CPU. some/full averages over 10s, 60s, and 300s.
	PathPressureCPU = "/proc/pressure/cpu"
	// Tasks stalled waiting for memory.
	PathPressureMemory = "/proc/pressure/memory"
	// Tasks stalled waiting for disk IO.
	PathPressureIO = "/proc/pressure/io"
	// Open files and the kernel ceiling. Fields are allocated, unused, max.
	PathFileNr = "/proc/sys/fs/file-nr"
	// Highest PID the kernel will assign.
	PathPIDMax = "/proc/sys/kernel/pid_max"
	// Highest number of threads the kernel will allow.
	PathThreadsMax = "/proc/sys/kernel/threads-max"
)

type Capacity struct {
	Load        Load        `json:"load"`
	CPUPercent  CPUPercent  `json:"cpu_percent"`
	Pressure    Pressure    `json:"pressure"`
	MemoryBytes CapacityMem `json:"memory_bytes"`
	Files       CountMax    `json:"files"`
	PIDs        CountMax    `json:"pids"`
	Threads     CountMax    `json:"threads"`
}

type Load struct {
	One     float64 `json:"1"`
	Five    float64 `json:"5"`
	Fifteen float64 `json:"15"`
}

type CPUPercent struct {
	SampleMS int     `json:"sample_ms"`
	User     float64 `json:"user"`
	Nice     float64 `json:"nice"`
	System   float64 `json:"system"`
	Idle     float64 `json:"idle"`
	IOWait   float64 `json:"iowait"`
	IRQ      float64 `json:"irq"`
	SoftIRQ  float64 `json:"softirq"`
	Steal    float64 `json:"steal"`
}

type Pressure struct {
	CPU    PressureLevel `json:"cpu"`
	Memory PressureLevel `json:"memory"`
	IO     PressureLevel `json:"io"`
}

type PressureLevel struct {
	SomeAvg10  float64 `json:"some_avg10"`
	SomeAvg60  float64 `json:"some_avg60"`
	SomeAvg300 float64 `json:"some_avg300"`
	FullAvg10  float64 `json:"full_avg10"`
	FullAvg60  float64 `json:"full_avg60"`
	FullAvg300 float64 `json:"full_avg300"`
}

type CapacityMem struct {
	RAMTotal     uint64 `json:"ram_total"`
	RAMAvailable uint64 `json:"ram_available"`
	RAMUsed      uint64 `json:"ram_used"`
	SwapTotal    uint64 `json:"swap_total"`
	SwapFree     uint64 `json:"swap_free"`
	SwapUsed     uint64 `json:"swap_used"`
}

type CountMax struct {
	Count uint64 `json:"count"`
	Max   uint64 `json:"max"`
}
