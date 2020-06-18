package mongostore

import (
	"net/mail"
	"net/textproto"
	"time"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/store"
)

type dbLetter struct {
	Subject     string         `bson:"subject"`
	From        mail.Address   `bson:"from"`
	To          []mail.Address `bson:"to"`
	CC          []mail.Address `bson:"cc"`
	BCC         []mail.Address `bson:"bcc"`
	Text        string         `bson:"text"`
	HTML        string         `bson:"html"`
	Attachments []attachment   `bson:"attachments"`
	SentAt      time.Time      `bson:"sentAt"`
	SendError   string         `bson:"sendError"`
}

type attachment struct {
	Filename    string               `bson:"filename"`
	Header      textproto.MIMEHeader `bson:"header"`
	ContentType string               `bson:"contentType"`
	Content     []byte               `bson:"content"`
	Size        int                  `bson:"size"`
}

func mapLetter(let store.Letter) dbLetter {
	return dbLetter{
		Subject:     let.Subject,
		From:        let.From,
		To:          let.To,
		CC:          let.CC,
		BCC:         let.BCC,
		Text:        let.Text,
		HTML:        let.HTML,
		Attachments: mapAttachments(let.Attachments),
		SentAt:      let.SentAt,
		SendError:   let.SendError,
	}
}

func mapAttachments(attachments []letter.Attachment) []attachment {
	mapped := make([]attachment, len(attachments))
	for i, attach := range attachments {
		mapped[i] = attachment{
			Filename:    attach.Filename,
			Header:      attach.Header,
			ContentType: attach.ContentType,
			Content:     attach.Content,
			Size:        len(attach.Content),
		}
	}
	return mapped
}

func (let dbLetter) store() store.Letter {
	return store.Letter{
		Letter: letter.Letter{
			Subject:     let.Subject,
			From:        let.From,
			To:          let.To,
			CC:          let.CC,
			BCC:         let.BCC,
			Text:        let.Text,
			HTML:        let.HTML,
			Attachments: let.storeAttachments(),
		},
		SentAt:    let.SentAt,
		SendError: let.SendError,
	}
}

func (let dbLetter) storeAttachments() []letter.Attachment {
	attachments := make([]letter.Attachment, len(let.Attachments))
	for i, attach := range let.Attachments {
		attachments[i] = letter.Attachment{
			Filename:    attach.Filename,
			Header:      attach.Header,
			ContentType: attach.ContentType,
			Content:     attach.Content,
		}
	}
	return attachments
}
