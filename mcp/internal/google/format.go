package google

import (
	"fmt"
	"strconv"
	"strings"
)

func FormatMails(id int, rows []Mail) string {
	lines := []string{fmt.Sprintf("search_id=%d", id)}
	if len(rows) == 0 {
		lines = append(lines, "no messages")
		return strings.Join(lines, "\n")
	}
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf(
			"index=%d id=%s from=%s subject=%s date=%s snippet=%s",
			row.Index, row.ID, oneLine(row.From), oneLine(row.Subject), oneLine(row.Date), oneLine(row.Snippet),
		))
	}
	return strings.Join(lines, "\n")
}

func FormatMessage(m Message) string {
	body := strings.TrimSpace(m.Body)
	if body == "" {
		body = strings.TrimSpace(m.Snippet)
	}
	return fmt.Sprintf(
		"id=%s from=%s subject=%s date=%s\n%s",
		m.ID, oneLine(m.From), oneLine(m.Subject), oneLine(m.Date), body,
	)
}

func FormatDraft(d Draft) string {
	return fmt.Sprintf("draft_id=%s message_id=%s", d.ID, d.MessageID)
}

func FormatContacts(rows []Contact) string {
	if len(rows) == 0 {
		return "no contacts"
	}
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("name=%s email=%s", oneLine(row.Name), oneLine(row.Email)))
	}
	return strings.Join(lines, "\n")
}

func FormatDrive(q DriveQuota) string {
	return fmt.Sprintf(
		"limit=%s usage=%s usage_in_drive=%s usage_in_trash=%s",
		humanBytes(q.Limit), humanBytes(q.Usage), humanBytes(q.UsageInDrive), humanBytes(q.UsageInDriveTrash),
	)
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func humanBytes(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "unknown"
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil || n < 0 {
		return raw
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for n >= 1024 && i < len(units)-1 {
		n /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%.0f %s", n, units[i])
	}
	return fmt.Sprintf("%.1f %s", n, units[i])
}
