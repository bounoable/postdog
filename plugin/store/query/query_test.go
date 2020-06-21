package query_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/store"
	"github.com/bounoable/postdog/plugin/store/query"
	"github.com/bounoable/postdog/plugin/store/query/mock_query"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cases := map[string]struct {
		options  []query.Option
		expected query.Query
	}{
		"SentAfter": {
			options: []query.Option{
				query.SentAfter(time.Date(2020, time.June, 1, 15, 30, 0, 0, time.UTC)),
			},
			expected: query.Query{
				SentAt: query.SentAtFilter{
					After: time.Date(2020, time.June, 1, 15, 30, 0, 0, time.UTC),
				},
			},
		},
		"SentBefore": {
			options: []query.Option{
				query.SentBefore(time.Date(2020, time.June, 1, 15, 30, 0, 0, time.UTC)),
			},
			expected: query.Query{
				SentAt: query.SentAtFilter{
					Before: time.Date(2020, time.June, 1, 15, 30, 0, 0, time.UTC),
				},
			},
		},
		"SentBetween": {
			options: []query.Option{
				query.SentBetween(
					time.Date(2020, time.June, 1, 15, 30, 0, 0, time.UTC),
					time.Date(2020, time.July, 1, 15, 30, 0, 0, time.UTC),
				),
			},
			expected: query.Query{
				SentAt: query.SentAtFilter{
					After:  time.Date(2020, time.June, 1, 15, 30, 0, 0, time.UTC).Add(-time.Nanosecond),
					Before: time.Date(2020, time.July, 1, 15, 30, 0, 0, time.UTC).Add(time.Nanosecond),
				},
			},
		},
		"SentInBetween": {
			options: []query.Option{
				query.SentInBetween(
					time.Date(2020, time.June, 1, 15, 30, 0, 0, time.UTC),
					time.Date(2020, time.July, 1, 15, 30, 0, 0, time.UTC),
				),
			},
			expected: query.Query{
				SentAt: query.SentAtFilter{
					After:  time.Date(2020, time.June, 1, 15, 30, 0, 0, time.UTC),
					Before: time.Date(2020, time.July, 1, 15, 30, 0, 0, time.UTC),
				},
			},
		},
		"Subjects": {
			options: []query.Option{
				query.Subject("subject 1"),
				query.Subject("subject 2", "subject 3"),
			},
			expected: query.Query{
				Subjects: []string{"subject 1", "subject 2", "subject 3"},
			},
		},
		"From": {
			options: []query.Option{
				query.From("sender 1"),
				query.From("sender 2", "sender 3"),
			},
			expected: query.Query{
				From: []string{"sender 1", "sender 2", "sender 3"},
			},
		},
		"To": {
			options: []query.Option{
				query.To("to 1"),
				query.To("to 2", "to 3"),
			},
			expected: query.Query{
				To: []string{"to 1", "to 2", "to 3"},
			},
		},
		"CC": {
			options: []query.Option{
				query.CC("cc 1"),
				query.CC("cc 2", "cc 3"),
			},
			expected: query.Query{
				CC: []string{"cc 1", "cc 2", "cc 3"},
			},
		},
		"BCC": {
			options: []query.Option{
				query.BCC("bcc 1"),
				query.BCC("bcc 2", "bcc 3"),
			},
			expected: query.Query{
				BCC: []string{"bcc 1", "bcc 2", "bcc 3"},
			},
		},
		"Attachments": {
			options: []query.Option{
				query.AttachmentName("attachment 1"),
				query.AttachmentName("attachment 2"),
				query.AttachmentContentType("text/plain"),
				query.AttachmentContentType("text/html"),
				query.AttachmentSize(27838),
				query.AttachmentSize(174858),
				query.AttachmentSizeRange(0, 1500),
				query.AttachmentSizeRange(400, 2000),
				query.AttachmentSizeRange(3000, 1000),
			},
			expected: query.Query{
				Attachment: query.AttachmentFilter{
					Names:        []string{"attachment 1", "attachment 2"},
					ContentTypes: []string{"text/plain", "text/html"},
					Size: query.AttachmentSizeFilter{
						Exact: []int{27838, 174858},
						Ranges: [][2]int{
							{0, 1500},
							{400, 2000},
							{1000, 3000},
						},
					},
				},
			},
		},
		"Sorting (SentAt asc)": {
			options: []query.Option{
				query.Sort(query.SortBySendDate, query.SortAsc),
			},
			expected: query.Query{
				Sort: query.SortConfig{
					SortBy: query.SortBySendDate,
					Dir:    query.SortAsc,
				},
			},
		},
		"Sorting (SentAt desc)": {
			options: []query.Option{
				query.Sort(query.SortBySendDate, query.SortDesc),
			},
			expected: query.Query{
				Sort: query.SortConfig{
					SortBy: query.SortBySendDate,
					Dir:    query.SortDesc,
				},
			},
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tcase.expected, query.New(tcase.options...))
		})
	}
}

func TestRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedCursor := mock_query.NewMockCursor(ctrl)

	opts := []query.Option{
		query.SentBefore(time.Now()),
		query.SentAfter(time.Now().Add(-time.Hour * 10)),
	}

	repo := mock_query.NewMockRepository(ctrl)
	repo.EXPECT().
		Query(context.Background(), query.New(opts...)).
		DoAndReturn(func(ctx context.Context, q query.Query) (query.Cursor, error) {
			return expectedCursor, nil
		})

	cur, err := query.Run(
		context.Background(),
		repo,
		opts...,
	)

	assert.Nil(t, err)
	assert.Equal(t, expectedCursor, cur)
}

func TestFind(t *testing.T) {
	ctx := context.Background()

	cases := map[string]struct {
		repoError   error
		assertError func(*testing.T, uuid.UUID, error, error)
	}{
		"found": {
			assertError: func(t *testing.T, _ uuid.UUID, repoError, findError error) {
				assert.Nil(t, repoError)
				assert.Nil(t, findError)
			},
		},
		"not found": {
			repoError: errors.New("not found"),
			assertError: func(t *testing.T, id uuid.UUID, repoError, findError error) {
				var notFoundError query.LetterNotFoundError
				assert.True(t, errors.As(findError, &notFoundError))
				assert.Equal(t, id, notFoundError.ID)
				assert.Equal(t, repoError, notFoundError.Err)
				assert.Equal(t, repoError, notFoundError.Unwrap())
			},
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			let := store.Letter{
				ID:     uuid.New(),
				Letter: letter.Write(letter.Subject(fmt.Sprintf("Test: %s", name))),
			}

			repo := mock_query.NewMockRepository(ctrl)

			if tcase.repoError != nil {
				repo.EXPECT().Get(ctx, let.ID).Return(store.Letter{}, tcase.repoError)
			} else {
				repo.EXPECT().Get(ctx, let.ID).Return(let, nil)
			}

			flet, err := query.Find(ctx, repo, let.ID)

			if tcase.repoError == nil {
				assert.Equal(t, let, flet)
			}

			tcase.assertError(t, let.ID, tcase.repoError, err)
		})
	}
}
