package should

import (
	"reflect"
)

// BeClosed determines if actual is a closed channel.
func BeClosed(actual interface{}, expected ...interface{}) string {
	if actual == nil {
		return "actual is nil"
	}

	if len(expected) > 0 {
		return "should.BeClosed() doesn't accept 'expected' parameters"
	}

	t := reflect.TypeOf(actual)
	v := reflect.ValueOf(actual)

	if t.Kind() != reflect.Chan {
		return "actual is not a channel"
	}

	if t.ChanDir() == reflect.SendDir {
		return "cannot determine if actual is closed, because it is a send-only channel"
	}

	chosen, _, open := reflect.Select([]reflect.SelectCase{
		{Dir: reflect.SelectRecv, Chan: v},
		{Dir: reflect.SelectDefault},
	})

	var closed bool
	if chosen == 0 {
		closed = !open
	}

	if !closed {
		return "actual is not closed"
	}

	return ""
}

// BeOpen determines if actual is an open channel.
func BeOpen(actual interface{}, expected ...interface{}) string {
	if actual == nil {
		return "actual is nil"
	}

	if len(expected) > 0 {
		return "should.BeOpen() doesn't accept 'expected' parameters"
	}

	t := reflect.TypeOf(actual)
	v := reflect.ValueOf(actual)

	if t.Kind() != reflect.Chan {
		return "actual is not a channel"
	}

	if t.ChanDir() == reflect.SendDir {
		return "cannot determine if actual is open, because it is a send-only channel"
	}

	chosen, _, open := reflect.Select([]reflect.SelectCase{
		{Dir: reflect.SelectRecv, Chan: v},
		{Dir: reflect.SelectDefault},
	})

	var closed bool
	if chosen == 0 {
		closed = !open
	}

	if closed {
		return "actual is closed"
	}

	return ""
}
