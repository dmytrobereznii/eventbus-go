package bus

import (
	"errors"
)

type Queue[T any] struct {
	vals     []T
	start    int
	capacity int
	zero     T
}

func NewQueue[T any](cap int) *Queue[T] {
	return &Queue[T]{
		vals:     make([]T, 0, cap),
		capacity: cap,
	}
}

func (q *Queue[T]) canAppend() bool {
	return q.capacity == 0 || q.capacity > len(q.vals) || q.capacity > cap(q.vals)
}

func (q *Queue[T]) Len() int {
	return len(q.vals) - q.start
}

func (q *Queue[T]) Full() bool {
	return q.capacity != 0 && q.Len() == q.capacity
}

func (q *Queue[T]) Empty() bool {
	return q.Len() == 0
}

func (q *Queue[T]) Add(v T) error {
	if !q.canAppend() {
		if q.start == 0 {
			return errors.New("queue is full")
		}

		n := copy(q.vals, q.vals[q.start:])
		clear(q.vals[n:])
		q.vals = q.vals[:n]
		q.start = 0
	}

	q.vals = append(q.vals, v)
	return nil
}

func (q *Queue[T]) Drop() {
	if q.Empty() {
		return
	}

	q.vals[q.start] = q.zero
	q.start++

	if q.Empty() {
		q.start = 0
		q.vals = q.vals[:0]
	}
}

func (q *Queue[T]) Peek() T {
	if q.Empty() {
		return q.zero
	}

	return q.vals[q.start]
}
