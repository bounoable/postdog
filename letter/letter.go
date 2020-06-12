package letter

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/mail"
	"net/textproto"
	"path/filepath"
	"strings"
)

// Letter represents a mail.
type Letter struct {
	Subject     string
	From        mail.Address
	To          []mail.Address
	CC          []mail.Address
	BCC         []mail.Address
	Text        string
	HTML        string
	Attachments []Attachment
}

// HasTo determines if the letter has addr as a "To" recipient.
func (let Letter) HasTo(addr mail.Address) bool {
	return containsAddress(addr, let.To...)
}

func containsAddress(addr mail.Address, addresses ...mail.Address) bool {
	for _, a := range addresses {
		if a == addr {
			return true
		}
	}
	return false
}

// HasCC determines of the letter has addr as a "CC" recipient.
func (let Letter) HasCC(addr mail.Address) bool {
	return containsAddress(addr, let.CC...)
}

// HasBCC determines of the letter has addr as a "BCC" recipient.
func (let Letter) HasBCC(addr mail.Address) bool {
	return containsAddress(addr, let.BCC...)
}

// RFC builds the mail according to the RFC 2822 spec.
func (let Letter) RFC() string {
	toHeader := RecipientsHeader(let.To)
	ccHeader := RecipientsHeader(let.CC)
	bccHeader := RecipientsHeader(let.BCC)

	return rfc(let.From.String(), toHeader, ccHeader, bccHeader, let.Subject, let.Text, let.HTML, let.Attachments).String()
}

// RecipientsHeader builds the mail header value for the given recipients.
func RecipientsHeader(recipients []mail.Address) string {
	toMails := make([]string, len(recipients))
	for i, rec := range recipients {
		toMails[i] = rec.String()
	}

	return strings.Join(toMails, ",")
}

func (let Letter) String() string {
	b, err := json.Marshal(let)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// Attachment is a file attachment.
type Attachment struct {
	Filename    string
	Header      textproto.MIMEHeader
	ContentType string
	Content     []byte
}

// New ...
func New(opts ...WriteOption) Letter {
	return Write(opts...)
}

// Write builds a letter with the provided options.
func Write(opts ...WriteOption) Letter {
	var let Letter
	for _, opt := range opts {
		opt(&let)
	}
	return let
}

// WriteOption configures a letter.
type WriteOption func(*Letter)

// Must ...
func Must(opt WriteOption, err error) WriteOption {
	if err != nil {
		panic(err)
	}
	return opt
}

// Subject sets the subject of the letter.
func Subject(s string) WriteOption {
	return func(let *Letter) {
		let.Subject = s
	}
}

// FromAddress ...
func FromAddress(addr mail.Address) WriteOption {
	return func(let *Letter) {
		let.From = addr
	}
}

// From ...
func From(name, addr string) WriteOption {
	return FromAddress(mail.Address{
		Name:    name,
		Address: addr,
	})
}

// ToAddress ...
func ToAddress(addresses ...mail.Address) WriteOption {
	return func(let *Letter) {
		for _, addr := range addresses {
			if let.HasTo(addr) {
				continue
			}
			let.To = append(let.To, addr)
		}
	}
}

// To ...
func To(name, addr string) WriteOption {
	return ToAddress(mail.Address{
		Name:    name,
		Address: addr,
	})
}

// CCAddress ...
func CCAddress(addresses ...mail.Address) WriteOption {
	return func(let *Letter) {
		for _, addr := range addresses {
			if let.HasCC(addr) {
				continue
			}
			let.CC = append(let.CC, addr)
		}
	}
}

// CC ...
func CC(name, addr string) WriteOption {
	return CCAddress(mail.Address{
		Name:    name,
		Address: addr,
	})
}

// BCCAddress ...
func BCCAddress(addresses ...mail.Address) WriteOption {
	return func(let *Letter) {
		for _, addr := range addresses {
			if let.HasBCC(addr) {
				continue
			}
			let.BCC = append(let.BCC, addr)
		}
	}
}

// BCC ...
func BCC(name, addr string) WriteOption {
	return BCCAddress(mail.Address{
		Name:    name,
		Address: addr,
	})
}

// HTML ...
func HTML(content string) WriteOption {
	return func(let *Letter) {
		let.HTML = content
	}
}

// Text ...
func Text(content string) WriteOption {
	return func(let *Letter) {
		let.Text = content
	}
}

// Content ...
func Content(text, html string) WriteOption {
	return func(let *Letter) {
		let.Text = text
		let.HTML = html
	}
}

// Attach ...
func Attach(r io.Reader, filename string, opts ...AttachOption) (WriteOption, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	attach := Attachment{
		Filename: filename,
		Header:   make(textproto.MIMEHeader),
		Content:  b,
	}

	for _, opt := range opts {
		opt(&attach)
	}

	if attach.ContentType == "" {
		if ext := filepath.Ext(filename); ext != "" {
			attach.ContentType = mime.TypeByExtension(ext)
		} else {
			if err == nil {
				attach.ContentType = http.DetectContentType(b)
			}
		}
	}

	if attach.ContentType == "" {
		attach.ContentType = "application/octet-stream"
	}

	attach.Header.Set("Content-Type", fmt.Sprintf(`%s; name="%s"`, attach.ContentType, filename))
	attach.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	attach.Header.Set("Content-ID", fmt.Sprintf("<%s_%s>", fmt.Sprintf("%x", sha1.Sum(attach.Content))[:12], filename))
	attach.Header.Set("Content-Transfer-Encoding", "base64")

	return func(let *Letter) {
		let.Attachments = append(let.Attachments, attach)
	}, nil
}

// MustAttach ...
func MustAttach(r io.Reader, filename string, opts ...AttachOption) WriteOption {
	return Must(Attach(r, filename, opts...))
}

// AttachFile ...
func AttachFile(path, filename string, opts ...AttachOption) (WriteOption, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return Attach(bytes.NewReader(b), filename, opts...)
}

// MustAttachFile ...
func MustAttachFile(path, filename string, opts ...AttachOption) WriteOption {
	return Must(AttachFile(path, filename, opts...))
}

// AttachOption ...
type AttachOption func(*Attachment)

// ContentType ...
func ContentType(ct string) AttachOption {
	return func(attach *Attachment) {
		attach.ContentType = ct
	}
}
