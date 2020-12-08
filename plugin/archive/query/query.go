package query

import (
	"context"
	"net/mail"

	"github.com/bounoable/postdog"
)

const (
	// SortAny is the default Sorting and sorts with an undefined / unpredictable behavior.
	SortAny = Sorting(iota)
	// SortSentAt sorts by the send time of the mails.
	SortSentAt
)

const (
	// SortAsc sorts in ascending order.
	SortAsc = SortDirection(iota)
	// SortDesc sorts in descending order.
	SortDesc
)

// Query provides filters, sorting & pagination for querying mails.
type Query struct {
	From          []mail.Address
	To            []mail.Address
	CC            []mail.Address
	BCC           []mail.Address
	Subjects      []string
	Attachment    AttachmentFilter
	Sorting       Sorting
	SortDirection SortDirection
	Pagination    Pagination
}

// AttachmentFilter is the query filter for attachments.
type AttachmentFilter struct {
	Filenames    []string
	ContentTypes []string
	Size         AttachmentSizeFilter
}

// AttachmentSizeFilter is the query filter for attachment file sizes.
type AttachmentSizeFilter struct {
	Exact  []int
	Ranges [][2]int
}

// Sorting is a sorting.
type Sorting int

// SortDirection is a sorting direction.
type SortDirection int

// Pagination is a pagination option.
type Pagination struct {
	Page    int
	PerPage int
}

// Option is a Query option.
type Option func(*Query)

// Cursor is a cursor postdog.Mails.
type Cursor interface {
	// Next advances the cursor to the next mail.
	// Implementations should return true if the next call to Current() would
	// return a non-nil postdog.Mail, or false if the Cursor reached the end or
	// if Next() failed because of an error. In the latter case, Err() should
	// return that error.
	Next(context.Context) bool

	// Current returns the current postdog.Mail.
	Current() postdog.Mail

	// All returns the remaining postdog.Mails from the Cursor and calls cur.Close(ctx) afterwards.
	All(context.Context) ([]postdog.Mail, error)

	// Err returns the current error that occurred during a previous Next() call.
	Err() error

	// Close closes the Cursor. Users must call Close() after using the Cursor if they don't call All().
	Close(context.Context) error
}

// New builds a Query using the provided opts.
func New(opts ...Option) Query {
	var q Query
	for _, opt := range opts {
		opt(&q)
	}
	return q
}

// From returns an Option that adds a `From` filter to a Query.
func From(addr ...mail.Address) Option {
	return func(q *Query) {
		q.From = append(q.From, addr...)
	}
}

// To returns an Option that adds a `To` filter to a Query.
func To(addr ...mail.Address) Option {
	return func(q *Query) {
		q.To = append(q.To, addr...)
	}
}

// CC returns an Option that adds a `Cc` filter to a Query.
func CC(addr ...mail.Address) Option {
	return func(q *Query) {
		q.CC = append(q.CC, addr...)
	}
}

// BCC returns an Option that adds a `Bcc` filter to a Query.
func BCC(addr ...mail.Address) Option {
	return func(q *Query) {
		q.BCC = append(q.BCC, addr...)
	}
}

// Subject returns an Option that adds a `Subject` filter to a Query.
func Subject(subjects ...string) Option {
	return func(q *Query) {
		q.Subjects = append(q.Subjects, subjects...)
	}
}

// AttachmentFilename returns an Option that adds an attachment filter to a Query.
// It filters attachments by their filename.
func AttachmentFilename(names ...string) Option {
	return func(q *Query) {
		q.Attachment.Filenames = append(q.Attachment.Filenames, names...)
	}
}

// AttachmentSize returns an Option that add an attachment filter to a Query.
// It filters attachments by their file size.
func AttachmentSize(sizes ...int) Option {
	return func(q *Query) {
		q.Attachment.Size.Exact = append(q.Attachment.Size.Exact, sizes...)
	}
}

// AttachmentSizeRange returns an Option that adds an attachment filter to a Query.
// It filters attachments by their file size, where the attachment's file size
// must be in the inclusive range (min, max).
func AttachmentSizeRange(min, max int) Option {
	return func(q *Query) {
		q.Attachment.Size.Ranges = append(q.Attachment.Size.Ranges, [2]int{min, max})
	}
}

// AttachmentContentType returns an Option that adds an attachment filter to a Query.
// It filters attachments by their MIME `Content-Type`.
func AttachmentContentType(cts ...string) Option {
	return func(q *Query) {
		q.Attachment.ContentTypes = append(q.Attachment.ContentTypes, cts...)
	}
}

// Sort returns an Option that configures the Sorting of a Query.
func Sort(by Sorting, dir SortDirection) Option {
	return func(q *Query) {
		q.Sorting = by
		q.SortDirection = dir
	}
}

// Paginate returns an Option that configures the Pagination of a Query.
func Paginate(page, perPage int) Option {
	return func(q *Query) {
		q.Pagination.Page = page
		q.Pagination.PerPage = perPage
	}
}
