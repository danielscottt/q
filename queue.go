package q

import (
	"errors"
	"sync"
)

var (
	// Returned when queue is empty
	ErrQueueEmpty = errors.New("queue empty")

	// Returned when bound is exceeded
	ErrLimitExceeded = errors.New("limit exceeded")
)

// A FIFO queue object
type Queue struct {

	// The function which to execute on each event or job in the queue
	Handler func(interface{})

	// Bound
	Limit int64

	// Chan to deliver events to
	Event chan interface{}

	// Chan to receive q errors from
	Err chan error

	items  []interface{}
	locker *Locker
}

// Sets up and returns a Queue.  Bound is passed in.
func NewQueue(limit int64) *Queue {
	q := &Queue{
		Limit: limit,
		Event: make(chan interface{}, 0),
		Err:   make(chan error, 0),

		items:  make([]interface{}, 0),
		locker: NewLocker(),
	}
	return q
}

// Add item to queue
func (q *Queue) Push(i interface{}) error {
	q.init()
	if q.Limit == 0 || int64(q.Length()+1) <= q.Limit {
		q.items = append(q.items, i)
		return nil
	}
	return ErrLimitExceeded
}

// Pop first thing off queue
func (q *Queue) Pop() (interface{}, error) {
	var item interface{}
	if q.items == nil || len(q.items) == 0 {
		return item, ErrQueueEmpty
	}
	item = q.items[0]
	q.items = q.items[1:]
	return item, nil
}

// Return current size of queue
func (q *Queue) Length() int {
	q.init()
	return len(q.items)
}

// Return whether or not a queue is empty
func (q *Queue) Empty() bool {
	if q.items == nil {
		return true
	}
	if len(q.items) == 0 {
		return true
	}
	return false
}

// The wrapper for working the evented queue.
// Receives on the Event chan, and locks the thread doing work.
// The lock is cleared once the queue is empty, allowing the next goroutine
// which delivers an event to begin working the queue.
func (q *Queue) Work() {
	for {
		select {
		case e := <-q.Event:
			if err := q.Push(e); err == ErrLimitExceeded {
				q.Err <- ErrLimitExceeded
			} else {
				if !q.locker.Locked {
					q.locker.Lock()
					go q.handle()
				}
			}
		}
	}
}

func (q *Queue) handle() {
	item, err := q.Pop()
	if err == ErrQueueEmpty {
		q.locker.Unlock()
		return
	}
	q.Handler(item)
	q.handle()
}

func (q *Queue) init() {
	if q.items == nil {
		q.items = make([]interface{}, 0)
	}
}

// A non-blocking mutex. Allows for check-if-locked without waiting to get lock.
// The outside mutex is used like a traditional wait-for mutex.
// The inside mutex is used to make an atomic change to the Locked property.
type Locker struct {

	// Property which is flipped to show if a lock is currently held.
	Locked bool

	outside *sync.Mutex
	inside  *sync.Mutex
}

// Sets up private mutexes for Locker and returns its pointer
func NewLocker() *Locker {
	l := &Locker{
		inside:  new(sync.Mutex),
		outside: new(sync.Mutex),
	}
	return l
}

// Lock the mutex, block if locked.
func (l *Locker) Lock() {
	l.inside.Lock()
	l.outside.Lock()
	l.Locked = true
	l.inside.Unlock()
}

// Unlock the mutex
func (l *Locker) Unlock() {
	l.inside.Lock()
	l.outside.Unlock()
	l.Locked = false
	l.inside.Unlock()
}
