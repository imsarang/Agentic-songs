package client

import (
	"os"
	"path/filepath"
	"strconv"

	"shanks/internal/machine/constants"
	"shanks/internal/machine/proc"
)

func ReadPower() constants.Power {
	out := constants.Power{Batteries: []constants.Battery{}}
	entries, err := os.ReadDir(constants.PathPowerSupply)
	if err != nil {
		return out
	}
	for _, entry := range entries {
		dir := filepath.Join(constants.PathPowerSupply, entry.Name())
		kind := proc.Text(filepath.Join(dir, "type"))
		switch kind {
		case "Mains":
			if proc.Text(filepath.Join(dir, "online")) == "1" {
				out.OnAC = true
			}
		case "Battery":
			bat := readBattery(entry.Name(), dir)
			if bat.Status == "Charging" || bat.Status == "Full" {
				out.OnAC = true
			}
			out.Batteries = append(out.Batteries, bat)
		}
	}
	return out
}

func readBattery(name, dir string) constants.Battery {
	full := energy(dir, "energy_full", "charge_full")
	design := energy(dir, "energy_full_design", "charge_full_design")
	percent, _ := strconv.Atoi(proc.Text(filepath.Join(dir, "capacity")))
	cycles, _ := strconv.Atoi(proc.Text(filepath.Join(dir, "cycle_count")))
	bat := constants.Battery{
		Name:            name,
		Status:          proc.Text(filepath.Join(dir, "status")),
		CapacityPercent: percent,
		EnergyFullWh:    float64(full) / 1e6,
		EnergyDesignWh:  float64(design) / 1e6,
		PowerWatts:      float64(proc.Uint(filepath.Join(dir, "power_now"))) / 1e6,
		CycleCount:      cycles,
	}
	if design > 0 {
		bat.FullRatio = float64(full) / float64(design)
	}
	return bat
}

func energy(dir, primary, fallback string) uint64 {
	if n := proc.Uint(filepath.Join(dir, primary)); n > 0 {
		return n
	}
	return proc.Uint(filepath.Join(dir, fallback))
}
