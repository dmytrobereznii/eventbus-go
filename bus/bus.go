package bus

import (
	"context"
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
	cancel context.CancelFunc
	done   chan struct{}
}

func NewBus() *Bus {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Bus{
		events: make(map[reflect.Type][]*subscribeState),
		write:  make(chan PublishedEvent),
		done:   make(chan struct{}),
		cancel: cancel,
	}
	go b.pump(ctx)
	return b
}

func (b *Bus) pump(ctx context.Context) {
	defer close(b.done)
	for {
		select {
		case e := <-b.write:
			t := reflect.TypeOf(e.Event)
			b.mu.RLock()
			subStates := b.events[t]
			b.mu.RUnlock()

			for _, ss := range subStates {
				ss.write <- DeliveredEvent{Event: e.Event, From: e.From, To: ss.client}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (b *Bus) subscribe(t reflect.Type, state *subscribeState) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events[t] = append(b.events[t], state)
}

func (b *Bus) Client(name string) *Client {
	return &Client{
		Bus:  b,
		Name: name,
	}
}

func (b *Bus) Close() {
	b.cancel()
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
		c.state = newSubscribeState(c)
	}
	return c.state
}

func (c *Client) Close() {
	c.mu.Lock()
	state := c.state
	c.state = nil
	c.mu.Unlock()
	if nil != state {
		state.cancel()
		<-state.done
	}
}

// subscribeState is Client's engine
type subscribeState struct {
	client *Client
	write  chan DeliveredEvent
	done   chan struct{}
	subs   map[reflect.Type]subscriber //map of event => subscriber
	mu     sync.RWMutex
	cancel context.CancelFunc
}

func newSubscribeState(c *Client) *subscribeState {
	ctx, cancel := context.WithCancel(context.Background())
	state := &subscribeState{
		client: c,
		write:  make(chan DeliveredEvent),
		done:   make(chan struct{}),
		subs:   make(map[reflect.Type]subscriber),
		cancel: cancel,
	}
	go state.pump(ctx)
	return state
}

func (state *subscribeState) pump(ctx context.Context) {
	defer close(state.done)
	var queue Queue[DeliveredEvent]

	acceptCh := func() chan DeliveredEvent {
		if queue.Full() {
			return nil
		}
		return state.write
	}

	for {
		if queue.Empty() {
			select {
			case val := <-state.write:
				queue.Add(val)
			case <-ctx.Done():
				return
			}
		} else {
			val := queue.Peek()
			state.mu.Lock()
			sub := state.subs[reflect.TypeOf(val.Event)]
			state.mu.Unlock()
			if sub == nil {
				queue.Drop()
				continue
			}

			if !sub.dispatch(ctx, &queue, acceptCh) {
				return
			}
		}
	}
}

// Subscriber hold a subscription for a single event type
type Subscriber[T any] struct {
	Ch chan T
}

func (sub *Subscriber[T]) dispatch(ctx context.Context, queue *Queue[DeliveredEvent], acceptCh func() chan DeliveredEvent) bool {
	t := queue.Peek().Event.(T)
	for {
		select {
		case sub.Ch <- t:
			queue.Drop()
			return true
		case val := <-acceptCh():
			queue.Add(val)
		case <-ctx.Done():
			return false
		}
	}
}

type subscriber interface {
	dispatch(ctx context.Context, vals *Queue[DeliveredEvent], acceptCh func() chan DeliveredEvent) bool
}

func NewSubscriber[T any]() *Subscriber[T] {
	return &Subscriber[T]{
		Ch: make(chan T),
	}
}

func Subscribe[T any](c *Client) *Subscriber[T] {
	t := reflect.TypeFor[T]()
	sub := NewSubscriber[T]()

	state := c.subscribeState()
	state.mu.Lock()
	state.subs[t] = sub // what if it existed before?
	state.mu.Unlock()

	c.Bus.subscribe(t, state)

	return sub
}

func Publish[T any](c *Client, e T) {
	c.Bus.write <- PublishedEvent{
		Event: e,
		From:  c,
	}
}
