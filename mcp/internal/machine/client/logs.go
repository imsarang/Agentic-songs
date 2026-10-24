package client

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"shanks/internal/machine/constants"
)

func ReadLogs(ctx context.Context, q constants.LogQuery) (constants.Logs, error) {
	source := strings.ToLower(strings.TrimSpace(q.Source))
	if source == "" {
		source = "journal"
	}
	if source != "journal" && source != "kernel" {
		return constants.Logs{}, fmt.Errorf("source must be journal or kernel")
	}
	if q.Priority != "" && !allowedPriority(q.Priority) {
		return constants.Logs{}, fmt.Errorf("priority must be emerg, alert, crit, err, warning, notice, info, debug, or 0-7")
	}
	if !allowedToken(q.Unit, 128) || !allowedToken(q.Since, 48) {
		return constants.Logs{}, fmt.Errorf("unit or since has an unsupported character")
	}
	if strings.ContainsAny(q.Match, "\n\r") || len(q.Match) > 80 {
		return constants.Logs{}, fmt.Errorf("match must be one line of at most 80 characters")
	}
	args := []string{"--no-pager", "-o", "short-iso", "-n", fmt.Sprint(constants.LogLineCap)}
	if source == "kernel" {
		args = append(args, "-k")
	}
	if q.Priority != "" {
		args = append(args, "-p", q.Priority)
	}
	if q.Unit != "" {
		args = append(args, "-u", q.Unit)
	}
	if q.Since != "" {
		args = append(args, "-S", q.Since)
	}
	if q.Match != "" {
		args = append(args, "-g", q.Match)
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	raw, err := exec.CommandContext(ctx, constants.Journalctl, args...).Output()
	if err != nil && len(raw) == 0 {
		return constants.Logs{}, err
	}
	lines := []string{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) >= constants.LogLineCap {
			break
		}
	}
	return constants.Logs{Source: source, Lines: lines}, nil
}

func allowedPriority(v string) bool {
	switch strings.ToLower(v) {
	case "0", "1", "2", "3", "4", "5", "6", "7", "emerg", "alert", "crit", "err", "warning", "notice", "info", "debug":
		return true
	default:
		return false
	}
}

func allowedToken(v string, max int) bool {
	if len(v) > max {
		return false
	}
	for _, r := range v {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune(" ._:+-", r) {
			continue
		}
		return false
	}
	return true
}
