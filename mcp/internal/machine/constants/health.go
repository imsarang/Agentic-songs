package constants

const (
	// Filesystems smaller than this are ignored by the disk check. /boot/efi is tiny.
	HealthDiskMinBytes = 1 << 30
	// Free space under 20% is warn. Under 10% is crit.
	HealthDiskWarn = 20.0
	HealthDiskCrit = 10.0
	// PSI some avg10. Memory and IO share the lower pair. CPU uses the higher pair.
	HealthPSIWarn = 5.0
	HealthPSICrit = 20.0
	HealthCPUWarn = 15.0
	HealthCPUCrit = 40.0
	// iowait percent from the 200ms CPU sample.
	HealthIOWaitWarn = 10.0
	HealthIOWaitCrit = 20.0
	// CLOSE_WAIT sockets. The app is not closing them.
	HealthCloseWaitWarn = 50
	HealthCloseWaitCrit = 200
	// Full charge divided by design charge.
	HealthBatteryWarn = 0.80
	HealthBatteryCrit = 0.60
	// Swap used divided by swap total.
	HealthSwapWarn = 0.50
	HealthSwapCrit = 0.90
)

type Health struct {
	Overall string  `json:"overall"`
	Checks  []Check `json:"checks"`
}

type Check struct {
	Name     string `json:"name"`
	Severity string `json:"severity"`
	Text     string `json:"text"`
}
