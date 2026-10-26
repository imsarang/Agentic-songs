package google

import "context"

type EmailType int

const (
	EmailTypePrimary EmailType = iota
	EmailTypeSocial
	EmailTypePromotion
	EmailTypeSpam
	EmailTypeSent
	EmailTypeDraft
)

type Client interface {
	EnsureAuth(ctx context.Context) error
	SwitchAccount(ctx context.Context) (string, error)

	CreateDraft(ctx context.Context, to, subject, body string) (Draft, error)
	CreateSubject(ctx context.Context, subject string) error
	GetReciever(ctx context.Context, name string) ([]Contact, error)
	SearchMails(ctx context.Context, emailType EmailType, query string, limit int) ([]Mail, error)
	GetMessage(ctx context.Context, messageID string) (Message, error)
	DeleteMails(ctx context.Context, messageID string) error
	DeleteDrafts(ctx context.Context, ids []string) error
	SendMail(ctx context.Context, draftID string) (Mail, error)

	CheckDriveSpace(ctx context.Context) (DriveQuota, error)
}
