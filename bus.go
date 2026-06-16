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

type Bus struct {
	mu     sync.RWMutex
	events map[reflect.Type][]func(any)
}

func NewBus() *Bus {
	return &Bus{events: make(map[reflect.Type][]func(any))}
}

func Subscribe[T any](b *Bus, h func(T)) {
	wrapper := func(e any) {
		h(e.(T))
	}

	b.mu.Lock()
	b.events[reflect.TypeFor[T]()] = append(b.events[reflect.TypeFor[T]()], wrapper)
	b.mu.Unlock()
}

func Publish[T any](b *Bus, e T) {
	b.mu.RLock()
	handlers, ok := b.events[reflect.TypeFor[T]()]
	b.mu.RUnlock()
	if !ok {
		return
	}

	for _, h := range handlers {
		h(e) // sync, so slow handler delays whole flow
	}
}

func main() {
	bus := NewBus()

	Subscribe[ChangeDelta](bus, func(e ChangeDelta) {
		fmt.Printf("Delta changed %s\n", e.NewDefaultRoute)
	})
	Subscribe[RouteUpdate](bus, func(e RouteUpdate) {
		fmt.Printf("Route changed %s:%s\n", e.Added, e.Removed)
	})

	Publish[ChangeDelta](bus, ChangeDelta{NewDefaultRoute: "192.168.1.1"})
	Publish[RouteUpdate](bus, RouteUpdate{Added: []string{"10.0.0.0/8"}})
}
