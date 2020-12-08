package cursor_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/bounoable/postdog/plugin/archive/cursor"
	"github.com/bounoable/postdog/plugin/archive/query"
	"github.com/stretchr/testify/assert"
)

var _ query.Cursor = (*cursor.Cursor)(nil)

var mockMails = []postdog.Mail{
	letter.Write(letter.From("Bob Belcher", "bob@example.com")),
	letter.Write(letter.From("Linda Belcher", "linda@example.com")),
	letter.Write(letter.From("Tina Belcher", "tina@example.com")),
}

func TestCursor_All_new(t *testing.T) {
	cur := cursor.New(mockMails...)
	mails, err := cur.All(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, mockMails, mails)
}

func TestCursor_All_push(t *testing.T) {
	cur := cursor.New()
	err := cur.Push(mockMails...)
	assert.Nil(t, err)
	mails, err := cur.All(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, mockMails, mails)
}

func TestCursor_All_newAndPush(t *testing.T) {
	cur := cursor.New(mockMails...)
	err := cur.Push(mockMails...)
	assert.Nil(t, err)
	mails, err := cur.All(context.Background())
	assert.Equal(t, append(mockMails, mockMails...), mails)
}

func TestCursor_All_closed(t *testing.T) {
	cur := cursor.New(mockMails...)
	err := cur.Close(context.Background())
	assert.Nil(t, err)
	mails, err := cur.All(context.Background())
	assert.True(t, errors.Is(err, cursor.ErrClosed))
	assert.Nil(t, mails)
}

func TestCursor_All_successive(t *testing.T) {
	cur := cursor.New(mockMails...)
	_, err := cur.All(context.Background())
	assert.Nil(t, err)
	_, err = cur.All(context.Background())
	assert.True(t, errors.Is(err, cursor.ErrClosed))
}

func TestCursor_All_remaining(t *testing.T) {
	cur := cursor.New(mockMails...)
	ok := cur.Next(context.Background())
	assert.True(t, ok)

	mails, err := cur.All(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, mockMails[1:], mails)
}

func TestCursor_Next(t *testing.T) {
	cur := cursor.New(mockMails...)

	for i := 0; i < len(mockMails); i++ {
		ok := cur.Next(context.Background())
		assert.True(t, ok)
		assert.Nil(t, cur.Err())
		assert.Equal(t, mockMails[i], cur.Current())
	}

	ok := cur.Next(context.Background())
	assert.False(t, ok)
	assert.Nil(t, cur.Err())
}

func TestCursor_Next_concurrent(t *testing.T) {
	cur := cursor.New(mockMails...)

	var wg sync.WaitGroup
	wg.Add(len(mockMails))

	for i := 0; i < len(mockMails); i++ {
		go func() {
			defer wg.Done()
			ok := cur.Next(context.Background())
			assert.True(t, ok)
		}()
	}

	wg.Wait()
	assert.Nil(t, cur.Err())
}

func TestCursor_Close(t *testing.T) {
	cur := cursor.New(mockMails...)
	err := cur.Close(context.Background())
	assert.Nil(t, err)
}

func TestCursor_Close_successive(t *testing.T) {
	cur := cursor.New(mockMails...)
	err := cur.Close(context.Background())
	assert.Nil(t, err)
	err = cur.Close(context.Background())
	assert.True(t, errors.Is(err, cursor.ErrClosed))
}

func TestCursor_Push(t *testing.T) {
	cur := cursor.New()
	err := cur.Push(mockMails...)
	assert.Nil(t, err)
}

func TestCursor_Push_closed(t *testing.T) {
	cur := cursor.New()
	err := cur.Push(mockMails[0])
	assert.Nil(t, err)
	err = cur.Close(context.Background())
	assert.Nil(t, err)
	err = cur.Push(mockMails[1])
	assert.True(t, errors.Is(err, cursor.ErrClosed))
}

func TestCursor_Push_concurrent(t *testing.T) {
	cur := cursor.New()

	var wg sync.WaitGroup
	wg.Add(len(mockMails))

	for i := 0; i < len(mockMails); i++ {
		go func(i int) {
			defer wg.Done()
			err := cur.Push(mockMails[i])
			assert.Nil(t, err)
		}(i)
	}

	wg.Wait()
	assert.Nil(t, cur.Err())
}
