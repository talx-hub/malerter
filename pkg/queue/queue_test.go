package queue_test

import (
	"testing"

	"github.com/talx-hub/malerter/pkg/queue"
)

func TestQueueBasicOperations(t *testing.T) {
	q := queue.New[int]()

	if q.Len() != 0 {
		t.Fatalf("expected length 0, got %d", q.Len())
	}

	q.Push(1)
	q.Push(2)
	q.Push(3)

	if q.Len() != 3 {
		t.Fatalf("expected length 3, got %d", q.Len())
	}

	v := q.Pop()
	if v != 1 {
		t.Errorf("expected 1, got %d", v)
	}

	if q.Len() != 2 {
		t.Errorf("expected length 2 after pop, got %d", q.Len())
	}
}

func TestQueueCloseAndOpen(t *testing.T) {
	q := queue.New[string]()

	q.Push("first")
	q.Close()

	if !q.IsClosed() {
		t.Error("expected queue to be closed")
	}

	q.Push("second") // должен быть проигнорирован

	if q.Len() != 1 {
		t.Errorf("expected length 1 after push to closed queue, got %d", q.Len())
	}

	q.Open()

	if q.IsClosed() {
		t.Error("expected queue to be open")
	}

	q.Push("second")

	if q.Len() != 2 {
		t.Errorf("expected length 2 after reopening, got %d", q.Len())
	}
}

func TestQueuePopOrder(t *testing.T) {
	q := queue.New[int]()
	for i := range 5 {
		q.Push(i)
	}

	for i := range 5 {
		v := q.Pop()
		if v != i {
			t.Errorf("expected %d, got %d", i, v)
		}
	}
}

func TestQueueConcurrencySafety(t *testing.T) {
	q := queue.New[int]()
	const total = 1000

	done := make(chan struct{})
	go func() {
		for i := range total {
			q.Push(i)
		}
		done <- struct{}{}
	}()

	go func() {
		for range total {
			_ = q.Len()
			_ = q.IsClosed()
		}
		done <- struct{}{}
	}()

	<-done
	<-done

	if q.Len() != total {
		t.Errorf("expected length %d, got %d", total, q.Len())
	}
}
