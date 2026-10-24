package constants

// NamespaceFiles are the /proc/<pid>/ns entries whose distinct inodes we count.
var NamespaceFiles = []string{"cgroup", "ipc", "mnt", "net", "pid", "time", "user", "uts"}

type Namespaces struct {
	Cgroup int `json:"cgroup"`
	IPC    int `json:"ipc"`
	Mnt    int `json:"mnt"`
	Net    int `json:"net"`
	PID    int `json:"pid"`
	Time   int `json:"time"`
	User   int `json:"user"`
	UTS    int `json:"uts"`
}
