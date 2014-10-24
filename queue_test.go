package q

import (
	"reflect"
	"testing"
	"time"
)

func TestGetNewQueue(t *testing.T) {

	queue := NewQueue(0)
	queue2 := new(Queue)

	qt := reflect.TypeOf(queue)
	qt2 := reflect.TypeOf(queue2)

	if qt != qt2 {
		t.Fatalf("Should return a pointer to a Queue")
	}
}

func TestSetBound(t *testing.T) {

	queue := NewQueue(5)

	if queue.Limit != 5 {
		t.Fatalf("Should correctly set the bound")
	}

}

func TestAddItem(t *testing.T) {

	queue := NewQueue(0)

	queue.Push(1)

	if queue.Length() != 1 {
		t.Fatalf("Should add an item to the queue")
	}

}

func TestDontExceedLimit(t *testing.T) {
	queue := NewQueue(1)

	queue.Push(1)
	err := queue.Push(2)

	if err != ErrLimitExceeded {
		t.Fatalf("Should not add to queue if bound is reached")
	}

}

func TestGetFirstItem(t *testing.T) {
	queue := NewQueue(2)

	queue.Push(1)
	queue.Push(2)

	i, _ := queue.Pop()

	if i != 1 {
		t.Fatalf("Should return the first-in item")
	}

}

func TestReturnQueueEmpty(t *testing.T) {
	queue := NewQueue(1)

	queue.Push(1)
	queue.Pop()

	_, err := queue.Pop()

	if err != ErrQueueEmpty {
		t.Fatalf("Should return ErrQueueEmpty if no items remain in queue")
	}

}

func TestLength(t *testing.T) {
	queue := NewQueue(1)

	queue.Push(1)

	if queue.Length() != 1 {
		t.Fatalf("Should return the correct size of the queue")
	}

}

func TestEmpty(t *testing.T) {
	queue := NewQueue(1)

	if !queue.Empty() {
		t.Fatalf("Should return true if queue is empty")
	}
}

func TestNotEmpty(t *testing.T) {
	queue := NewQueue(1)

	queue.Push(1)

	if queue.Empty() {
		t.Fatalf("Should return false if queue is not empty")
	}
}

func TestAddWork(t *testing.T) {

	queue := NewQueue(2)

	worked := false

	queue.Handler = func(event interface{}) {
		e := event.(*bool)
		*e = true
	}

	go queue.Work()

	queue.Event <- &worked

	// sleep for 100ms to make sure we get time to complete the goroutine
	time.Sleep(time.Millisecond * time.Duration(100))
	if !worked {
		t.Fatalf("Should work given item")
	}

}

func TestQueueUp(t *testing.T) {

	queue := NewQueue(10)

	queue.Handler = func(event interface{}) {
		e := event.(time.Duration)
		time.Sleep(time.Second * e)
	}

	go queue.Work()

	for _, d := range []time.Duration{1, 2, 3} {
		go func() {
			queue.Event <- d
		}()
	}

	// 2 because the first item is being worked
	time.Sleep(time.Millisecond * 100)
	if queue.Length() != 2 {
		t.Fatalf("Should queue up other items while work is being done")
	}
}

func TestSerialQueue(t *testing.T) {

	a := make([]time.Duration, 0)

	queue := NewQueue(10)

	// if items are not worked serially, they will be ordered ascendingly
	// into the array
	queue.Handler = func(event interface{}) {
		e := event.(time.Duration)
		time.Sleep(time.Millisecond * e)
		a = append(a, e)
	}

	go queue.Work()

	b := []time.Duration{4, 1, 2, 5, 3}
	for _, d := range b {
		go func(d time.Duration) {
			queue.Event <- d
		}(d)
	}

	time.Sleep(time.Second * 1)

	// this proves that the items in the array are in the same order that they were passed in,
	// not reordered by their sleep duration.
	if a[3] != b[3] {
		t.Fatalf("Should work items in serial order")
	}

}

func TestLocked(t *testing.T) {
	l := NewLocker()

	l.Lock()

	if !l.Locked {
		t.Fatalf("Should lock mutex")
	}

}

func TestUnlocked(t *testing.T) {
	l := NewLocker()

	l.Lock()
	l.Unlock()

	if l.Locked {
		t.Fatalf("Should unlock mutex")
	}

}
