package bus_test

import (
	"eventbus/bus"
	"fmt"
	"testing"
)

func TestAddReflectsInLen(t *testing.T) {
	q := bus.NewQueue[string](0)
	want := 3

	for i := range want {
		q.Add(fmt.Sprintf("val%d", i))
	}

	if q.Len() != want {
		t.Errorf("q.Len() = %d, want %d", q.Len(), want)
	}
}

func TestFullWhenReachedCapacity(t *testing.T) {
	c := 3
	q := bus.NewQueue[int](c)

	if q.Full() {
		t.Errorf("q.Full() = true, want false")
	}

	for i := range c {
		q.Add(i)
	}

	if !q.Full() {
		t.Errorf("q.Full() = false, want true")
	}
}

func TestEmpty(t *testing.T) {
	q := bus.NewQueue[int](0)

	if !q.Empty() {
		t.Errorf("q.Empty() = false, want true")
	}

	q.Add(1)

	if q.Empty() {
		t.Errorf("q.Full() = true, want false")
	}
}

func TestDropOnEmptyNoop(t *testing.T) {
	q := bus.NewQueue[int](0)
	q.Drop()
}

func TestDropRecalculatesLen(t *testing.T) {
	q := bus.NewQueue[int](0)
	l := 3

	for i := range l {
		q.Add(i)
	}

	q.Drop()

	want := l - 1
	if q.Len() != want {
		t.Errorf("q.Len() = %d, want %d", q.Len(), want)
	}
}

func TestEmptyingQueueSetsLenToZero(t *testing.T) {
	q := bus.NewQueue[int](1)
	q.Add(1)
	q.Drop()

	if q.Full() {
		t.Errorf("got true, want false")
	}

	if q.Len() != 0 {
		t.Errorf("got %d, want 0", q.Len())
	}
}

func TestCantBeFullWithUnboundedCapacity(t *testing.T) {
	q := bus.NewQueue[int](0)

	if q.Full() {
		t.Errorf("got true, want false")
	}
}

func TestPeek(t *testing.T) {
	q := bus.NewQueue[int](0)
	want := 1
	q.Add(want)
	v := q.Peek()

	if v != want {
		t.Errorf("got %d, want %d", v, want)
	}
}

func TestPeekOnEmptyQueueReturnsZeroValue(t *testing.T) {
	q := bus.NewQueue[string](0)
	v := q.Peek()
	if v != "" {
		t.Errorf("got %v, want zero value", v)
	}
}

func TestCapacityIsLimited(t *testing.T) {
	q := bus.NewQueue[int](1)

	err := q.Add(1)
	if err != nil {
		t.Errorf("got %v, want nil", err)
	}

	err = q.Add(2)
	if err == nil {
		t.Errorf("got nil, want error")
	}
}

func TestCanAddAfterDroppingElement(t *testing.T) {
	q := bus.NewQueue[int](2)

	q.Add(1)
	q.Add(2)
	q.Drop()
	q.Add(3)
	err := q.Add(4)
	if err == nil {
		t.Errorf("got nil, want error")
	}
}
