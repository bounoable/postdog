package mongo

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/textproto"
	"time"

	"github.com/bounoable/mongoutil/index"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/archive"
	"github.com/bounoable/postdog/plugin/archive/query"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Store is the mongo store.
type Store struct {
	client                   *mongo.Client
	databaseName             string
	collectionName           string
	wantIndexes              bool
	withoutAttachmentContent bool
	col                      *mongo.Collection
}

// Option is a Store option.
type Option func(*Store)

type dbmail struct {
	ID          uuid.UUID    `bson:"id"`
	From        address      `bson:"from"`
	Recipients  []address    `bson:"recipients"`
	To          []address    `bson:"to"`
	CC          []address    `bson:"cc"`
	BCC         []address    `bson:"bcc"`
	ReplyTo     []address    `bson:"replyTo"`
	Attachments []attachment `bson:"attachments"`
	Subject     string       `bson:"subject"`
	Text        string       `bson:"text"`
	HTML        string       `bson:"html"`
	RFC         string       `bson:"rfc"`
	SendError   string       `bson:"sendError"`
	SentAt      time.Time    `bson:"sentAt"`
}

type address struct {
	Name    string `bson:"name"`
	Address string `bson:"address"`
}

type attachment struct {
	Filename    string               `bson:"filename"`
	Content     []byte               `bson:"content"`
	ContentType string               `bson:"contentType"`
	Size        int                  `bson:"size"`
	Header      textproto.MIMEHeader `bson:"header"`
}

type cursor struct {
	cur     *mongo.Cursor
	current archive.Mail
	err     error
}

// NewStore returns a mongo store. It returns an error if either client is nil
// or CreateIndexes() is used and index creation fails.
func NewStore(ctx context.Context, client *mongo.Client, opts ...Option) (*Store, error) {
	if client == nil {
		return nil, errors.New("client must not be nil")
	}
	s := Store{client: client, databaseName: "postdog", collectionName: "mails"}
	for _, opt := range opts {
		opt(&s)
	}
	s.col = s.client.Database(s.databaseName).Collection(s.collectionName)
	if s.wantIndexes {
		if err := s.createIndexes(ctx); err != nil {
			return nil, fmt.Errorf("create indexes: %w", err)
		}
	}
	return &s, nil
}

// Database returns an Option that specifies the used database.
func Database(name string) Option {
	return func(s *Store) {
		s.databaseName = name
	}
}

// Collection returns an Option that specifies the used collection.
func Collection(name string) Option {
	return func(s *Store) {
		s.collectionName = name
	}
}

// CreateIndexes returns an Option that creates the indexes for the mails collection.
func CreateIndexes(ci bool) Option {
	return func(s *Store) {
		s.wantIndexes = ci
	}
}

// WithoutAttachmentContent returns an Option that empties the attachment
// contents of mails before they are stored in the database.
func WithoutAttachmentContent(ac bool) Option {
	return func(s *Store) {
		s.withoutAttachmentContent = ac
	}
}

// Insert stores m into the database. If there's already a stored mail with the
// same ID as m, m will override the previously stored mail.
func (s *Store) Insert(ctx context.Context, m archive.Mail) error {
	attachments := make([]attachment, len(m.Attachments()))
	for i, at := range m.Attachments() {
		content := []byte{}
		if !s.withoutAttachmentContent {
			content = at.Content()
		}
		attachments[i] = attachment{
			Filename:    at.Filename(),
			Content:     content,
			Size:        at.Size(),
			ContentType: at.ContentType(),
			Header:      at.Header(),
		}
	}

	dbm := dbmail{
		ID:          m.ID(),
		From:        rmapAddress(m.From()),
		Recipients:  rmapAddresses(m.Recipients()...),
		To:          rmapAddresses(m.To()...),
		CC:          rmapAddresses(m.CC()...),
		BCC:         rmapAddresses(m.BCC()...),
		ReplyTo:     rmapAddresses(m.ReplyTo()...),
		Attachments: attachments,
		Subject:     m.Subject(),
		Text:        m.Text(),
		HTML:        m.HTML(),
		RFC:         m.RFC(),
		SendError:   m.SendError(),
		SentAt:      m.SentAt(),
	}

	if _, err := s.col.ReplaceOne(ctx, bson.M{"id": m.ID()}, dbm, options.Replace().SetUpsert(true)); err != nil {
		return fmt.Errorf("mongo: %w", err)
	}

	return nil
}

// Find fetches the mail with the given id from the database. If it can't find
// the mail, it returns archive.ErrNotFound.
func (s *Store) Find(ctx context.Context, id uuid.UUID) (archive.Mail, error) {
	res := s.col.FindOne(ctx, bson.M{"id": id})
	return decode(res)
}

// Query queries the database for mails matching the query q.
func (s *Store) Query(ctx context.Context, q query.Query) (archive.Cursor, error) {
	filter := newFilter(q)
	opts := options.Find()
	opts = withSorting(opts, q)
	opts = withPagination(opts, q)
	cur, err := s.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("mongo: %w", err)
	}
	return &cursor{cur: cur}, nil
}

// Remove deletes the mail m from the database.
func (s *Store) Remove(ctx context.Context, m archive.Mail) error {
	if _, err := s.col.DeleteOne(ctx, bson.M{"id": m.ID()}); err != nil {
		return fmt.Errorf("mongo: %w", err)
	}
	return nil
}

func (s *Store) createIndexes(ctx context.Context) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
	}
	_, err := index.CreateFromConfig(ctx, s.col.Database(), index.Config{
		s.collectionName: []mongo.IndexModel{
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "sentAt", Value: 1}}},
			{Keys: bson.D{{Key: "subject", Value: 1}}},
			{Keys: bson.D{{Key: "from.name", Value: 1}}},
			{Keys: bson.D{{Key: "from.address", Value: 1}}},
			{Keys: bson.D{{Key: "recipients.name", Value: 1}}},
			{Keys: bson.D{{Key: "recipients.address", Value: 1}}},
			{Keys: bson.D{{Key: "to.name", Value: 1}}},
			{Keys: bson.D{{Key: "to.address", Value: 1}}},
			{Keys: bson.D{{Key: "cc.name", Value: 1}}},
			{Keys: bson.D{{Key: "cc.address", Value: 1}}},
			{Keys: bson.D{{Key: "bcc.name", Value: 1}}},
			{Keys: bson.D{{Key: "bcc.address", Value: 1}}},
			{Keys: bson.D{{Key: "attachments.filename", Value: 1}}},
			{Keys: bson.D{{Key: "attachments.contentType", Value: 1}}},
			{Keys: bson.D{{Key: "attachments.content", Value: 1}}},
			{Keys: bson.D{{Key: "attachments.size", Value: 1}}},
		},
	})
	return err
}

func (addr address) netMail() mail.Address {
	return mail.Address{Name: addr.Name, Address: addr.Address}
}

func (cur *cursor) Next(ctx context.Context) bool {
	if !cur.cur.Next(ctx) {
		cur.err = cur.cur.Err()
		return false
	}

	var mail dbmail
	if cur.err = cur.cur.Decode(&mail); cur.err != nil {
		cur.current = archive.Mail{}
		return false
	}

	attachments := make([]letter.Option, len(mail.Attachments))
	for i, at := range mail.Attachments {
		attachments[i] = letter.Attach(at.Filename, at.Content, letter.AttachmentType(at.ContentType))
	}

	cur.current = archive.
		ExpandMail(letter.Write(append([]letter.Option{
			letter.FromAddress(mail.From.netMail()),
			letter.RecipientAddress(netMails(mail.Recipients...)...),
			letter.ToAddress(netMails(mail.To...)...),
			letter.CCAddress(netMails(mail.CC...)...),
			letter.BCCAddress(netMails(mail.BCC...)...),
			letter.ReplyToAddress(netMails(mail.ReplyTo...)...),
			letter.Subject(mail.Subject),
			letter.Content(mail.Text, mail.HTML),
			letter.RFC(mail.RFC),
		}, attachments...)...)).
		WithID(mail.ID).
		WithSendError(mail.SendError).
		WithSendTime(mail.SentAt)

	return true
}

func (cur *cursor) Current() archive.Mail {
	return cur.current
}

func (cur *cursor) All(ctx context.Context) ([]archive.Mail, error) {
	var mails []archive.Mail
	for cur.Next(ctx) {
		mails = append(mails, cur.current)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	if err := cur.Close(ctx); err != nil {
		return mails, err
	}
	return mails, nil
}

func (cur *cursor) Err() error {
	return cur.err
}

func (cur *cursor) Close(ctx context.Context) error {
	cur.err = cur.cur.Close(ctx)
	return cur.err
}

func decode(res *mongo.SingleResult) (archive.Mail, error) {
	var m dbmail
	if err := res.Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return archive.Mail{}, archive.ErrNotFound
		}
		return archive.Mail{}, fmt.Errorf("decode: %w", err)
	}

	var attachments []letter.Option
	for _, at := range m.Attachments {
		attachments = append(attachments, letter.Attach(
			at.Filename,
			at.Content,
			letter.AttachmentType(at.ContentType),
			letter.AttachmentSize(at.Size),
		))
	}

	opts := append([]letter.Option{
		letter.FromAddress(mapAddress(m.From)),
		letter.RecipientAddress(mapAddresses(m.Recipients...)...),
		letter.ToAddress(mapAddresses(m.To...)...),
		letter.CCAddress(mapAddresses(m.CC...)...),
		letter.BCCAddress(mapAddresses(m.BCC...)...),
		letter.Subject(m.Subject),
		letter.Text(m.Text),
		letter.HTML(m.HTML),
		letter.RFC(m.RFC),
	}, attachments...)

	return archive.
		ExpandMail(letter.Write(opts...)).
		WithID(m.ID).
		WithSendError(m.SendError).
		WithSendTime(m.SentAt), nil
}

func mapAddress(addr address) mail.Address {
	return mail.Address{Name: addr.Name, Address: addr.Address}
}

func mapAddresses(addrs ...address) []mail.Address {
	res := make([]mail.Address, len(addrs))
	for i, addr := range addrs {
		res[i] = mapAddress(addr)
	}
	return res
}

func rmapAddress(addr mail.Address) address {
	return address{Name: addr.Name, Address: addr.Address}
}

func rmapAddresses(addrs ...mail.Address) []address {
	res := make([]address, len(addrs))
	for i, addr := range addrs {
		res[i] = rmapAddress(addr)
	}
	return res
}

func newFilter(q query.Query) bson.D {
	filter := bson.D{}

	if len(q.Subjects) > 0 {
		filter = withFilter(filter, []string{"subject"}, regexInValues(q.Subjects))
	}

	if len(q.Texts) > 0 {
		filter = withFilter(filter, []string{"text"}, regexInValues(q.Texts))
	}

	if len(q.HTML) > 0 {
		filter = withFilter(filter, []string{"html"}, regexInValues(q.HTML))
	}

	if len(q.RFC) > 0 {
		filter = withFilter(filter, []string{"rfc"}, regexInValues(q.RFC))
	}

	if len(q.From) > 0 {
		filter = withAddressesFilter(filter, "from", q.From...)
	}

	if len(q.Recipients) > 0 {
		filter = withAddressesFilter(filter, "recipients", q.Recipients...)
	}

	if len(q.To) > 0 {
		filter = withAddressesFilter(filter, "to", q.To...)
	}

	if len(q.CC) > 0 {
		filter = withAddressesFilter(filter, "cc", q.CC...)
	}

	if len(q.BCC) > 0 {
		filter = withAddressesFilter(filter, "bcc", q.BCC...)
	}

	if len(q.Attachment.Filenames) > 0 {
		filter = withFilter(filter, []string{"attachments.filename"}, regexInValues(q.Attachment.Filenames))
	}

	if len(q.Attachment.ContentTypes) > 0 {
		filter = withFilter(filter, []string{"attachments.contentType"}, regexInValues(q.Attachment.ContentTypes))
	}

	if len(q.Attachment.Size.Exact) > 0 {
		filter = append(filter, bson.E{Key: "attachments.size", Value: bson.D{{Key: "$in", Value: q.Attachment.Size.Exact}}})
	}

	if len(q.Attachment.Size.Ranges) > 0 {
		or := make([]bson.D, len(q.Attachment.Size.Ranges))
		for i, rang := range q.Attachment.Size.Ranges {
			or[i] = bson.D{{Key: "attachments.size", Value: bson.D{
				{Key: "$gte", Value: rang[0]},
				{Key: "$lte", Value: rang[1]},
			}}}
		}
		filter = append(filter, bson.E{Key: "$or", Value: or})
	}

	if len(q.Attachment.Contents) > 0 {
		filter = append(filter, bson.E{Key: "attachments.content", Value: bson.D{{Key: "$in", Value: q.Attachment.Contents}}})
	}

	if len(q.SendTime.Before) > 0 {
		or := make([]bson.D, len(q.SendTime.Before))
		for i, before := range q.SendTime.Before {
			// mongo only has millisecond precision
			before = before.Round(time.Millisecond)
			or[i] = bson.D{{Key: "sentAt", Value: bson.D{{Key: "$lt", Value: before}}}}
		}
		filter = append(filter, bson.E{Key: "$or", Value: or})
	}

	if len(q.SendTime.After) > 0 {
		or := make([]bson.D, len(q.SendTime.After))
		for i, after := range q.SendTime.After {
			// mongo only has millisecond precision
			after = after.Round(time.Millisecond)
			or[i] = bson.D{{Key: "sentAt", Value: bson.D{{Key: "$gt", Value: after}}}}
		}
		filter = append(filter, bson.E{Key: "$or", Value: or})
	}

	if len(q.SendTime.Exact) > 0 {
		or := make([]bson.D, len(q.SendTime.Exact))
		for i, exact := range q.SendTime.Exact {
			// mongo only has millisecond precision
			exact = exact.Round(time.Millisecond)
			or[i] = bson.D{{Key: "sentAt", Value: exact}}
		}
		filter = append(filter, bson.E{Key: "$or", Value: or})
	}

	if len(q.SendErrors) > 0 {
		filter = append(filter, bson.E{Key: "sendError", Value: regexInValues(q.SendErrors)})
	}

	return filter
}

func withFilter(filter bson.D, keys []string, vals interface{}) bson.D {
	switch len(keys) {
	case 0:
		return filter
	case 1:
		return append(filter, bson.E{Key: keys[0], Value: vals})
	}

	or := make([]bson.D, len(keys))
	for i, key := range keys {
		or[i] = bson.D{{Key: key, Value: vals}}
	}

	return append(filter, bson.E{Key: "$or", Value: or})
}

func withAddressesFilter(filter bson.D, field string, addrs ...mail.Address) bson.D {
	names := addressNames(addrs...)
	addresses := addresses(addrs...)

	if len(names) == 0 && len(addresses) == 0 {
		return filter
	}

	if len(names) == 0 {
		return append(filter, bson.E{Key: "$or", Value: []bson.D{
			{{Key: field + ".address", Value: addresses}},
		}})
	}

	if len(addresses) == 0 {
		return append(filter, bson.E{Key: "$or", Value: []bson.D{
			{{Key: field + ".name", Value: names}},
		}})
	}

	return append(filter, bson.E{Key: "$or", Value: []bson.D{
		{{Key: field + ".name", Value: names}},
		{{Key: field + ".address", Value: addresses}},
	}})
}

func regexInValues(texts []string) bson.D {
	exprs := make([]primitive.Regex, len(texts))
	for i, text := range texts {
		exprs[i] = primitive.Regex{
			Pattern: text,
			Options: "i",
		}
	}
	return bson.D{{
		Key:   "$in",
		Value: exprs,
	}}
}

func addressNames(addrs ...mail.Address) []string {
	names := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		if addr.Name != "" {
			names = append(names, addr.Name)
		}
	}
	return names
}

func addresses(addrs ...mail.Address) []string {
	mails := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		if addr.Address != "" {
			mails = append(mails, addr.Address)
		}
	}
	return mails
}

func newSort(q query.Query) (sort bson.D) {
	if q.Sorting == query.SortAny {
		return
	}

	defer func() {
		if q.SortDirection == query.SortDesc {
			sort[0].Value = -1
			return
		}
		sort[0].Value = 1
	}()

	sort = bson.D{{}}

	switch q.Sorting {
	case query.SortSendTime:
		sort[0].Key = "sentAt"
	case query.SortSubject:
		sort[0].Key = "subject"
	}
	return
}

func netMails(addrs ...address) []mail.Address {
	res := make([]mail.Address, len(addrs))
	for i, addr := range addrs {
		res[i] = addr.netMail()
	}
	return res
}

func withSorting(opts *options.FindOptions, q query.Query) *options.FindOptions {
	if sort := newSort(q); sort != nil {
		return opts.SetSort(sort)
	}
	return opts
}

func withPagination(opts *options.FindOptions, q query.Query) *options.FindOptions {
	if q.Pagination.Page == 0 {
		return opts
	}
	return opts.
		SetSkip(int64((q.Pagination.Page - 1) * q.Pagination.PerPage)).
		SetLimit(int64(q.Pagination.PerPage))
}
