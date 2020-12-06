package should_test

import (
	"testing"

	"github.com/bounoable/postdog/internal/testing/should"
	"github.com/stretchr/testify/assert"
)

func TestBeClosed(t *testing.T) {
	tests := []struct {
		name     string
		actual   interface{}
		expected []interface{}
		want     string
	}{
		{
			name: "actual=nil",
			want: "actual is nil",
		},
		{
			name:   "actual is send-only",
			actual: make(chan<- struct{}),
			want:   "cannot determine if actual is closed, because it is a send-only channel",
		},
		{
			name:     "len(expected) != 0",
			actual:   make(chan struct{}),
			expected: []interface{}{1},
			want:     "should.BeClosed() doesn't accept 'expected' parameters",
		},
		{
			name:   "actual is not closed",
			actual: make(chan struct{}),
			want:   "actual is not closed",
		},
		{
			name:   "actual is closed",
			actual: closedChan(),
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, should.BeClosed(tt.actual, tt.expected...))
		})
	}
}

func TestBeOpen(t *testing.T) {
	tests := []struct {
		name     string
		actual   interface{}
		expected []interface{}
		want     string
	}{
		{
			name: "actual=nil",
			want: "actual is nil",
		},
		{
			name:   "actual is send-only",
			actual: make(chan<- struct{}),
			want:   "cannot determine if actual is open, because it is a send-only channel",
		},
		{
			name:     "len(expected) != 0",
			actual:   make(chan struct{}),
			expected: []interface{}{1},
			want:     "should.BeOpen() doesn't accept 'expected' parameters",
		},
		{
			name:   "actual is closed",
			actual: closedChan(),
			want:   "actual is closed",
		},
		{
			name:   "actual is not closed",
			actual: make(chan struct{}),
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, should.BeOpen(tt.actual, tt.expected...))
		})
	}
}

func closedChan() chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
