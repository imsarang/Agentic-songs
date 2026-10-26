package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/google"
	"shanks/internal/google/cache"
)

func Register(s *server.MCPServer, client google.Client) {
	found := &cache.Search{}

	s.AddTool(mcp.NewTool("google_switch_account",
		mcp.WithDescription("Sign in again and pick a Google account. Opens the account chooser. Later mail tools use the account you pick."),
	), switchAccountHandler(client))

	s.AddTool(mcp.NewTool("google_create_subject",
		mcp.WithDescription("Store a subject for the next google_create_draft when that call omits subject."),
		mcp.WithString("subject", mcp.Required(), mcp.Description("Mail subject")),
	), createSubjectHandler(client))

	s.AddTool(mcp.NewTool("google_get_receiver",
		mcp.WithDescription("Look up email addresses for a name in the user's contacts. Returns no contacts when nothing matches. Do not invent an address."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Person name to search")),
	), getReceiverHandler(client))

	s.AddTool(mcp.NewTool("google_create_draft",
		mcp.WithDescription("Create a Gmail draft. Does not send. Returns draft_id for google_send_mail."),
		mcp.WithString("to", mcp.Required(), mcp.Description("Recipient email from google_get_receiver")),
		mcp.WithString("subject", mcp.Description("Subject. Uses the stored subject when empty.")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Plain-text body")),
	), createDraftHandler(client))

	s.AddTool(mcp.NewTool("google_send_mail",
		mcp.WithDescription("Send an existing draft by draft_id. Call only after the user asks to send that draft."),
		mcp.WithString("draft_id", mcp.Required(), mcp.Description("draft_id from google_create_draft")),
	), sendMailHandler(client))

	searchTool := func(name, desc string, kind google.EmailType) {
		s.AddTool(mcp.NewTool(name,
			mcp.WithDescription(desc),
			mcp.WithString("query", mcp.Description("Extra Gmail search text")),
			mcp.WithNumber("limit", mcp.Description("How many rows to return, from 1 to 20")),
		), searchBox(client, found, kind))
	}
	searchTool("google_search_inbox", "Search Primary inbox. Returns search_id and message handles. Does not return full bodies.", google.EmailTypePrimary)
	searchTool("google_search_social", "Search Social mail.", google.EmailTypeSocial)
	searchTool("google_search_promotions", "Search Promotions mail.", google.EmailTypePromotion)
	searchTool("google_search_spam", "Search Spam.", google.EmailTypeSpam)
	searchTool("google_search_sent", "Search Sent mail.", google.EmailTypeSent)
	searchTool("google_search_drafts", "Search Gmail drafts. Returns search_id and draft ids. Does not delete.", google.EmailTypeDraft)

	s.AddTool(mcp.NewTool("google_delete_drafts",
		mcp.WithDescription("Delete drafts from the latest google_search_drafts. Pass result_indexes like 0,2. Omit it to delete every row in that search. This removes drafts, it does not trash inbox mail."),
		mcp.WithNumber("search_id", mcp.Required(), mcp.Description("search_id from google_search_drafts")),
		mcp.WithString("result_indexes", mcp.Description("Comma-separated indexes. Empty deletes every row in that search.")),
	), deleteDraftsHandler(client, found))

	s.AddTool(mcp.NewTool("google_get_message",
		mcp.WithDescription("Read one message body. Pass message_id, or search_id plus result_index from the latest search."),
		mcp.WithString("message_id", mcp.Description("Gmail message id")),
		mcp.WithNumber("search_id", mcp.Description("search_id from the latest mailbox search")),
		mcp.WithNumber("result_index", mcp.Description("index from that search")),
	), getMessageHandler(client, found))

	s.AddTool(mcp.NewTool("google_delete_mail",
		mcp.WithDescription("Move mail to trash. Works for inbox, social, promotions, spam, and sent. Pass message_id, search_id plus result_index, or search_id plus result_indexes like 0,2. Omit indexes to trash every row in that search. Does not delete drafts."),
		mcp.WithString("message_id", mcp.Description("Gmail message id")),
		mcp.WithNumber("search_id", mcp.Description("search_id from the latest mailbox search")),
		mcp.WithNumber("result_index", mcp.Description("One index from that search")),
		mcp.WithString("result_indexes", mcp.Description("Comma-separated indexes. Empty with search_id trashes every row in that search.")),
	), deleteMailHandler(client, found))

	s.AddTool(mcp.NewTool("google_drive_space",
		mcp.WithDescription("Read Google Drive storage quota."),
	), driveSpaceHandler(client))
}
