package main

import (
	"fmt"
	"reflect"
	"sync"
)

type ChangeDeltaEvent struct {
	NewDefaultRoute string
}

type RouteUpdateEvent struct {
	Added   []string
	Removed []string
}

type Bus struct {
	mu     sync.RWMutex
	events map[reflect.Type][]any
}

func NewBus() *Bus {
	return &Bus{events: make(map[reflect.Type][]any)}
}

type Consumer[T any] struct {
	queue chan T
}

func NewConsumer[T any]() *Consumer[T] {
	return &Consumer[T]{queue: make(chan T, 16)}
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
		consumer.(*Consumer[T]).queue <- event
	}
}

func main() {
	bus := NewBus()

	deltaSub := Subscribe[ChangeDeltaEvent](bus)
	routeSub := Subscribe[RouteUpdateEvent](bus)

	go func() {
		Publish(bus, ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"})
		Publish(bus, RouteUpdateEvent{Added: []string{"10.0.0.0/8"}})
	}()

	v1 := <-deltaSub.queue
	fmt.Printf("new default route: %v\n", v1)

	v2 := <-routeSub.queue
	fmt.Printf("new route: %v\n", v2)
}
