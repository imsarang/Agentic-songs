package google

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"shanks/internal/google/auth"
)

const (
	httpTimeout   = 20 * time.Second
	searchDefault = 10
	searchMax     = 20
	bodyMax       = 8000

	gmailAPI      = "https://gmail.googleapis.com/gmail/v1/users/me"
	draftsURL     = gmailAPI + "/drafts"
	draftsSendURL = gmailAPI + "/drafts/send"
	messagesURL   = gmailAPI + "/messages"
	trashPath     = "/trash"

	peopleSearchURL = "https://people.googleapis.com/v1/people:searchContacts"
	contactReadMask = "names,emailAddresses"

	driveAboutURL = "https://www.googleapis.com/drive/v3/about?fields=storageQuota"

	headerAuth        = "Authorization"
	headerContentType = "Content-Type"
	contentJSON       = "application/json"
	authBearerPrefix  = "Bearer "
	mailContentType   = "Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"

	formatMetadata = "metadata"
	formatFull     = "full"
	headerFrom     = "From"
	headerSubject  = "Subject"
	headerDate     = "Date"
	mimePlain      = "text/plain"

	boxPrimary    = "category:primary"
	boxSocial     = "category:social"
	boxPromotions = "category:promotions"
	boxSpam       = "in:spam"
	boxSent       = "in:sent"
)

type HTTPClient struct {
	http    *http.Client
	tokens  *auth.Source
	mu      sync.Mutex
	subject string
}

func NewHTTPClient(t *auth.Source) *HTTPClient {
	return &HTTPClient{
		http:   &http.Client{Timeout: httpTimeout},
		tokens: t,
	}
}

func (c *HTTPClient) EnsureAuth(ctx context.Context) error {
	return c.tokens.EnsureReady(ctx)
}

func (c *HTTPClient) SwitchAccount(ctx context.Context) (string, error) {
	if err := c.tokens.ReLogin(ctx); err != nil {
		return "", err
	}
	var out struct {
		EmailAddress string `json:"emailAddress"`
	}
	if err := c.call(ctx, http.MethodGet, gmailAPI+"/profile", nil, &out); err != nil {
		return "", err
	}
	return out.EmailAddress, nil
}

func (c *HTTPClient) CreateSubject(_ context.Context, subject string) error {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return fmt.Errorf("subject is required")
	}
	c.mu.Lock()
	c.subject = subject
	c.mu.Unlock()
	return nil
}

func (c *HTTPClient) CreateDraft(ctx context.Context, to, subject, body string) (Draft, error) {
	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)
	if subject == "" {
		c.mu.Lock()
		subject = c.subject
		c.mu.Unlock()
	}
	if subject == "" {
		return Draft{}, fmt.Errorf("subject is required")
	}
	var msg strings.Builder
	if to != "" {
		msg.WriteString("To: " + to + "\r\n")
	}
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString(mailContentType)
	msg.WriteString(body)
	raw := base64.RawURLEncoding.EncodeToString([]byte(msg.String()))
	var out gmailDraft
	err := c.call(ctx, http.MethodPost, draftsURL, map[string]any{
		"message": map[string]string{"raw": raw},
	}, &out)
	if err != nil {
		return Draft{}, err
	}
	if out.ID == "" {
		return Draft{}, fmt.Errorf("draft id missing")
	}
	return Draft{ID: out.ID, MessageID: out.Message.ID}, nil
}

func (c *HTTPClient) GetReciever(ctx context.Context, name string) ([]Contact, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	q := url.Values{
		"query":    {name},
		"readMask": {contactReadMask},
	}
	var out peopleResponse
	if err := c.call(ctx, http.MethodGet, peopleSearchURL+"?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	var rows []Contact
	for _, result := range out.Results {
		display := ""
		if len(result.Person.Names) > 0 {
			display = result.Person.Names[0].DisplayName
		}
		for _, email := range result.Person.Emails {
			if strings.TrimSpace(email.Value) == "" {
				continue
			}
			rows = append(rows, Contact{Name: display, Email: email.Value})
		}
	}
	return rows, nil
}

func (c *HTTPClient) SearchMails(ctx context.Context, emailType EmailType, query string, limit int) ([]Mail, error) {
	if limit <= 0 {
		limit = searchDefault
	}
	if limit > searchMax {
		limit = searchMax
	}
	if emailType == EmailTypeDraft {
		return c.searchDrafts(ctx, query, limit)
	}
	q := strings.TrimSpace(boxQuery(emailType) + " " + strings.TrimSpace(query))
	v := url.Values{
		"q":          {q},
		"maxResults": {fmt.Sprintf("%d", limit)},
	}
	var list gmailList
	if err := c.call(ctx, http.MethodGet, messagesURL+"?"+v.Encode(), nil, &list); err != nil {
		return nil, err
	}
	rows := make([]Mail, 0, len(list.Messages))
	for i, item := range list.Messages {
		mail, err := c.messageMeta(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		mail.Index = i
		rows = append(rows, mail)
	}
	return rows, nil
}

func (c *HTTPClient) searchDrafts(ctx context.Context, query string, limit int) ([]Mail, error) {
	v := url.Values{"maxResults": {fmt.Sprintf("%d", limit)}}
	if q := strings.TrimSpace(query); q != "" {
		v.Set("q", q)
	}
	var list struct {
		Drafts []struct {
			ID      string `json:"id"`
			Message struct {
				ID       string `json:"id"`
				ThreadID string `json:"threadId"`
			} `json:"message"`
		} `json:"drafts"`
	}
	if err := c.call(ctx, http.MethodGet, draftsURL+"?"+v.Encode(), nil, &list); err != nil {
		return nil, err
	}
	rows := make([]Mail, 0, len(list.Drafts))
	for i, item := range list.Drafts {
		mail, err := c.messageMeta(ctx, item.Message.ID)
		if err != nil {
			return nil, err
		}
		mail.Index = i
		mail.ID = item.ID
		if mail.ThreadID == "" {
			mail.ThreadID = item.Message.ThreadID
		}
		rows = append(rows, mail)
	}
	return rows, nil
}

func (c *HTTPClient) GetMessage(ctx context.Context, messageID string) (Message, error) {
	id := strings.TrimSpace(messageID)
	if id == "" {
		return Message{}, fmt.Errorf("message id is required")
	}
	v := url.Values{"format": {formatFull}}
	var raw gmailMessage
	err := c.call(ctx, http.MethodGet, messagesURL+"/"+url.PathEscape(id)+"?"+v.Encode(), nil, &raw)
	if err != nil {
		return Message{}, err
	}
	msg := Message{Mail: mailFrom(raw)}
	msg.Body = partText(raw.Payload)
	if len(msg.Body) > bodyMax {
		msg.Body = msg.Body[:bodyMax] + "\n[truncated]"
	}
	return msg, nil
}

func (c *HTTPClient) DeleteMails(ctx context.Context, messageID string) error {
	id := strings.TrimSpace(messageID)
	if id == "" {
		return fmt.Errorf("message id is required")
	}
	return c.call(ctx, http.MethodPost, messagesURL+"/"+url.PathEscape(id)+trashPath, nil, nil)
}

func (c *HTTPClient) DeleteDrafts(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("no drafts")
	}
	var failed []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if err := c.call(ctx, http.MethodDelete, draftsURL+"/"+url.PathEscape(id), nil, nil); err != nil {
			failed = append(failed, id)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("delete failed: %s", strings.Join(failed, ","))
	}
	return nil
}

func (c *HTTPClient) SendMail(ctx context.Context, draftID, to, name string) (Mail, error) {
	draftID = strings.TrimSpace(draftID)
	if draftID == "" {
		return Mail{}, fmt.Errorf("draft id is required")
	}
	email := strings.TrimSpace(to)
	if email == "" && strings.TrimSpace(name) != "" {
		found, err := c.onlyContact(ctx, name)
		if err != nil {
			return Mail{}, err
		}
		email = found
	}
	if email != "" {
		if err := c.addressDraft(ctx, draftID, email); err != nil {
			return Mail{}, err
		}
	}
	var raw gmailMessage
	err := c.call(ctx, http.MethodPost, draftsSendURL, map[string]string{"id": draftID}, &raw)
	if err != nil {
		return Mail{}, err
	}
	return mailFrom(raw), nil
}

func (c *HTTPClient) onlyContact(ctx context.Context, name string) (string, error) {
	rows, err := c.GetReciever(ctx, name)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("no contacts")
	}
	if len(rows) > 1 {
		lines := make([]string, 0, len(rows))
		for _, row := range rows {
			lines = append(lines, fmt.Sprintf("name=%s email=%s", oneLine(row.Name), oneLine(row.Email)))
		}
		return "", fmt.Errorf("multiple contacts\n%s", strings.Join(lines, "\n"))
	}
	return rows[0].Email, nil
}

func (c *HTTPClient) addressDraft(ctx context.Context, draftID, email string) error {
	v := url.Values{"format": {formatFull}}
	var draft struct {
		ID      string       `json:"id"`
		Message gmailMessage `json:"message"`
	}
	if err := c.call(ctx, http.MethodGet, draftsURL+"/"+url.PathEscape(draftID)+"?"+v.Encode(), nil, &draft); err != nil {
		return err
	}
	subject := headerValue(draft.Message.Payload.Headers, headerSubject)
	if subject == "" {
		return fmt.Errorf("subject is required")
	}
	body := partText(draft.Message.Payload)
	if body == "" {
		body = draft.Message.Snippet
	}
	raw := base64.RawURLEncoding.EncodeToString([]byte(
		"To: " + email + "\r\n" +
			"Subject: " + subject + "\r\n" +
			mailContentType +
			body,
	))
	return c.call(ctx, http.MethodPut, draftsURL+"/"+url.PathEscape(draftID), map[string]any{
		"id":      draftID,
		"message": map[string]string{"raw": raw},
	}, nil)
}

func (c *HTTPClient) CheckDriveSpace(ctx context.Context) (DriveQuota, error) {
	var out driveAbout
	if err := c.call(ctx, http.MethodGet, driveAboutURL, nil, &out); err != nil {
		return DriveQuota{}, err
	}
	q := out.StorageQuota
	return DriveQuota{
		Limit:             q.Limit,
		Usage:             q.Usage,
		UsageInDrive:      q.UsageInDrive,
		UsageInDriveTrash: q.UsageInDriveTrash,
	}, nil
}

// Drive write/search stubs — interface placeholders until implemented.
func (c *HTTPClient) SearchDrive(context.Context)          {}
func (c *HTTPClient) DeleteDriveMediaBulk(context.Context) {}
func (c *HTTPClient) AddToDrive(context.Context)           {}

func (c *HTTPClient) messageMeta(ctx context.Context, id string) (Mail, error) {
	v := url.Values{
		"format":          {formatMetadata},
		"metadataHeaders": {headerFrom, headerSubject, headerDate},
	}
	var raw gmailMessage
	err := c.call(ctx, http.MethodGet, messagesURL+"/"+url.PathEscape(id)+"?"+v.Encode(), nil, &raw)
	if err != nil {
		return Mail{}, err
	}
	return mailFrom(raw), nil
}

func boxQuery(t EmailType) string {
	switch t {
	case EmailTypeSocial:
		return boxSocial
	case EmailTypePromotion:
		return boxPromotions
	case EmailTypeSpam:
		return boxSpam
	case EmailTypeSent:
		return boxSent
	default:
		return boxPrimary
	}
}

func (c *HTTPClient) call(ctx context.Context, method, urlStr string, in, out any) error {
	var payload []byte
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		payload = b
	}
	for attempt := 0; attempt < 2; attempt++ {
		tok, err := c.tokens.AccessToken(ctx)
		if err != nil {
			return err
		}
		var body io.Reader
		if payload != nil {
			body = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
		if err != nil {
			return err
		}
		req.Header.Set(headerAuth, authBearerPrefix+tok)
		if payload != nil {
			req.Header.Set(headerContentType, contentJSON)
		}
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		b, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode == http.StatusUnauthorized && attempt == 0 {
			if err := c.tokens.ReLogin(ctx); err != nil {
				return err
			}
			continue
		}
		if res.StatusCode >= 300 {
			return fmt.Errorf("%s %s: %d %s", method, urlStr, res.StatusCode, b)
		}
		if out == nil || len(b) == 0 {
			return nil
		}
		if err := json.Unmarshal(b, out); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("%s %s: retry failed", method, urlStr)
}

func mailFrom(raw gmailMessage) Mail {
	return Mail{
		ID:       raw.ID,
		ThreadID: raw.ThreadID,
		From:     headerValue(raw.Payload.Headers, headerFrom),
		Subject:  headerValue(raw.Payload.Headers, headerSubject),
		Date:     headerValue(raw.Payload.Headers, headerDate),
		Snippet:  raw.Snippet,
	}
}

func headerValue(headers []gmailHeader, name string) string {
	for _, h := range headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

func partText(p gmailPart) string {
	if strings.EqualFold(p.MimeType, mimePlain) && p.Body.Data != "" {
		if text, ok := decodeMail(p.Body.Data); ok {
			return text
		}
	}
	for _, child := range p.Parts {
		if text := partText(child); text != "" {
			return text
		}
	}
	return ""
}

func decodeMail(data string) (string, bool) {
	for _, enc := range []*base64.Encoding{base64.RawURLEncoding, base64.URLEncoding} {
		b, err := enc.DecodeString(data)
		if err == nil {
			return string(b), true
		}
	}
	return "", false
}

type gmailList struct {
	Messages []struct {
		ID       string `json:"id"`
		ThreadID string `json:"threadId"`
	} `json:"messages"`
}

type gmailHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type gmailPart struct {
	MimeType string        `json:"mimeType"`
	Headers  []gmailHeader `json:"headers"`
	Body     struct {
		Data string `json:"data"`
	} `json:"body"`
	Parts []gmailPart `json:"parts"`
}

type gmailMessage struct {
	ID       string    `json:"id"`
	ThreadID string    `json:"threadId"`
	Snippet  string    `json:"snippet"`
	Payload  gmailPart `json:"payload"`
}

type gmailDraft struct {
	ID      string `json:"id"`
	Message struct {
		ID string `json:"id"`
	} `json:"message"`
}

type peopleResponse struct {
	Results []struct {
		Person struct {
			Names []struct {
				DisplayName string `json:"displayName"`
			} `json:"names"`
			Emails []struct {
				Value string `json:"value"`
			} `json:"emailAddresses"`
		} `json:"person"`
	} `json:"results"`
}

type driveAbout struct {
	StorageQuota struct {
		Limit             string `json:"limit"`
		Usage             string `json:"usage"`
		UsageInDrive      string `json:"usageInDrive"`
		UsageInDriveTrash string `json:"usageInDriveTrash"`
	} `json:"storageQuota"`
}

var _ Client = (*HTTPClient)(nil)
