package letter

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/internal/encode"
	"github.com/bounoable/postdog/letter/rfc"
)

// Write a letter with the given opts. Panics if TryWrite() returns an error.
func Write(opts ...Option) Letter {
	return Must(TryWrite(opts...))
}

// Must panics if err is not nil and otherwise returns let.
func Must(let Letter, err error) Letter {
	if err != nil {
		panic(err)
	}
	return let
}

// TryWrite a letter with the given opts.
func TryWrite(opts ...Option) (Letter, error) {
	var let Letter
	var err error
	for _, opt := range opts {
		if err = opt(&let); err != nil {
			return let, err
		}
	}
	return let, nil
}

// New is an alias for TryWrite().
func New(opts ...Option) (Letter, error) {
	return TryWrite(opts...)
}

// Letter represents a mail.
type Letter struct {
	subject     string
	from        mail.Address
	recipients  []mail.Address
	to          []mail.Address
	cc          []mail.Address
	bcc         []mail.Address
	replyTo     []mail.Address
	rfc         string
	text        string
	html        string
	attachments []Attachment
}

// Option modifies a letter.
type Option func(*Letter) error

// Subject sets the `Subject` header.
func Subject(s string) Option {
	return func(l *Letter) error {
		l.subject = s
		return nil
	}
}

// From sets the sender of the letter.
func From(name, addr string) Option {
	return FromAddress(mail.Address{Name: name, Address: addr})
}

// FromAddress sets sender of the letter.
func FromAddress(addr mail.Address) Option {
	return func(l *Letter) error {
		l.from = addr
		return nil
	}
}

// Recipient returns an Option that adds a recipient to a mail.
// It does NOT add the recipient as to the `To` header of a mail.
func Recipient(name, addr string) Option {
	return RecipientAddress(mail.Address{Name: name, Address: addr})
}

// RecipientAddress returns an Option that adds a recipient to a mail.
// It does NOT add the recipient as to the `To` header of a mail.
func RecipientAddress(addrs ...mail.Address) Option {
	return func(l *Letter) error {
		for _, addr := range addrs {
			if !containsAddress(l.recipients, addr) {
				l.recipients = append(l.recipients, addr)
			}
		}
		return nil
	}
}

// To adds a `To` recipient to the letter.
func To(name, addr string) Option {
	return ToAddress(mail.Address{Name: name, Address: addr})
}

// ToAddress adds a `To` recipient to the letter.
func ToAddress(addrs ...mail.Address) Option {
	return func(l *Letter) error {
		for _, addr := range addrs {
			if !containsAddress(l.to, addr) {
				l.to = append(l.to, addr)
			}
		}
		return nil
	}
}

// CC adds a `Cc` recipient to the letter.
func CC(name, addr string) Option {
	return CCAddress(mail.Address{Name: name, Address: addr})
}

// CCAddress adds a `Cc` recipient to the letter.
func CCAddress(addrs ...mail.Address) Option {
	return func(l *Letter) error {
		for _, addr := range addrs {
			if !containsAddress(l.cc, addr) {
				l.cc = append(l.cc, addr)
			}
		}
		return nil
	}
}

// BCC adds a `Bcc` recipient to the letter.
func BCC(name, addr string) Option {
	return BCCAddress(mail.Address{Name: name, Address: addr})
}

// BCCAddress adds a `Bcc` recipient to the letter.
func BCCAddress(addrs ...mail.Address) Option {
	return func(l *Letter) error {
		for _, addr := range addrs {
			if !containsAddress(l.bcc, addr) {
				l.bcc = append(l.bcc, addr)
			}
		}
		return nil
	}
}

// ReplyTo adds a `Reply-To` recipient to the letter.
func ReplyTo(name, addr string) Option {
	return ReplyToAddress(mail.Address{Name: name, Address: addr})
}

// ReplyToAddress adds a `Reply-To` recipient to the letter.
func ReplyToAddress(addrs ...mail.Address) Option {
	return func(l *Letter) error {
		for _, addr := range addrs {
			if !containsAddress(l.replyTo, addr) {
				l.replyTo = append(l.replyTo, addr)
			}
		}
		return nil
	}
}

// Text sets the text content of the letter.
func Text(s string) Option {
	return func(l *Letter) error {
		l.text = s
		return nil
	}
}

// HTML sets the HTML content of the letter.
func HTML(s string) Option {
	return func(l *Letter) error {
		l.html = s
		return nil
	}
}

// Content sets both the text and HTML content of the letter.
func Content(text, html string) Option {
	return func(l *Letter) error {
		l.text = text
		l.html = html
		return nil
	}
}

// RFC returns an Option that
func RFC(body string) Option {
	return func(l *Letter) error {
		l.rfc = body
		return nil
	}
}

// Attach adds a file attachment to the letter.
func Attach(filename string, content []byte, opts ...AttachmentOption) Option {
	return func(l *Letter) error {
		at := Attachment{
			filename: filename,
			content:  content,
			header:   make(textproto.MIMEHeader),
		}

		for _, opt := range opts {
			opt(&at)
		}

		if at.contentType == "" {
			if ext := filepath.Ext(filename); ext != "" {
				at.contentType = mime.TypeByExtension(ext)
			}
		}

		if at.contentType == "" {
			at.contentType = http.DetectContentType(content)
		}

		filename8 := encode.UTF8(at.filename)
		filenameASCII := encode.ToASCII(at.filename)

		at.header.Set("Content-Type", fmt.Sprintf(`%s; name="%s"`, at.contentType, filename8))
		at.header.Set("Content-ID", fmt.Sprintf("<%s_%s>", fmt.Sprintf("%x", sha1.Sum(at.Content()))[:12], filenameASCII))
		at.header.Set("Content-Disposition", fmt.Sprintf(`attachment; size=%d; filename="%s"`, at.Size(), filename8))
		at.header.Set("Content-Transfer-Encoding", "base64")

		l.attachments = append(l.attachments, at)

		return nil
	}
}

// AttachReader adds a file attachment to the letter.
func AttachReader(filename string, r io.Reader, opts ...AttachmentOption) Option {
	return func(l *Letter) error {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		return Attach(filename, b, opts...)(l)
	}
}

// AttachFile adds the file in path as an attachment to the letter.
func AttachFile(filename, path string, opts ...AttachmentOption) Option {
	return func(l *Letter) error {
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file %s: %w", path, err)
		}

		if err = AttachReader(filename, f, opts...)(l); err != nil {
			if err = f.Close(); err != nil {
				return fmt.Errorf("close file %s: %w", path, err)
			}
			return err
		}

		if err = f.Close(); err != nil {
			return fmt.Errorf("close file %s: %w", path, err)
		}

		return nil
	}
}

// AttachmentOption configures an attachment.
type AttachmentOption func(*Attachment)

// ContentType sets the `Content-Type` of the attachment.
func ContentType(ct string) AttachmentOption {
	return func(at *Attachment) {
		at.contentType = ct
	}
}

// Expand converts the postdog.Mail pm to a Letter.
//
// Add additional information
//
// If pm implements any of the optional methods To(), CC(), BCC(), ReplyTo(),
// Subject(), Text(), HTML() or Attachments(), those methods will be called to
// retrieve the information which will be added to the returned Letter.
//
// If pm has an Attachments() method, the return type of that method must be
// a slice of a type that implements the following methods: Filename() string,
// Content() []byte, ContentType() string, Header() textproto.MIMEHeader.
func Expand(pm postdog.Mail) Letter {
	opts := []Option{
		FromAddress(pm.From()),
		RecipientAddress(pm.Recipients()...),
		RFC(pm.RFC()),
	}

	if toMail, ok := pm.(interface{ To() []mail.Address }); ok {
		opts = append(opts, ToAddress(toMail.To()...))
	}

	if ccMail, ok := pm.(interface{ CC() []mail.Address }); ok {
		opts = append(opts, CCAddress(ccMail.CC()...))
	}

	if bccMail, ok := pm.(interface{ BCC() []mail.Address }); ok {
		opts = append(opts, BCCAddress(bccMail.BCC()...))
	}

	if rtMail, ok := pm.(interface{ ReplyTo() []mail.Address }); ok {
		opts = append(opts, ReplyToAddress(rtMail.ReplyTo()...))
	}

	if sMail, ok := pm.(interface{ Subject() string }); ok {
		opts = append(opts, Subject(sMail.Subject()))
	}

	if textMail, ok := pm.(interface{ Text() string }); ok {
		opts = append(opts, Text(textMail.Text()))
	}

	if htmlMail, ok := pm.(interface{ HTML() string }); ok {
		opts = append(opts, HTML(htmlMail.HTML()))
	}

	l := Write(opts...)
	l.attachments = getAttachments(pm)

	return l
}

// Subject returns the subject of the letter.
func (l Letter) Subject() string {
	return l.subject
}

// From returns the sender of the letter.
func (l Letter) From() mail.Address {
	return l.from
}

// To returns the `To` recipients of the letter.
func (l Letter) To() []mail.Address {
	return l.to
}

// CC returns the `Cc` recipients of the letter.
func (l Letter) CC() []mail.Address {
	return l.cc
}

// BCC returns the `Bcc` recipients of the letter.
func (l Letter) BCC() []mail.Address {
	return l.bcc
}

// ReplyTo returns the `Reply-To` recipients of the letter.
func (l Letter) ReplyTo() []mail.Address {
	return l.replyTo
}

// Recipients returns all recipients of the letter.
func (l Letter) Recipients() []mail.Address {
	count := len(l.recipients) + len(l.to) + len(l.cc) + len(l.bcc)
	if count == 0 {
		return nil
	}
	rcpts := make([]mail.Address, 0, count)
	rcpts = append(rcpts, l.recipients...)
	rcpts = append(rcpts, l.to...)
	rcpts = append(rcpts, l.cc...)
	rcpts = append(rcpts, l.bcc...)
	return rcpts
}

// Text returns the text content of the letter.
func (l Letter) Text() string {
	return l.text
}

// HTML returns the HTML content of the letter.
func (l Letter) HTML() string {
	return l.html
}

// Content returns both the text and HTML content of the letter.
func (l Letter) Content() (text string, html string) {
	return l.text, l.html
}

// Attachments returns the attachments of the letter.
func (l Letter) Attachments() []Attachment {
	return l.attachments
}

// RFC returns the letter as a RFC 5322 string.
func (l Letter) RFC() string {
	if l.rfc != "" {
		return l.rfc
	}
	return rfc.Build(rfc.Mail{
		Subject:     l.Subject(),
		From:        l.From(),
		To:          l.To(),
		CC:          l.CC(),
		BCC:         l.BCC(),
		ReplyTo:     l.ReplyTo(),
		Text:        l.Text(),
		HTML:        l.HTML(),
		Attachments: rfcAttachments(l.Attachments()),
	})
}

func (l Letter) String() string {
	return l.RFC()
}

func rfcAttachments(ats []Attachment) []rfc.Attachment {
	res := make([]rfc.Attachment, len(ats))
	for i, at := range ats {
		res[i] = rfc.Attachment{
			Filename: at.Filename(),
			Content:  at.Content(),
			Header:   at.Header(),
		}
	}
	return res
}

func containsAddress(addrs []mail.Address, addr mail.Address) bool {
	for _, a := range addrs {
		if a == addr {
			return true
		}
	}
	return false
}

// Attachment is a file attachment.
type Attachment struct {
	filename    string
	content     []byte
	contentType string
	header      textproto.MIMEHeader
}

// Filename returns the filename of the Attachment.
func (at Attachment) Filename() string {
	return at.filename
}

// Size returns the filesize of the Attachment.
func (at Attachment) Size() int {
	return len(at.content)
}

// Content returns the file contents of the Attachment.
func (at Attachment) Content() []byte {
	return at.content
}

// ContentType returns the `Content-Type` of the Attachment.
func (at Attachment) ContentType() string {
	return at.contentType
}

// Header returns the MIME headers of the Attachment.
func (at Attachment) Header() textproto.MIMEHeader {
	return at.header
}

// getAttachments returns m.Attachments() as an []Attachment slice if the
// return type of m.Attachments() is a slice of a type that implements the
// methods Filename() string, Content() []byte, ContentType() string and
// Header() textproto.MIMEHeader.
func getAttachments(m postdog.Mail) []Attachment {
	type attachment interface {
		Filename() string
		Content() []byte
		ContentType() string
		Header() textproto.MIMEHeader
	}

	typ := reflect.TypeOf(m)
	val := reflect.ValueOf(m)

	method, ok := typ.MethodByName("Attachments")
	if !ok {
		return nil
	}

	methodType := method.Type
	if methodType.NumOut() != 1 {
		return nil
	}

	outType := methodType.Out(0)
	if outType.Kind() != reflect.Slice {
		return nil
	}

	outSliceType := outType.Elem()

	attachType := reflect.TypeOf(new(attachment)).Elem()
	if !outSliceType.Implements(attachType) {
		return nil
	}

	ret := method.Func.Call([]reflect.Value{val})
	if len(ret) == 0 {
		return nil
	}

	len := ret[0].Len()
	if len == 0 {
		return nil
	}

	result := make([]Attachment, len)

	for i := 0; i < len; i++ {
		at := ret[0].Index(i).Interface().(attachment)
		result[i] = Attachment{
			filename:    at.Filename(),
			content:     at.Content(),
			contentType: at.ContentType(),
			header:      at.Header(),
		}
	}

	return result
}
