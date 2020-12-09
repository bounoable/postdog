package query

import (
	"net/mail"
)

const (
	// SortAny is the default Sorting and sorts with an undefined / unpredictable behavior.
	SortAny = Sorting(iota)
	// SortSendTime sorts mails by their send time.
	SortSendTime
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
	Recipients    []mail.Address
	Subjects      []string
	Texts         []string
	HTML          []string
	RFC           []string
	SendErrors    []string
	Attachment    AttachmentFilter
	Sorting       Sorting
	SortDirection SortDirection
	Pagination    Pagination
}

// AttachmentFilter is the query filter for attachments.
type AttachmentFilter struct {
	Filenames    []string
	ContentTypes []string
	Contents     [][]byte
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

// Recipient returns an Option that adds a `Recipient` filter to a Query.
func Recipient(rcpts ...mail.Address) Option {
	return func(q *Query) {
		q.Recipients = append(q.Recipients, rcpts...)
	}
}

// Subject returns an Option that adds a `Subject` filter to a Query.
func Subject(subjects ...string) Option {
	return func(q *Query) {
		q.Subjects = append(q.Subjects, subjects...)
	}
}

// Text returns an Option that adds a `Text` filter to a Query.
func Text(texts ...string) Option {
	return func(q *Query) {
		q.Texts = append(q.Texts, texts...)
	}
}

// HTML returns an Option that adds an `HTML` filter to a Query.
func HTML(html ...string) Option {
	return func(q *Query) {
		q.HTML = append(q.HTML, html...)
	}
}

// RFC returns an Option that adds an `RFC` filter to a Query.
func RFC(rfc ...string) Option {
	return func(q *Query) {
		q.RFC = append(q.RFC, rfc...)
	}
}

// SendError returns an Option that adds a `send error` filter to a Query.
func SendError(errs ...string) Option {
	return func(q *Query) {
		q.SendErrors = append(q.SendErrors, errs...)
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

// AttachmentContent returns an Option that adds an attachment filter to a Query.
// It filters attachments by their actual file contents.
func AttachmentContent(content ...[]byte) Option {
	return func(q *Query) {
		q.Attachment.Contents = append(q.Attachment.Contents, content...)
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
