package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugins/store"
	"github.com/bounoable/postdog/plugins/store/query"
	"github.com/bounoable/timefn"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Insert ...
func Insert(t *testing.T, s store.Store) {
	assert.Nil(t, s.Insert(context.Background(), store.Letter{
		Letter: letter.Write(),
		SentAt: time.Now(),
	}))
}

// Query ...
func Query(t *testing.T, createRepo func(...store.Letter) query.Repository) {
	letters := []store.Letter{
		{
			Letter: letter.Write(letter.Subject("Letter 1")),
			SentAt: timefn.StartOfDay(time.Now()),
		},
		{
			Letter: letter.Write(letter.Subject("Letter 2")),
			SentAt: timefn.StartOfDay(time.Now().AddDate(0, 0, 1)),
		},
		{
			Letter: letter.Write(letter.Subject("Letter 3")),
			SentAt: timefn.StartOfDay(time.Now().AddDate(0, 0, 2)),
		},
		{
			Letter: letter.Write(letter.Subject("Letter 4")),
			SentAt: timefn.StartOfDay(time.Now().AddDate(0, 0, 3)),
		},
	}

	cases := map[string]struct {
		query    query.Query
		expected []store.Letter
	}{
		"query all": {
			expected: letters,
		},
		"filter SentAt (Before)": {
			query: query.Query{
				SentAt: query.SentAtFilter{
					Before: timefn.StartOfDay(time.Now().AddDate(0, 0, 3)),
				},
			},
			expected: letters[:3],
		},
		"filter SentAt (After)": {
			query: query.Query{
				SentAt: query.SentAtFilter{
					After: timefn.StartOfDay(time.Now()),
				},
			},
			expected: letters[1:],
		},
		"filter SentAt (Between)": {
			query: query.Query{
				SentAt: query.SentAtFilter{
					After:  timefn.StartOfDay(time.Now()),
					Before: timefn.StartOfDay(time.Now().AddDate(0, 0, 3)),
				},
			},
			expected: letters[1:3],
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
