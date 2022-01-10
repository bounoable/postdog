package letter

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime"
	"net/http"
	"net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/internal/encode"
	"github.com/bounoable/postdog/letter/mapper"
	"github.com/bounoable/postdog/letter/rfc"
)

// Letter represents a mail.
type Letter struct {
	L

	rfcConfig rfc.Config
}

// L contains the fields of a Letter.
type L struct {
	Subject     string
	From        mail.Address
	Recipients  []mail.Address
	To          []mail.Address
	CC          []mail.Address
	BCC         []mail.Address
	ReplyTo     []mail.Address
	RFC         string
	Text        string
	HTML        string
	Attachments []Attachment
}

// Attachment is a file attachment.
type Attachment struct {
	filename    string
	content     []byte
	contentType string
	size        int // overrides the actual size if != 0
	header      textproto.MIMEHeader
}

// Option modifies a letter.
type Option func(*Letter) error

// AttachmentOption configures an attachment.
type AttachmentOption func(*Attachment)

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
	let.normalize()

	return let, nil
}

// New is an alias for TryWrite().
func New(opts ...Option) (Letter, error) {
	return TryWrite(opts...)
}

// Subject sets the `Subject` header.
func Subject(s string) Option {
	return func(l *Letter) error {
		l.L.Subject = s
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
		l.L.From = addr
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
			if !containsAddress(l.L.Recipients, addr) {
				l.L.Recipients = append(l.L.Recipients, addr)
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
			if !containsAddress(l.L.To, addr) {
				l.L.To = append(l.L.To, addr)
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
			if !containsAddress(l.L.CC, addr) {
				l.L.CC = append(l.L.CC, addr)
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
			if !containsAddress(l.L.BCC, addr) {
				l.L.BCC = append(l.L.BCC, addr)
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
			if !containsAddress(l.L.ReplyTo, addr) {
				l.L.ReplyTo = append(l.L.ReplyTo, addr)
			}
		}
		return nil
	}
}

// Text sets the text content of the letter.
func Text(s string) Option {
	return func(l *Letter) error {
		l.L.Text = s
		return nil
	}
}

// HTML sets the HTML content of the letter.
func HTML(s string) Option {
	return func(l *Letter) error {
		l.L.HTML = s
		return nil
	}
}

// Content sets both the text and HTML content of the letter.
func Content(text, html string) Option {
	return func(l *Letter) error {
		l.L.Text = text
		l.L.HTML = html
		return nil
	}
}

// RFC returns an Option that
func RFC(body string) Option {
	return func(l *Letter) error {
		l.L.RFC = body
		return nil
	}
}

// Attach adds a file attachment to the letter.
func Attach(filename string, content []byte, opts ...AttachmentOption) Option {
	return func(l *Letter) error {
		l.L.Attachments = append(l.L.Attachments, NewAttachment(filename, content, opts...))
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

// AttachmentType sets the `Content-Type` of the attachment.
func AttachmentType(ct string) AttachmentOption {
	return func(at *Attachment) {
		at.contentType = ct
	}
}

// AttachmentSize returns an AttachmentOption that explicitly sets / overrides it's size.
func AttachmentSize(s int) AttachmentOption {
	return func(at *Attachment) {
		at.size = s
	}
}

// NewAttachment creates an Attachment from the given filename, content and opts.
func NewAttachment(filename string, content []byte, opts ...AttachmentOption) Attachment {
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

	return at
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
//
// If pm implements an RFCConfig() method, it will be used to add an rfc.Config
// to the Letter.
func Expand(pm postdog.Mail) Letter {
	if l, ok := pm.(Letter); ok {
		return l
	}

	letterOpts := []Option{
		FromAddress(pm.From()),
		RecipientAddress(pm.Recipients()...),
	}

	if toMail, ok := pm.(interface{ To() []mail.Address }); ok {
		letterOpts = append(letterOpts, ToAddress(toMail.To()...))
	}

	if ccMail, ok := pm.(interface{ CC() []mail.Address }); ok {
		letterOpts = append(letterOpts, CCAddress(ccMail.CC()...))
	}

	if bccMail, ok := pm.(interface{ BCC() []mail.Address }); ok {
		letterOpts = append(letterOpts, BCCAddress(bccMail.BCC()...))
	}

	if rtMail, ok := pm.(interface{ ReplyTo() []mail.Address }); ok {
		letterOpts = append(letterOpts, ReplyToAddress(rtMail.ReplyTo()...))
	}

	if sMail, ok := pm.(interface{ Subject() string }); ok {
		letterOpts = append(letterOpts, Subject(sMail.Subject()))
	}

	if textMail, ok := pm.(interface{ Text() string }); ok {
		letterOpts = append(letterOpts, Text(textMail.Text()))
	}

	if htmlMail, ok := pm.(interface{ HTML() string }); ok {
		letterOpts = append(letterOpts, HTML(htmlMail.HTML()))
	}

	if attachments := getAttachments(pm); len(attachments) > 0 {
		for _, at := range attachments {
			letterOpts = append(letterOpts, Attach(at.Filename(), at.Content(), AttachmentType(at.ContentType())))
		}
	}

	l := Write(letterOpts...)

	if rfcm, ok := pm.(interface{ RFCConfig() rfc.Config }); ok {
		l = l.WithRFCConfig(rfcm.RFCConfig())
	}

	return l
}

// Subject returns the subject of the letter.
func (l Letter) Subject() string {
	return l.L.Subject
}

// WithSubject returns a copy of l withs it's subject set to s.
func (l Letter) WithSubject(s string) Letter {
	l.L.Subject = s
	return l
}

// From returns the sender of the letter.
func (l Letter) From() mail.Address {
	return l.L.From
}

// WithFrom returns a copy of l with an updated `From` field.
func (l Letter) WithFrom(name, addr string) Letter {
	return l.WithFromAddress(mail.Address{Name: name, Address: addr})
}

// WithFromAddress returns a copy of l with addr as it's `From` field.
func (l Letter) WithFromAddress(addr mail.Address) Letter {
	l.L.From = addr
	return l
}

// To returns the `To` recipients of the letter.
func (l Letter) To() []mail.Address {
	return l.L.To
}

// WithTo returns a copy of l with addrs as it's `To` recipients.
func (l Letter) WithTo(addrs ...mail.Address) Letter {
	l.L.To = addrs
	return l
}

// CC returns the `Cc` recipients of the letter.
func (l Letter) CC() []mail.Address {
	return l.L.CC
}

// WithCC returns a copy of l with addrs as it's `CC` recipients.
func (l Letter) WithCC(addrs ...mail.Address) Letter {
	l.L.CC = addrs
	return l
}

// BCC returns the `Bcc` recipients of the letter.
func (l Letter) BCC() []mail.Address {
	return l.L.BCC
}

// WithBCC returns a copy of l with addrs as it's `BCC` recipients.
func (l Letter) WithBCC(addrs ...mail.Address) Letter {
	l.L.BCC = addrs
	return l
}

// ReplyTo returns the `Reply-To` recipients of the letter.
func (l Letter) ReplyTo() []mail.Address {
	return l.L.ReplyTo
}

// WithReplyTo returns a copy of l with addrs as it's `ReplyTo` field.
func (l Letter) WithReplyTo(addrs ...mail.Address) Letter {
	l.L.ReplyTo = addrs
	return l
}

// Recipients returns all recipients of the letter.
func (l Letter) Recipients() []mail.Address {
	count := len(l.L.Recipients) + len(l.L.To) + len(l.L.CC) + len(l.L.BCC)
	if count == 0 {
		return nil
	}
	rcpts := make([]mail.Address, 0, count)
	rcpts = append(rcpts, l.L.Recipients...)
	rcpts = append(rcpts, l.L.To...)
	rcpts = append(rcpts, l.L.CC...)
	rcpts = append(rcpts, l.L.BCC...)
	return rcpts
}

// WithRecipients returns a copy of l with addrs as it's recipients.
func (l Letter) WithRecipients(addrs ...mail.Address) Letter {
	l.L.Recipients = addrs
	return l
}

// Text returns the text content of the letter.
func (l Letter) Text() string {
	return l.L.Text
}

// WithText returns a copy of l with t as it's text content.
func (l Letter) WithText(t string) Letter {
	l.L.Text = t
	return l
}

// HTML returns the HTML content of the letter.
func (l Letter) HTML() string {
	return l.L.HTML
}

// WithHTML returns a copy of h with t as it's HTML content.
func (l Letter) WithHTML(h string) Letter {
	l.L.HTML = h
	return l
}

// Content returns both the text and HTML content of the letter.
func (l Letter) Content() (text string, html string) {
	return l.L.Text, l.L.HTML
}

// WithContent returns a copy of l with text as it's text content and html as
// it's HTML content.
func (l Letter) WithContent(text, html string) Letter {
	return l.WithText(text).WithHTML(html)
}

// Attachments returns the attachments of the letter.
func (l Letter) Attachments() []Attachment {
	return l.L.Attachments
}

// WithAttachments returns a copy of l with attach as it's attachments.
func (l Letter) WithAttachments(attach ...Attachment) Letter {
	l.L.Attachments = attach
	return l
}

// RFCConfig returns the RFC config that is used when calling l.L.RFC().
func (l Letter) RFCConfig() rfc.Config {
	return l.rfcConfig
}

// WithRFCOptions returns a copy of l with it's rfc configuration replaced.
func (l Letter) WithRFCOptions(opts ...rfc.Option) Letter {
	l.rfcConfig = rfc.Config{}
	for _, opt := range opts {
		opt(&l.rfcConfig)
	}
	return l
}

// WithRFCConfig returns a copy of l with it's rfc configuration replaced by cfg.
func (l Letter) WithRFCConfig(cfg rfc.Config) Letter {
	l.rfcConfig = cfg
	return l
}

// RFC returns the letter as a RFC 5322 string.
func (l Letter) RFC() string {
	if l.L.RFC != "" {
		return l.L.RFC
	}
	return rfc.BuildConfig(rfc.Mail{
		Subject:     l.Subject(),
		From:        l.From(),
		To:          l.To(),
		CC:          l.CC(),
		BCC:         l.BCC(),
		ReplyTo:     l.ReplyTo(),
		Text:        l.Text(),
		HTML:        l.HTML(),
		Attachments: rfcAttachments(l.Attachments()),
	}, l.rfcConfig)
}

// WithRFC returns a copy of l with it's rfc body replaced by rfc.
func (l Letter) WithRFC(rfc string) Letter {
	l.L.RFC = rfc
	return l
}

func (l Letter) String() string {
	return l.RFC()
}

// Map maps l to a map[string]interface{}. Use WithoutContent() option to
// clear the attachment contents in the map.
func (l Letter) Map(opts ...mapper.Option) map[string]interface{} {
	attachments := make([]interface{}, len(l.L.Attachments))
	for i, at := range l.L.Attachments {
		attachments[i] = at.Map(opts...)
	}

	var rfc string
	if l.L.RFC != "" {
		rfc = l.L.RFC
	}

	return map[string]interface{}{
		"from":        mapAddress(l.From()),
		"recipients":  mapAddresses(l.Recipients()...),
		"to":          mapAddresses(l.To()...),
		"cc":          mapAddresses(l.CC()...),
		"bcc":         mapAddresses(l.BCC()...),
		"replyTo":     mapAddresses(l.ReplyTo()...),
		"subject":     l.Subject(),
		"text":        l.Text(),
		"html":        l.HTML(),
		"rfc":         rfc,
		"attachments": attachments,
	}
}

// Parse parses m into l.
func (l *Letter) Parse(m map[string]interface{}) {
	if from, ok := m["from"].(map[string]interface{}); ok {
		l.L.From = parseAddress(from)
	}

	if recipients, ok := m["recipients"].([]interface{}); ok && len(recipients) > 0 {
		l.L.Recipients = parseIFaceAddresses(recipients...)
	}

	if to, ok := m["to"].([]interface{}); ok && len(to) > 0 {
		l.L.To = parseIFaceAddresses(to...)
	}

	if cc, ok := m["cc"].([]interface{}); ok && len(cc) > 0 {
		l.L.CC = parseIFaceAddresses(cc...)
	}

	if bcc, ok := m["bcc"].([]interface{}); ok && len(bcc) > 0 {
		l.L.BCC = parseIFaceAddresses(bcc...)
	}

	if replyTo, ok := m["replyTo"].([]interface{}); ok && len(replyTo) > 0 {
		l.L.ReplyTo = parseIFaceAddresses(replyTo...)
	}

	if subject, ok := m["subject"].(string); ok && len(subject) > 0 {
		l.L.Subject = subject
	}

	if text, ok := m["text"].(string); ok && len(text) > 0 {
		l.L.Text = text
	}

	if html, ok := m["html"].(string); ok && len(html) > 0 {
		l.L.HTML = html
	}

	if rfc, ok := m["rfc"].(string); ok && len(rfc) > 0 {
		l.L.RFC = rfc
	}

	if attachments, ok := m["attachments"].([]interface{}); ok && len(attachments) > 0 {
		ats := make([]Attachment, 0, len(attachments))
		for _, v := range attachments {
			if m, ok := v.(map[string]interface{}); ok {
				var at Attachment
				at.Parse(m)
				ats = append(ats, at)
			}
		}
		l.L.Attachments = ats
	}

	l.normalize()
}

func (l *Letter) normalize() {
	l.normalizeRecipients()
	l.normalizeAttachments()
}

func (l *Letter) normalizeRecipients() {
	l.removeRecipients(l.L.To)
	l.removeRecipients(l.L.CC)
	l.removeRecipients(l.L.BCC)
}

func (l *Letter) removeRecipients(addrs []mail.Address) {
	remaining := l.L.Recipients[:0]
L:
	for _, rcpt := range l.L.Recipients {
		for _, addr := range addrs {
			if rcpt == addr {
				continue L
			}
		}
		remaining = append(remaining, rcpt)
	}
	if len(remaining) == 0 {
		remaining = nil
	}
	l.L.Recipients = remaining
}

func (l *Letter) normalizeAttachments() {
	for i := range l.L.Attachments {
		l.L.Attachments[i].normalize()
	}
}

// Filename returns the filename of the Attachment.
func (at Attachment) Filename() string {
	return at.filename
}

// Size returns the filesize of the Attachment.
func (at Attachment) Size() int {
	if at.size != 0 {
		return at.size
	}
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

// Map maps at to a map[string]interface{}. Use WithoutContent() option to
// clear the attachment content in the map.
func (at Attachment) Map(opts ...mapper.Option) map[string]interface{} {
	cfg := mapper.Configure(opts...)

	m := map[string]interface{}{
		"filename":    at.filename,
		"content":     "",
		"size":        float64(at.Size()),
		"contentType": at.contentType,
		"header":      headerToMap(at.header),
	}

	if !cfg.WithoutAttachmentContent {
		m["content"] = base64.StdEncoding.EncodeToString(at.Content())
	}

	return m
}

// Parse parses the map m and applies the values to at.
func (at *Attachment) Parse(m map[string]interface{}) {
	if name, ok := m["filename"].(string); ok {
		at.filename = name
	}

	if content, ok := m["content"].(string); ok {
		if b, err := base64.StdEncoding.DecodeString(content); err == nil {
			at.content = b
		}
	}

	if contentType, ok := m["contentType"].(string); ok {
		at.contentType = contentType
	}

	if size, ok := m["size"].(float64); ok {
		at.size = int(math.Round(size))
	}

	if header, ok := m["header"].(map[string]interface{}); ok {
		at.header = mapToHeader(header)
	}

	at.normalize()
}

func (at *Attachment) normalize() {
	if at.size != 0 && at.size == len(at.content) {
		at.size = 0
	}
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

func mapAddress(addr mail.Address) map[string]interface{} {
	return map[string]interface{}{
		"name":    addr.Name,
		"address": addr.Address,
	}
}

func mapAddresses(addrs ...mail.Address) []interface{} {
	res := make([]interface{}, len(addrs))
	for i, addr := range addrs {
		res[i] = mapAddress(addr)
	}
	return res
}

func parseAddress(m map[string]interface{}) mail.Address {
	name, _ := m["name"].(string)
	addr, _ := m["address"].(string)
	return mail.Address{
		Name:    name,
		Address: addr,
	}
}

func parseIFaceAddresses(ifs ...interface{}) []mail.Address {
	addrs := make([]mail.Address, 0, len(ifs))
	for _, iface := range ifs {
		if m, ok := iface.(map[string]interface{}); ok {
			addrs = append(addrs, parseAddress(m))
		}
	}
	return addrs
}

func headerToMap(h textproto.MIMEHeader) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range h {
		iv := make([]interface{}, len(v))
		for i, vv := range v {
			iv[i] = vv
		}
		m[k] = iv
	}
	return m
}

func mapToHeader(m map[string]interface{}) textproto.MIMEHeader {
	h := make(textproto.MIMEHeader, len(m))
	for k, v := range m {
		vals, ok := v.([]interface{})
		if !ok {
			continue
		}
		svals := make([]string, 0, len(vals))
		for _, val := range vals {
			if sval, ok := val.(string); ok {
				svals = append(svals, sval)
			}
		}
		h[k] = svals
	}
	return h
}
