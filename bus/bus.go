package bus

import (
	"reflect"
	"sync"
)

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
		states := b.events[t]
		b.mu.RUnlock()

		for _, s := range states {
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
	Bus   *Bus
	mu    sync.RWMutex
	state *subscribeState
	Name  string
}

func (c *Client) subscribeState() *subscribeState {
	c.mu.Lock()
	defer c.mu.Unlock()
	if nil == c.state {
		c.state = &subscribeState{
			write: make(chan DeliveredEvent),
			subs:  make(map[reflect.Type]subscriber),
		}
		go c.state.pump()
	}
	return c.state
}

// subscribeState is Client's engine
type subscribeState struct {
	write chan DeliveredEvent
	subs  map[reflect.Type]subscriber //map of event => subscriber
}

func (state *subscribeState) pump() {
	for e := range state.write {
		sub := state.subs[reflect.TypeOf(e.Event)]
		sub.send(e.Event)
	}
}

// Subscriber hold a subscription for a single event type
type Subscriber[T any] struct {
	Queue chan T
}

func (sub *Subscriber[T]) send(event any) {
	sub.Queue <- event.(T)
}

type subscriber interface {
	send(event any)
}

func NewSubscriber[T any]() *Subscriber[T] {
	return &Subscriber[T]{
		Queue: make(chan T),
	}
}

func Subscribe[T any](c *Client) *Subscriber[T] {
	t := reflect.TypeFor[T]()
	sub := NewSubscriber[T]()

	state := c.subscribeState()

	c.mu.Lock()
	state.subs[t] = sub // what if it existed before?
	c.mu.Unlock()

	c.Bus.events[t] = append(c.Bus.events[t], state)

	return sub
}

func Publish[T any](c *Client, event T) {
	c.Bus.write <- PublishedEvent{
		Event: event,
		From:  c,
	}
}
