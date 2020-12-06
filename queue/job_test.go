package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/letter"
	"github.com/stretchr/testify/assert"
)

func TestJob_Mail(t *testing.T) {
	m := letter.Write(letter.From("Bob Belcher", "bob@example.com"))
	j := &Job{mail: m}
	assert.Equal(t, m, j.Mail())
}

func TestJob_SendOptions(t *testing.T) {
	j := &Job{sendOptions: []postdog.SendOption{
		postdog.Use("a"),
		postdog.Use("b"),
	}}
	assert.Len(t, j.SendOptions(), 2)
}

func TestJob_Context(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKey("key"), "val")
	j := &Job{ctx: ctx}
	assert.Equal(t, ctx, j.Context())
}

func TestJob_DispatchedAt(t *testing.T) {
	ti := time.Now().Add(time.Minute * 20)
	j := &Job{dispatchedAt: ti}
	assert.Equal(t, ti, j.DispatchedAt())
}

func TestJob_Runtime_active(t *testing.T) {
	now := time.Now()
	sod := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dur := now.Sub(sod)
	j := &Job{dispatchedAt: sod}
	assert.InDelta(t, dur, j.Runtime(), float64(time.Millisecond))
}

func TestJob_Runtime_completed(t *testing.T) {
	now := time.Now()
	sod := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	f := sod.Add(time.Second * 3)
	j := &Job{dispatchedAt: sod, finishedAt: f}
	assert.Equal(t, time.Second*3, j.Runtime())
}

func TestJob_Done_notDone(t *testing.T) {
	j := &Job{done: make(chan struct{})}

	select {
	case _, ok := <-j.Done():
		if !ok {
			t.Fatal("j.Done() should not be closed")
		}
		t.Fatal("j.Done() should not contain any values")
	default:
	}
}

func TestJob_Done_done(t *testing.T) {
	j := &Job{done: make(chan struct{})}
	j.ctx, j.cancel = context.WithCancel(context.Background())
	j.finish(nil)

	select {
	case _, ok := <-j.Done():
		assert.False(t, ok, "j.Done() should be closed")
	default:
	}
}

func TestJob_Done_successive(t *testing.T) {
	j := &Job{done: make(chan struct{})}
	j.ctx, j.cancel = context.WithCancel(context.Background())
	d1 := j.Done()
	d2 := j.Done()
	d3 := j.Done()
	assert.Equal(t, d1, d2)
	assert.Equal(t, d2, d3)
	j.finish(nil)
	d4 := j.Done()
	d5 := j.Done()
	assert.Equal(t, d3, d4)
	assert.Equal(t, d4, d5)
}

func TestJob_Err(t *testing.T) {
	err := errors.New("mock error")
	j := &Job{err: err}
	assert.Equal(t, err, j.Err())
}

type ctxKey string
