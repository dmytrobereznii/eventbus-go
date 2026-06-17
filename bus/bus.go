package bus

import (
	"reflect"
	"sync"
)

type Bus struct {
	mu     sync.RWMutex
	events map[reflect.Type][]*subscribeState
	write  chan PublishedEvent
	done   chan struct{}
}

func NewBus() *Bus {
	b := &Bus{
		events: make(map[reflect.Type][]*subscribeState),
		write:  make(chan PublishedEvent),
		done:   make(chan struct{}),
	}
	go b.pump()
	return b
}

func (b *Bus) pump() {
	defer close(b.done)
	for e := range b.write { //blocks waiting for values, exists only on channel close
		t := reflect.TypeOf(e.Event)
		b.mu.RLock()
		subStates := b.events[t]
		b.mu.RUnlock()

		for _, s := range subStates {
			s.write <- DeliveredEvent{Event: e.Event}
		}
	}
}

func (b *Bus) Close() {
	close(b.write)
	<-b.done
}

type PublishedEvent struct {
	Event any
}

type DeliveredEvent struct {
	Event any
}

type subscribeState struct {
	write chan DeliveredEvent
}

type Consumer[T any] struct {
	state *subscribeState
	Queue chan T
}

func NewConsumer[T any](state *subscribeState) *Consumer[T] {
	return &Consumer[T]{
		state: state,
		Queue: make(chan T),
	}
}

func (c *Consumer[T]) pump() {
	for e := range c.state.write {
		c.Queue <- e.Event.(T)
	}
}

func Subscribe[T any](b *Bus) *Consumer[T] {
	state := &subscribeState{write: make(chan DeliveredEvent)}

	con := NewConsumer[T](state)

	b.mu.Lock()
	b.events[reflect.TypeFor[T]()] = append(b.events[reflect.TypeFor[T]()], state)
	b.mu.Unlock()

	go con.pump()

	return con
}

func Publish[T any](b *Bus, event T) {
	b.write <- PublishedEvent{Event: event}
}
