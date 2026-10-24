package constants

const (
	// Kernel ceiling for open files on the whole machine.
	PathFileMax = "/proc/sys/fs/file-max"
	// Kernel ceiling for the accept backlog of one listening socket.
	PathSomaxconn = "/proc/sys/net/core/somaxconn"
	// Soft and hard ulimits of this process.
	PathSelfLimits = "/proc/self/limits"
)

type Limits struct {
	PIDMax        uint64 `json:"pid_max"`
	ThreadsMax    uint64 `json:"threads_max"`
	FileMax       uint64 `json:"file_max"`
	Somaxconn     uint64 `json:"somaxconn"`
	OpenFilesSoft string `json:"open_files_soft"`
	OpenFilesHard string `json:"open_files_hard"`
	ProcessesSoft string `json:"processes_soft"`
	ProcessesHard string `json:"processes_hard"`
}
