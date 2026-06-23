package bus

import (
	"reflect"
	"sync"
)

// Bus routes events from publishers to subscribers
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

func (b *Bus) Client(name string) *Client {
	return &Client{
		Bus:  b,
		Name: name,
	}
}

func (b *Bus) Close() {
	close(b.write)
	<-b.done
}

type Client struct {
	Bus  *Bus
	Name string
}

// PublishedEvent typed wrapper over generic event sent to bus
type PublishedEvent struct {
	Event any
	From  *Client
	To    *Client
}

// DeliveredEvent typed wrapper over generic event received by subscriber
type DeliveredEvent struct {
	Event any
	From  *Client
	To    *Client
}

// subscribeState is Client's engine
type subscribeState struct {
	write chan DeliveredEvent
}

// Subscriber hold a subscription for a single event type
type Subscriber[T any] struct {
	state *subscribeState
	Queue chan T
}

func NewSubscriber[T any](state *subscribeState) *Subscriber[T] {
	return &Subscriber[T]{
		state: state,
		Queue: make(chan T),
	}
}

func (c *Subscriber[T]) pump() {
	for e := range c.state.write {
		c.Queue <- e.Event.(T)
	}
}

func Subscribe[T any](c *Client) *Subscriber[T] {
	b := c.Bus
	state := &subscribeState{write: make(chan DeliveredEvent)}

	sub := NewSubscriber[T](state)

	b.mu.Lock()
	b.events[reflect.TypeFor[T]()] = append(b.events[reflect.TypeFor[T]()], state)
	b.mu.Unlock()

	go sub.pump()

	return sub
}

func Publish[T any](c *Client, event T) {
	c.Bus.write <- PublishedEvent{
		Event: event,
		From:  c,
	}
}
