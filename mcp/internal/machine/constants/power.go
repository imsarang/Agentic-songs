package constants

const (
	// One directory per supply. BAT0 is the battery. ADP0 is the mains adapter.
	PathPowerSupply = "/sys/class/power_supply"
)

type Power struct {
	OnAC      bool      `json:"on_ac"`
	Batteries []Battery `json:"batteries"`
}

type Battery struct {
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	CapacityPercent int     `json:"capacity_percent"`
	EnergyFullWh    float64 `json:"energy_full_wh"`
	EnergyDesignWh  float64 `json:"energy_design_wh"`
	FullRatio       float64 `json:"full_ratio"`
	PowerWatts      float64 `json:"power_watts"`
	CycleCount      int     `json:"cycle_count"`
}
