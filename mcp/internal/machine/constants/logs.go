package constants

const (
	// journalctl prints the systemd journal. Kernel lines use the same command with -k.
	Journalctl = "journalctl"
	// Hard cap so a tool result cannot dump the whole journal.
	LogLineCap = 40
)

type LogQuery struct {
	Source   string
	Since    string
	Priority string
	Unit     string
	Match    string
}

type Logs struct {
	Source string   `json:"source"`
	Lines  []string `json:"lines"`
}
