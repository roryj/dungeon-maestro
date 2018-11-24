package slack

type Request struct {
	Token       string `json:"token"`
	TeamId      string `json:"team_id"`
	Text        string `json:"text"`
	TeamDomain  string `json:"team_domain"`
	ChannelId   string `json:"channel_id"`
	UserId      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Command     string `json:"command"`
	ResponseUrl string `json:"response_url"`
}

type WebhookResponse struct {
	Text string `json:"text"`
	Attachments []WebhookResponseAttachment
}

type WebhookResponseAttachment struct {
	Title string `json:"title"`
	// ***** Used for images *****
	Fields []WebhookResponseAttachmentField `json:"fields"`
	AuthorName string                       `json:"author_name"`
	AuthorIcon string                       `json:"author_icon"`
	ImageUrl string                         `json:"image_url"`
	// ***** Used for simple text posts *****
	Text string `json:"text"`
	// ***** Used for polls *****
	Fallback string                           `json:"fallback"`
	CallbackId string                         `json:"callback_id"`
	Color string                              `json:"color"`
	AttachmentType string                     `json:"attachment_type"`
	Actions []WebhookResponseAttachmentAction `json:"actions"`
}

type WebhookResponseAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool `json:"short"`
}

type WebhookResponseAttachmentAction struct {
	Name string `json:"name"`
	Text string `json:"text"`
	Type string `json:"type"`
	Value string `json:"value"`
}
