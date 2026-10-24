package client

import (
	"fmt"

	"shanks/internal/machine/constants"
)

func ReadHealth() constants.Health {
	spec, specErr := ReadSpec()
	cap, capErr := ReadCapacity()
	checks := []constants.Check{
		diskCheck(spec, specErr),
		memoryCheck(cap, capErr),
		cpuCheck(spec, specErr, cap, capErr),
		ioCheck(cap, capErr),
		netCheck(),
		batteryCheck(),
	}
	overall := "ok"
	for _, check := range checks {
		overall = worse(overall, check.Severity)
	}
	return constants.Health{Overall: overall, Checks: checks}
}

func diskCheck(spec constants.Spec, err error) constants.Check {
	if err != nil {
		return constants.Check{Name: "disk", Severity: "warn", Text: err.Error()}
	}
	check := constants.Check{Name: "disk", Severity: "ok", Text: "no filesystem larger than 1 GB"}
	worstFree := 101.0
	found := false
	for _, fs := range spec.Filesystems {
		if fs.TotalBytes < constants.HealthDiskMinBytes || fs.TotalBytes == 0 {
			continue
		}
		free := float64(fs.FreeBytes) / float64(fs.TotalBytes) * 100
		severity := "ok"
		if free < constants.HealthDiskCrit {
			severity = "crit"
		} else if free < constants.HealthDiskWarn {
			severity = "warn"
		}
		if !found || rank(severity) > rank(check.Severity) || (severity == check.Severity && free < worstFree) {
			found = true
			worstFree = free
			check = constants.Check{
				Name: "disk", Severity: severity,
				Text: fmt.Sprintf("%s free %.1f%% (%s of %s)", fs.Mount, free, gb(fs.FreeBytes), gb(fs.TotalBytes)),
			}
		}
	}
	return check
}

func memoryCheck(cap constants.Capacity, err error) constants.Check {
	if err != nil {
		return constants.Check{Name: "memory", Severity: "warn", Text: err.Error()}
	}
	severity := psiSeverity(cap.Pressure.Memory.SomeAvg10)
	mem := cap.MemoryBytes
	if mem.SwapTotal > 0 {
		used := float64(mem.SwapUsed) / float64(mem.SwapTotal)
		if used >= constants.HealthSwapCrit {
			severity = worse(severity, "crit")
		} else if used >= constants.HealthSwapWarn {
			severity = worse(severity, "warn")
		}
	}
	text := fmt.Sprintf("available %s of %s, swap used %s of %s, PSI memory some avg10 %.2f",
		gb(mem.RAMAvailable), gb(mem.RAMTotal), gb(mem.SwapUsed), gb(mem.SwapTotal), cap.Pressure.Memory.SomeAvg10)
	return constants.Check{Name: "memory", Severity: severity, Text: text}
}

func cpuCheck(spec constants.Spec, specErr error, cap constants.Capacity, capErr error) constants.Check {
	if specErr != nil {
		return constants.Check{Name: "cpu", Severity: "warn", Text: specErr.Error()}
	}
	if capErr != nil {
		return constants.Check{Name: "cpu", Severity: "warn", Text: capErr.Error()}
	}
	cores := spec.CPU.Cores
	if cores < 1 {
		cores = 1
	}
	severity := "ok"
	if cap.Pressure.CPU.SomeAvg10 >= constants.HealthCPUCrit {
		severity = "crit"
	} else if cap.Pressure.CPU.SomeAvg10 >= constants.HealthCPUWarn {
		severity = "warn"
	}
	perCore := cap.Load.One / float64(cores)
	if perCore >= 2 {
		severity = worse(severity, "crit")
	} else if perCore >= 1 {
		severity = worse(severity, "warn")
	}
	text := fmt.Sprintf("load1 %.2f on %d cores, iowait %.1f%%, PSI cpu some avg10 %.2f",
		cap.Load.One, cores, cap.CPUPercent.IOWait, cap.Pressure.CPU.SomeAvg10)
	return constants.Check{Name: "cpu", Severity: severity, Text: text}
}

func ioCheck(cap constants.Capacity, err error) constants.Check {
	if err != nil {
		return constants.Check{Name: "io", Severity: "warn", Text: err.Error()}
	}
	severity := psiSeverity(cap.Pressure.IO.SomeAvg10)
	if cap.CPUPercent.IOWait >= constants.HealthIOWaitCrit {
		severity = worse(severity, "crit")
	} else if cap.CPUPercent.IOWait >= constants.HealthIOWaitWarn {
		severity = worse(severity, "warn")
	}
	text := fmt.Sprintf("PSI io some avg10 %.2f, iowait %.1f%%", cap.Pressure.IO.SomeAvg10, cap.CPUPercent.IOWait)
	return constants.Check{Name: "io", Severity: severity, Text: text}
}

func netCheck() constants.Check {
	n := ReadNet()
	severity := "ok"
	if n.TCP.CloseWait >= constants.HealthCloseWaitCrit {
		severity = "crit"
	} else if n.TCP.CloseWait >= constants.HealthCloseWaitWarn {
		severity = "warn"
	}
	text := fmt.Sprintf("CLOSE_WAIT %d, retransmits %d cumulative, listen overflows %d cumulative",
		n.TCP.CloseWait, n.RetransSegs, n.ListenDrops.Overflows)
	return constants.Check{Name: "network", Severity: severity, Text: text}
}

func batteryCheck() constants.Check {
	power := ReadPower()
	if len(power.Batteries) == 0 {
		return constants.Check{Name: "battery", Severity: "ok", Text: "no battery"}
	}
	bat := power.Batteries[0]
	severity := "ok"
	if bat.FullRatio > 0 && bat.FullRatio < constants.HealthBatteryCrit {
		severity = "crit"
	} else if bat.FullRatio > 0 && bat.FullRatio < constants.HealthBatteryWarn {
		severity = "warn"
	}
	text := fmt.Sprintf("full charge is %.1f%% of design, %d cycles, now %d%%, drawing %.2f W",
		bat.FullRatio*100, bat.CycleCount, bat.CapacityPercent, bat.PowerWatts)
	return constants.Check{Name: "battery", Severity: severity, Text: text}
}

func psiSeverity(avg10 float64) string {
	if avg10 >= constants.HealthPSICrit {
		return "crit"
	}
	if avg10 >= constants.HealthPSIWarn {
		return "warn"
	}
	return "ok"
}

func worse(a, b string) string {
	if rank(b) > rank(a) {
		return b
	}
	return a
}

func rank(s string) int {
	switch s {
	case "crit":
		return 2
	case "warn":
		return 1
	default:
		return 0
	}
}

func gb(n uint64) string {
	return fmt.Sprintf("%.1f GB", float64(n)/1e9)
}
