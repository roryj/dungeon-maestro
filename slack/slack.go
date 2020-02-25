package slack

import (
	"fmt"
	"reflect"

	"github.com/roryj/dungeon-maestro/spells"
)

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

const (
	ShowResponseToAll = "in_channel"
)

type WebhookResponse struct {
	Text         string                      `json:"text"`
	ResponseType string                      `json:"response_type"`
	Attachments  []WebhookResponseAttachment `json:"attachments"`
}

type WebhookResponseAttachment struct {
	Title string `json:"title"`
	// ***** Used for images *****
	Fields     []WebhookResponseAttachmentField `json:"fields"`
	AuthorName string                           `json:"author_name"`
	AuthorIcon string                           `json:"author_icon"`
	ImageUrl   string                           `json:"image_url"`
	// ***** Used for simple text posts *****
	Text string `json:"text"`
	// ***** Used for polls *****
	Fallback       string                            `json:"fallback"`
	CallbackId     string                            `json:"callback_id"`
	Color          string                            `json:"color"`
	AttachmentType string                            `json:"attachment_type"`
	Actions        []WebhookResponseAttachmentAction `json:"actions"`
}

type WebhookResponseAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type WebhookResponseAttachmentAction struct {
	Name  string `json:"name"`
	Text  string `json:"text"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func NewWebhookResponseFromSpellResponse(sr spells.SpellResponse) WebhookResponse {
	spellAttributes := [][]string{
		{"Level", "Casting Time", "Range", "Duration"},
		{"Description"},
		{"Higher Levels"},
		{"Components", "School"},
		{"Classes", "Material"},
	}
	// get the number of separate rows to display in slack based on the number of spell attributes, and the max number
	// to display in a row. We use separate attachments to identify rows for slack, and separate fields within an
	// attachment to specify columns
	var attachments []WebhookResponseAttachment

	for _, row := range spellAttributes {

		var fields []WebhookResponseAttachmentField
		for _, a := range row {
			if a == "Description" || a == "Higher Levels" {
				fields = append(fields, getFieldForAttribute(sr, a, false))
			} else {
				fields = append(fields, getFieldForAttribute(sr, a, true))
			}
		}
		attachments = append(attachments, WebhookResponseAttachment{
			Fields: fields,
		})
	}

	return WebhookResponse{
		Text:         fmt.Sprintf("Description for %s", sr.Name),
		Attachments:  attachments,
		ResponseType: ShowResponseToAll,
	}
}

func getFieldForAttribute(sr spells.SpellResponse, attribute string, short bool) WebhookResponseAttachmentField {

	switch attribute {
	case "CastingTime":
		r := sr.CastingTime
		if sr.Ritual == "yes" {
			r = fmt.Sprintf("%s [R]", r)
		}
		return WebhookResponseAttachmentField{
			Title: "Casting Time",
			Value: r,
			Short: true,
		}
	case "Classes":
		return WebhookResponseAttachmentField{
			Title: "Classes",
			Value: sr.DnDClass,
			Short: true,
		}
	default:
		r := reflect.ValueOf(sr)
		f := reflect.Indirect(r).FieldByName(attribute)
		return WebhookResponseAttachmentField{
			Title: attribute,
			Value: f.String(),
			Short: true,
		}
	}
}
