// Package queue предоставляет потокобезопасную универсальную (generic) очередь с возможностью
// временного закрытия на запись.
//
// Очередь реализована на срезах и защищена мьютексами для конкурентного доступа.
// Поддерживает основные операции: добавление, извлечение, проверка длины и флаг "закрыта/открыта".
//
// Используется для асинхронного сбора и обработки данных, таких как метрики.
package queue

import "sync"

// Queue представляет потокобезопасную очередь с элементами произвольного типа.
//
// Очередь поддерживает динамическое добавление (`Push`) и удаление (`Pop`) элементов,
// а также может быть "закрыта" методом `Close`, после чего `Push` не будет добавлять новые элементы.
type Queue[T any] struct {
	data     []T          // Хранилище элементов очереди
	isClosed bool         // Флаг закрытия очереди
	m        sync.RWMutex // Мьютекс для синхронизации доступа
}

// New создаёт новую пустую очередь указанного типа T.
func New[T any]() Queue[T] {
	return Queue[T]{
		data: make([]T, 0),
	}
}

// Push добавляет элемент в конец очереди.
//
// Если очередь закрыта (`IsClosed` == true), элемент не будет добавлен.
func (q *Queue[T]) Push(e T) {
	if q.isClosed {
		return
	}

	q.m.Lock()
	defer q.m.Unlock()

	q.data = append(q.data, e)
}

// Pop извлекает первый элемент из очереди.
//
// Поведение не определено, если очередь пуста — это ответственность вызывающего.
func (q *Queue[T]) Pop() T {
	q.m.Lock()
	defer q.m.Unlock()

	h := q.data
	var top T
	top, q.data = h[0], h[1:]
	return top
}

// Len возвращает текущее количество элементов в очереди.
func (q *Queue[T]) Len() int {
	q.m.RLock()
	defer q.m.RUnlock()
	return len(q.data)
}

// Close закрывает очередь для добавления новых элементов.
//
// После вызова `Close`, любые последующие вызовы `Push` будут игнорироваться.
func (q *Queue[T]) Close() {
	q.m.Lock()
	defer q.m.Unlock()

	q.isClosed = true
}

// Open открывает очередь после её закрытия.
//
// После вызова `Open` снова можно добавлять элементы с помощью `Push`.
func (q *Queue[T]) Open() {
	q.m.Lock()
	defer q.m.Unlock()

	q.isClosed = false
}

// IsClosed возвращает true, если очередь закрыта для добавления новых элементов.
func (q *Queue[T]) IsClosed() bool {
	q.m.RLock()
	defer q.m.RUnlock()

	return q.isClosed
}
