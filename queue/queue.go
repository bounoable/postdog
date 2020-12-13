package queue

//go:generate mockgen -source=queue.go -destination=./mocks/queue.go

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bounoable/postdog"
	"github.com/bounoable/postdog/queue/dispatch"
	"github.com/bounoable/postdog/send"
)

var (
	// ErrStarted means the queue has already been started.
	ErrStarted = errors.New("queue already started")
	// ErrNotStarted means the queue has not been started yet.
	ErrNotStarted = errors.New("queue not started")
	// ErrCanceled means a job has been canceled by either calling job.Cancel() or by canceling it's context.
	ErrCanceled = errors.New("job canceled")
	// ErrFinished means a job has already been finished and can therefore not be canceled.
	ErrFinished = errors.New("job already finished")
)

// Queue is the mailer queue.
type Queue struct {
	mailer     Mailer
	bufferSize int
	workers    int

	mux  sync.Mutex
	jobs chan *Job
	done chan struct{}
}

// Mailer is an interface for *postdog.Dog.
type Mailer interface {
	Send(context.Context, postdog.Mail, ...send.Option) error
}

// Job is a queue job.
type Job struct {
	ctx    context.Context
	cancel context.CancelFunc

	mail         postdog.Mail
	sendOptions  []send.Option
	dispatchedAt time.Time
	finishedAt   time.Time
	done         chan struct{}

	mux sync.RWMutex
	err error
}

// Option is a queue option.
type Option func(*Queue)

// New returns a new *Queue that sends mails through the Mailer m.
func New(m Mailer, opts ...Option) *Queue {
	q := &Queue{mailer: m, workers: 1}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

// Buffer returns an Option that sets the buffer size of a *Queue.
func Buffer(s int) Option {
	return func(q *Queue) {
		q.bufferSize = s
	}
}

// Workers returns an Option that sets the worker count of a *Queue.
func Workers(w int) Option {
	return func(q *Queue) {
		q.workers = w
	}
}

// Start the queue workers in a new goroutine.
func (q *Queue) Start() error {
	if q.started() {
		return ErrStarted
	}
	q.jobs = make(chan *Job, q.bufferSize)
	q.done = make(chan struct{})
	go q.run()
	return nil
}

// Started determines if the queue has been started.
func (q *Queue) Started() bool {
	return q.started()
}

func (q *Queue) started() bool {
	q.mux.Lock()
	defer q.mux.Unlock()
	return q.jobs != nil
}

func (q *Queue) run() {
	var wg sync.WaitGroup
	wg.Add(q.workers)
	go func() {
		wg.Wait()
		close(q.done)
	}()

	for i := 0; i < q.workers; i++ {
		go func() {
			defer wg.Done()
			for job := range q.jobs {
				err := q.mailer.Send(job.ctx, job.mail, job.sendOptions...)
				job.finish(err)
			}
		}()
	}
}

// Stop the queue. If the queue has not been started yet, Stop() returns
// ErrNotStarted. If ctx is canceled before the remaining jobs have been
// processed, Stop() returns ctx.Err().
func (q *Queue) Stop(ctx context.Context) error {
	if !q.started() {
		return ErrNotStarted
	}

	q.mux.Lock()
	defer q.mux.Unlock()

	close(q.jobs)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-q.done:
		q.jobs = nil
		q.done = nil
		return nil
	}
}

// Dispatch adds m to q and returns the queue *Job, or an error if the dispatch failed.
func (q *Queue) Dispatch(ctx context.Context, m postdog.Mail, opts ...dispatch.Option) (*Job, error) {
	return q.DispatchConfig(ctx, m, dispatch.Configure(opts...))
}

// DispatchConfig does the same as Dispatch() but accepts a dispatch.Config instead if dispatch.Options.
func (q *Queue) DispatchConfig(ctx context.Context, m postdog.Mail, cfg dispatch.Config) (*Job, error) {
	if !q.started() {
		return nil, ErrNotStarted
	}

	var cancel context.CancelFunc

	if cfg.Timeout == 0 {
		ctx, cancel = context.WithCancel(ctx)
	} else {
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
	}

	j := &Job{
		ctx:         ctx,
		cancel:      cancel,
		mail:        m,
		sendOptions: cfg.SendOptions,
		done:        make(chan struct{}),
	}

	select {
	case <-ctx.Done():
		cancel()
		return nil, ctx.Err()
	case q.jobs <- j:
		j.dispatchedAt = time.Now()
		return j, nil
	}
}

// Context returns the job's context that has been passed to the (*Queue).Dispatch() method.
func (j *Job) Context() context.Context {
	return j.ctx
}

// Mail returns the queued mail.
func (j *Job) Mail() postdog.Mail {
	return j.mail
}

// SendOptions returns the job's send.Options.
func (j *Job) SendOptions() []send.Option {
	return j.sendOptions
}

// DispatchedAt returns the time at which j was dispatched.
func (j *Job) DispatchedAt() time.Time {
	return j.dispatchedAt
}

// Err returns a non-nil error if the job failed.
func (j *Job) Err() error {
	j.mux.RLock()
	defer j.mux.RUnlock()
	return j.err
}

// Runtime returns the current runtime time.Now().Sub(j.DispatchedAt()) if the
// job isn't done yet. Otherwise it returns the total duration between
// j.DispatchedAt() and the time the job has completed.
func (j *Job) Runtime() time.Duration {
	j.mux.RLock()
	defer j.mux.RUnlock()
	if j.finishedAt.IsZero() {
		return time.Now().Sub(j.dispatchedAt)
	}
	return j.finishedAt.Sub(j.dispatchedAt)
}

// Done returns a channel that's closed when the job is done.
// After the returned channel has been closed, j.Err() returns either nil or
// an error if the job failed.
func (j *Job) Done() <-chan struct{} {
	return j.done
}

// Cancel the job. If the job is already finished, Cancel() returns ErrFinished.
func (j *Job) Cancel(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-j.Done():
		return ErrFinished
	default:
	}

	j.cancel()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-j.Done():
		return nil
	}
}

func (j *Job) finish(err error) {
	defer close(j.done)
	defer j.cancel()
	j.mux.Lock()
	defer j.mux.Unlock()
	j.finishedAt = time.Now()
	if err == nil {
		return
	}

	if errors.Is(err, context.Canceled) {
		err = fmt.Errorf("%w: %v", ErrCanceled, err)
	}

	j.err = err
}
