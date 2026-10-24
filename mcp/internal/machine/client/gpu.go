package client

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"shanks/internal/machine/constants"
	"shanks/internal/machine/proc"
)

func ReadGPUs() constants.GPUs {
	out := constants.GPUs{Devices: []constants.GPU{}, NVIDIA: readNVIDIA()}
	entries, err := os.ReadDir(constants.PathDRM)
	if err != nil {
		return out
	}
	for _, entry := range entries {
		if !cardDir(entry.Name()) {
			continue
		}
		dev := filepath.Join(constants.PathDRM, entry.Name(), "device")
		gpu := constants.GPU{
			Card:      entry.Name(),
			Vendor:    proc.Text(filepath.Join(dev, "vendor")),
			Device:    proc.Text(filepath.Join(dev, "device")),
			VRAMUsed:  proc.Uint(filepath.Join(dev, "mem_info_vram_used")),
			VRAMTotal: proc.Uint(filepath.Join(dev, "mem_info_vram_total")),
		}
		if raw := proc.Text(filepath.Join(dev, "gpu_busy_percent")); raw != "" {
			gpu.BusyPercent, _ = strconv.ParseFloat(raw, 64)
			gpu.BusyKnown = true
		}
		if raw := firstGlob(filepath.Join(dev, "hwmon", "hwmon*", "temp1_input")); raw != "" {
			milli, _ := strconv.ParseFloat(raw, 64)
			gpu.TempC = milli / 1000
			gpu.TempKnown = true
		}
		if raw := firstGlob(filepath.Join(dev, "hwmon", "hwmon*", "power1_input")); raw != "" {
			micro, _ := strconv.ParseFloat(raw, 64)
			gpu.PowerWatts = micro / 1e6
			gpu.PowerKnown = true
		}
		out.Devices = append(out.Devices, gpu)
	}
	return out
}

func cardDir(name string) bool {
	rest, ok := strings.CutPrefix(name, "card")
	if !ok || rest == "" {
		return false
	}
	for _, c := range rest {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func firstGlob(pattern string) string {
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}
	return proc.Text(matches[0])
}

func readNVIDIA() constants.NVIDIA {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "nvidia-smi",
		"--query-gpu=name,utilization.gpu,memory.used,memory.total,power.draw,temperature.gpu",
		"--format=csv,noheader")
	raw, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(raw))
	if err != nil {
		if text == "" {
			text = err.Error()
		}
		if len(text) > 200 {
			text = text[:200]
		}
		return constants.NVIDIA{Attached: false, Summary: text}
	}
	return constants.NVIDIA{Attached: true, Summary: text}
}
