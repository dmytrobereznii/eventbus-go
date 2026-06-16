package bus

import (
	"reflect"
	"sync"
)

type Bus struct {
	mu     sync.RWMutex
	events map[reflect.Type][]any
}

func NewBus() *Bus {
	return &Bus{events: make(map[reflect.Type][]any)}
}

type Consumer[T any] struct {
	Queue chan T
}

func NewConsumer[T any]() *Consumer[T] {
	return &Consumer[T]{Queue: make(chan T, 16)}
}

func Subscribe[T any](b *Bus) *Consumer[T] {
	sub := NewConsumer[T]()

	b.mu.Lock()
	b.events[reflect.TypeFor[T]()] = append(b.events[reflect.TypeFor[T]()], sub)
	b.mu.Unlock()

	return sub
}

func Publish[T any](b *Bus, event T) {
	b.mu.RLock()
	consumers, ok := b.events[reflect.TypeFor[T]()]
	b.mu.RUnlock()
	if !ok {
		return
	}

	for _, consumer := range consumers {
		consumer.(*Consumer[T]).Queue <- event
	}
}
