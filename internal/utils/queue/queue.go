package queue

import "sync"

// Нужна бесконечная очередь для метрик.

type Queue[T any] struct {
	data     []T
	isClosed bool
	m        sync.RWMutex
}

func (q *Queue[T]) Push(e T) {
	if q.isClosed {
		return
	}

	q.m.Lock()
	defer q.m.Unlock()

	q.data = append(q.data, e)
}

func (q *Queue[T]) Pop() T {
	q.m.Lock()
	defer q.m.Unlock()

	h := q.data
	var top T
	top, q.data = h[0], h[1:]
	return top
}

func (q *Queue[T]) Len() int {
	q.m.RLock()
	defer q.m.RUnlock()
	return len(q.data)
}

func (q *Queue[T]) Close() {
	q.m.Lock()
	defer q.m.Unlock()

	q.isClosed = true
}

func (q *Queue[T]) Open() {
	q.m.Lock()
	defer q.m.Unlock()

	q.isClosed = false
}

func (q *Queue[T]) IsClosed() bool {
	q.m.RLock()
	defer q.m.RUnlock()

	return q.isClosed
}

func New[T any]() Queue[T] {
	return Queue[T]{
		data: make([]T, 0),
	}
}
