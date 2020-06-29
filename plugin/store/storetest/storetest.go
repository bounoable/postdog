// Package storetest provides testing utilities that can be used to test store implementations.
package storetest

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/query"
	"github.com/bounoable/timefn"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Insert tests the Insert() method of s.
func Insert(t *testing.T, s store.Store) {
	assert.Nil(t, s.Insert(context.Background(), store.Letter{
		Letter: letter.Write(),
		SentAt: time.Now(),
	}))
}

// Query tests the Query() method of a query repository.
func Query(t *testing.T, createRepo func(...store.Letter) query.Repository) {
	letters := []store.Letter{
		{
			ID: uuid.New(),
			Letter: letter.Write(
				letter.Subject("Letter 1"),
				letter.From("Sender 1", "sender1@example.test"),
				letter.To("Recipient 1", "to1@example.test"),
				letter.CC("CC Recipient 1", "cc1@example.test"),
				letter.BCC("BCC Recipient 1", "bcc1@example.test"),
				letter.MustAttach(bytes.NewReader([]byte{1}), "Attachment 1", letter.ContentType("text/plain")),
			),
			SentAt: timefn.StartOfDay(time.Now()).UTC(),
		},
		{
			ID: uuid.New(),
			Letter: letter.Write(
				letter.Subject("Letter 2"),
				letter.From("Sender 2", "sender2@example.test"),
				letter.To("Recipient 2", "to2@example.test"),
				letter.CC("CC Recipient 2", "cc2@example.test"),
				letter.BCC("BCC Recipient 2", "bcc2@example.test"),
				letter.MustAttach(bytes.NewReader([]byte{2, 2}), "Attachment 2", letter.ContentType("text/html")),
			),
			SentAt: timefn.StartOfDay(time.Now().AddDate(0, 0, 1)).UTC(),
		},
		{
			ID: uuid.New(),
			Letter: letter.Write(
				letter.Subject("Letter 3"),
				letter.From("Sender 3", "sender3@example.test"),
				letter.To("Recipient 3", "to3@example.test"),
				letter.CC("CC Recipient 3", "cc3@example.test"),
				letter.BCC("BCC Recipient 3", "bcc3@example.test"),
				letter.MustAttach(bytes.NewReader([]byte{3, 3, 3}), "Attachment 3", letter.ContentType("application/octet-stream")),
			),
			SentAt: timefn.StartOfDay(time.Now().AddDate(0, 0, 2)).UTC(),
		},
		{
			ID: uuid.New(),
			Letter: letter.Write(
				letter.Subject("Letter 4"),
				letter.From("Sender 4", "sender4@example.test"),
				letter.To("Recipient 4", "to4@example.test"),
				letter.CC("CC Recipient 4", "cc4@example.test"),
				letter.BCC("BCC Recipient 4", "bcc4@example.test"),
				letter.MustAttach(bytes.NewReader([]byte{4, 4, 4, 4}), "Attachment 4", letter.ContentType("application/pdf")),
			),
			SentAt: timefn.StartOfDay(time.Now().AddDate(0, 0, 3)).UTC(),
		},
	}

	cases := map[string]struct {
		query    query.Query
		expected []store.Letter
	}{
		"query all": {
			expected: letters,
		},
		"SentAt.Before": {
			query:    query.New(query.SentBefore(timefn.StartOfDay(time.Now().AddDate(0, 0, 3)))),
			expected: letters[:3],
		},
		"SentAt.After": {
			query:    query.New(query.SentAfter(timefn.StartOfDay(time.Now()))),
			expected: letters[1:],
		},
		"SentAt.InBetween": {
			query: query.New(
				query.SentInBetween(
					timefn.StartOfDay(time.Now()),
					timefn.StartOfDay(time.Now().AddDate(0, 0, 3)),
				),
			),
			expected: letters[1:3],
		},
		"SentAt.Between": {
			query: query.New(
				query.SentBetween(
					timefn.StartOfDay(time.Now()),
					timefn.StartOfDay(time.Now().AddDate(0, 0, 3)),
				),
			),
			expected: letters,
		},
		"Subject": {
			query: query.New(query.Subject("1", "4")),
			expected: []store.Letter{
				letters[0],
				letters[3],
			},
		},
		"Subject (substring)": {
			query:    query.New(query.Subject("Letter")),
			expected: letters,
		},
		"From": {
			query: query.New(query.From("Sender 2", "Sender 4")),
			expected: []store.Letter{
				letters[1],
				letters[3],
			},
		},
		"From (substring)": {
			query:    query.New(query.From("Sender")),
			expected: letters,
		},
		"To": {
			query: query.New(query.To("Recipient 1", "Recipient 3")),
			expected: []store.Letter{
				letters[0],
				letters[2],
			},
		},
		"To (substring)": {
			query:    query.New(query.To("Recipient")),
			expected: letters,
		},
		"CC": {
			query: query.New(query.CC("CC Recipient 1", "CC Recipient 2")),
			expected: []store.Letter{
				letters[0],
				letters[1],
			},
		},
		"CC (substring)": {
			query:    query.New(query.CC("CC")),
			expected: letters,
		},
		"BCC": {
			query: query.New(query.BCC("BCC Recipient 2", "BCC Recipient 3")),
			expected: []store.Letter{
				letters[1],
				letters[2],
			},
		},
		"BCC (substring)": {
			query:    query.New(query.BCC("BCC")),
			expected: letters,
		},
		"Attachment name": {
			query: query.New(
				query.AttachmentName("2"),
				query.AttachmentName("4"),
			),
			expected: []store.Letter{
				letters[1],
				letters[3],
			},
		},
		"Attachment Content-Type": {
			query: query.New(
				query.AttachmentContentType("text/html"),
				query.AttachmentContentType("application/pdf"),
			),
			expected: []store.Letter{
				letters[1],
				letters[3],
			},
		},
		"Attachment Content-Type (substring)": {
			query: query.New(
				query.AttachmentContentType("text"),
			),
			expected: []store.Letter{
				letters[0],
				letters[1],
			},
		},
		"Attachment size": {
			query: query.New(
				query.AttachmentSize(1, 3),
			),
			expected: []store.Letter{
				letters[0],
				letters[2],
			},
		},
		"Attachment size range": {
			query: query.New(
				query.AttachmentSizeRange(2, 4),
			),
			expected: []store.Letter{
				letters[1],
				letters[2],
				letters[3],
			},
		},
		"Attachment size range (multiple)": {
			query: query.New(
				query.AttachmentSizeRange(1, 2),
				query.AttachmentSizeRange(2, 3),
			),
			expected: []store.Letter{
				letters[0],
				letters[1],
				letters[2],
			},
		},
		"Sorting (SentAt asc)": {
			query: query.New(
				query.Sort(query.SortBySendDate, query.SortAsc),
			),
			expected: letters,
		},
		"Sorting (SentAt desc)": {
			query: query.New(
				query.Sort(query.SortBySendDate, query.SortDesc),
			),
			expected: []store.Letter{
				letters[3],
				letters[2],
				letters[1],
				letters[0],
			},
		},
		"Paginate": {
			query:    query.New(query.Paginate(2, 2)),
			expected: letters[2:],
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			testQuery(t, createRepo(letters...), tcase.query, tcase.expected)
		})
	}
}

func testQuery(t *testing.T, s query.Repository, q query.Query, expected []store.Letter) {
	ctx := context.Background()

	var letters []store.Letter

	cur, err := s.Query(ctx, q)
	assert.Nil(t, err)
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		letters = append(letters, cur.Current())
	}

	assert.Nil(t, cur.Err())
	assert.Equal(t, expected, letters)
}

// Get tests the Get() method of a query repository.
func Get(t *testing.T, createRepo func(...store.Letter) query.Repository) {
	getExists(t, createRepo)
	getNotExists(t, createRepo)
}

func getExists(t *testing.T, createRepo func(...store.Letter) query.Repository) {
	let := store.Letter{
		ID:     uuid.New(),
		Letter: letter.Write(letter.Subject("Hello")),
	}

	repo := createRepo(let)
	flet, err := repo.Get(context.Background(), let.ID)
	assert.Nil(t, err)
	assert.Equal(t, let, flet)
}

func getNotExists(t *testing.T, createRepo func(...store.Letter) query.Repository) {
	repo := createRepo()
	_, err := repo.Get(context.Background(), uuid.New())
	assert.NotNil(t, err)
}
