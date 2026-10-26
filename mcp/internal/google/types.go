package google

type Mail struct {
	Index    int
	ID       string
	ThreadID string
	From     string
	Subject  string
	Date     string
	Snippet  string
}

type Message struct {
	Mail
	Body string
}

type Draft struct {
	ID        string
	MessageID string
}

type Contact struct {
	Name  string
	Email string
}

type DriveQuota struct {
	Limit             string
	Usage             string
	UsageInDrive      string
	UsageInDriveTrash string
}
