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
	"regexp"
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
	Filename string
	Header   textproto.MIMEHeader
	Content  []byte
}

// New is an alias to Write()
func New(opts ...WriteOption) Letter {
	return Write(opts...)
}

// Write builds a letter with the provided options.
func Write(opts ...WriteOption) Letter {
	let := Letter{
		To:          []mail.Address{},
		CC:          []mail.Address{},
		BCC:         []mail.Address{},
		Attachments: []Attachment{},
	}
	for _, opt := range opts {
		opt(&let)
	}
	return let
}

// WriteOption configures a letter.
type WriteOption func(*Letter)

// Must returns opt if err is nil and panics otherwise.
func Must(opt WriteOption, err error) WriteOption {
	if err != nil {
		panic(err)
	}
	return opt
}

// Subject sets the subject of a letter.
func Subject(s string) WriteOption {
	return func(let *Letter) {
		let.Subject = s
	}
}

// FromAddress sets the "From" field of a letter.
func FromAddress(addr mail.Address) WriteOption {
	return func(let *Letter) {
		let.From = addr
	}
}

// From sets the "From" field of a letter.
func From(name, addr string) WriteOption {
	return FromAddress(mail.Address{
		Name:    name,
		Address: addr,
	})
}

// ToAddress adds a recipient to the "To" field of a letter.
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

// To adds a recipient to the "To" field of a letter.
func To(name, addr string) WriteOption {
	return ToAddress(mail.Address{
		Name:    name,
		Address: addr,
	})
}

// CCAddress adds a recipient to the "CC" field of a letter.
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

// CC adds a recipient to the "CC" field of a letter.
func CC(name, addr string) WriteOption {
	return CCAddress(mail.Address{
		Name:    name,
		Address: addr,
	})
}

// BCCAddress adds a recipient to the "BCC" field of a letter.
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

// BCC adds a recipient to the "BCC" field of a letter.
func BCC(name, addr string) WriteOption {
	return BCCAddress(mail.Address{
		Name:    name,
		Address: addr,
	})
}

// HTML sets the HTML body of a letter.
func HTML(content string) WriteOption {
	return func(let *Letter) {
		let.HTML = content
	}
}

// Text sets the text body of a letter.
func Text(content string) WriteOption {
	return func(let *Letter) {
		let.Text = content
	}
}

// Content sets both the HTML and text body of a letter.
func Content(text, html string) WriteOption {
	return func(let *Letter) {
		HTML(html)(let)
		Text(text)(let)
	}
}

// Attach attaches the content of r to a letter, with the given filename.
// It returns an error if it fails to read from r.
// Available options:
//	ContentType(): Set the attachment's "Content-Type" header.
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

	ct := attach.ContentType()

	if ct == "" {
		if ext := filepath.Ext(filename); ext != "" {
			ct = mime.TypeByExtension(ext)
		} else if err == nil {
			ct = http.DetectContentType(b)
		}
	}

	if ct == "" {
		ct = "application/octet-stream"
	}

	attach.Header.Set("Content-Type", fmt.Sprintf(`%s; name="%s"`, ct, filename))
	attach.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	attach.Header.Set("Content-ID", fmt.Sprintf("<%s_%s>", fmt.Sprintf("%x", sha1.Sum(attach.Content))[:12], filename))
	attach.Header.Set("Content-Transfer-Encoding", "base64")

	return func(let *Letter) {
		let.Attachments = append(let.Attachments, attach)
	}, nil
}

// MustAttach does the same as Attach(), but panic if Attach() returns an error.
func MustAttach(r io.Reader, filename string, opts ...AttachOption) WriteOption {
	return Must(Attach(r, filename, opts...))
}

// AttachFile attaches the file at path to a letter, with the given filename.
// It returns an error if it fails to read the file.
// Available options:
//	ContentType(): Set the attachment's "Content-Type" header.
func AttachFile(path, filename string, opts ...AttachOption) (WriteOption, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return Attach(bytes.NewReader(b), filename, opts...)
}

// MustAttachFile does the same as AttachFile(), but panics if AttachFile() returns an error.
func MustAttachFile(path, filename string, opts ...AttachOption) WriteOption {
	return Must(AttachFile(path, filename, opts...))
}

// AttachOption is an attachment option.
type AttachOption func(*Attachment)

// ContentType sets the "Content-Type" header of an attachment.
func ContentType(ct string) AttachOption {
	return func(attach *Attachment) {
		attach.Header.Set("Content-Type", fmt.Sprintf(`%s; name="%s"`, ct, attach.Filename))
	}
}

var attachmentNameExpr = regexp.MustCompile(`^(.+); name=".+"$`)

// ContentType returns the "Content-Type" of the attachment.
func (attach Attachment) ContentType() string {
	return attachmentNameExpr.ReplaceAllString(attach.Header.Get("Content-Type"), "$1")
}
