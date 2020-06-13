package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/bounoable/postdog/plugins/store/query"
	"github.com/bounoable/postdog/plugins/store/query/mock_query"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
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
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			expectedCursor := mock_query.NewMockCursor(ctrl)

			repo := mock_query.NewMockRepository(ctrl)
			repo.EXPECT().
				Query(context.Background(), tcase.expected).
				DoAndReturn(func(ctx context.Context, q query.Query) (query.Cursor, error) {
					return expectedCursor, nil
				})

			cur, err := query.Run(
				context.Background(),
				repo,
				tcase.options...,
			)

			assert.Nil(t, err)
			assert.Same(t, expectedCursor, cur)
		})
	}
}
