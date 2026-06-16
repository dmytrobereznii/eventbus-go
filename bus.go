package main

import (
	"fmt"
	"reflect"
	"sync"
)

type ChangeDelta struct {
	NewDefaultRoute string
}

type RouteUpdate struct {
	Added   []string
	Removed []string
}

type Handler func(e any)

type Bus struct {
	mu     sync.RWMutex
	topics map[reflect.Type][]Handler
}

func NewBus() *Bus {
	return &Bus{topics: make(map[reflect.Type][]Handler)}
}

func Subscribe[T any](b *Bus, h Handler) {
	b.mu.Lock()
	b.topics[reflect.TypeFor[T]()] = append(b.topics[reflect.TypeFor[T]()], h)
	b.mu.Unlock()
}

func Publish[T any](b *Bus, e T) {
	b.mu.RLock()
	subscribers, ok := b.topics[reflect.TypeFor[T]()]
	b.mu.RUnlock()
	if !ok {
		return
	}

	for _, handler := range subscribers {
		handler(e) // slow handler delays whole flow
	}
}

func main() {
	bus := NewBus()

	Subscribe[ChangeDelta](bus, func(e any) {
		fmt.Printf("%s\n", e)
	})
	Subscribe[RouteUpdate](bus, func(e any) {
		fmt.Printf("%s\n", e)
	})

	Publish[ChangeDelta](bus, ChangeDelta{NewDefaultRoute: "192.168.1.1"})
	Publish[RouteUpdate](bus, RouteUpdate{Added: []string{"10.0.0.0/8"}})
}
