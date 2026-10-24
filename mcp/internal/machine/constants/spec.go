package constants

type Spec struct {
	Hostname      string       `json:"hostname"`
	Vendor        string       `json:"vendor"`
	Product       string       `json:"product"`
	Version       string       `json:"version"`
	Chassis       string       `json:"chassis"`
	Virtual       string       `json:"virtual"`
	OS            string       `json:"os"`
	Kernel        string       `json:"kernel"`
	Arch          string       `json:"arch"`
	CPU           CPU          `json:"cpu"`
	MemoryBytes   MemoryBytes  `json:"memory_bytes"`
	UptimeSeconds float64      `json:"uptime_seconds"`
	Filesystems   []Filesystem `json:"filesystems"`
}

type CPU struct {
	Model   string `json:"model"`
	Sockets int    `json:"sockets"`
	Cores   int    `json:"cores"`
	Threads int    `json:"threads"`
}

type MemoryBytes struct {
	RAM  uint64 `json:"ram"`
	Swap uint64 `json:"swap"`
}

type Filesystem struct {
	Mount      string `json:"mount"`
	Source     string `json:"source"`
	FSType     string `json:"fstype"`
	TotalBytes uint64 `json:"total_bytes"`
	UsedBytes  uint64 `json:"used_bytes"`
	FreeBytes  uint64 `json:"free_bytes"`
}

var RealFS = map[string]struct{}{
	"ext4": {}, "xfs": {}, "btrfs": {}, "vfat": {}, "ntfs": {}, "ntfs3": {},
	"exfat": {}, "f2fs": {}, "zfs": {},
}

var Hypervisors = []string{"kvm", "qemu", "vmware", "virtualbox", "microsoft", "xen", "bochs", "parallels"}
