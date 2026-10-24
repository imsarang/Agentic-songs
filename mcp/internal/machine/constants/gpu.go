package constants

const (
	// DRM cards. cardN/device is the GPU. cardN-eDP-1 is a connector, not a GPU.
	PathDRM = "/sys/class/drm"
)

type GPUs struct {
	Devices []GPU  `json:"devices"`
	NVIDIA  NVIDIA `json:"nvidia"`
}

type GPU struct {
	Card        string  `json:"card"`
	Vendor      string  `json:"vendor"`
	Device      string  `json:"device"`
	BusyPercent float64 `json:"busy_percent"`
	BusyKnown   bool    `json:"busy_known"`
	VRAMUsed    uint64  `json:"vram_used_bytes"`
	VRAMTotal   uint64  `json:"vram_total_bytes"`
	TempC       float64 `json:"temp_c"`
	TempKnown   bool    `json:"temp_known"`
	PowerWatts  float64 `json:"power_watts"`
	PowerKnown  bool    `json:"power_known"`
}

type NVIDIA struct {
	Attached bool   `json:"attached"`
	Summary  string `json:"summary"`
}
