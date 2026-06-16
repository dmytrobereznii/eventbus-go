package main

import (
	"fmt"
	"sync"
)

type Handler func(e any)

type Bus struct {
	mu     sync.RWMutex
	topics map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{topics: make(map[string][]Handler)}
}

func (b *Bus) Subscribe(topic string, f Handler) {
	b.mu.Lock()
	b.topics[topic] = append(b.topics[topic], f)
	b.mu.Unlock()
}

func (b *Bus) Publish(topic string, msg any) {
	b.mu.RLock()
	subscribers, ok := b.topics[topic]
	b.mu.RUnlock()
	if !ok {
		return
	}

	for _, handler := range subscribers {
		handler(msg) // slow handler delays whole flow
	}
}

func main() {
	bus := NewBus()

	bus.Subscribe("user.created", func(msg any) {
		fmt.Println("user.created - msg:", msg)
	})
	bus.Subscribe("user.updated", func(msg any) {
		fmt.Println("user.updated - msg:", msg)
	})

	bus.Publish("user.created", "John Doe")
}
