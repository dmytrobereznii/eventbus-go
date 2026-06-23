package bus_test

import (
	"eventbus/bus"
	"fmt"
	"reflect"
	"testing"
)

type ChangeDeltaEvent struct {
	NewDefaultRoute string
}

type RouteUpdateEvent struct {
	Added   []string
	Removed []string
}

func TestSingleEvent(t *testing.T) {
	b := bus.NewBus()
	netmon := b.Client("netmon")
	backend := b.Client("ipnlocal")

	deltaSub := bus.Subscribe[ChangeDeltaEvent](backend)

	want1 := ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"}

	bus.Publish(netmon, want1)

	got1 := <-deltaSub.Queue
	if !reflect.DeepEqual(got1, want1) {
		t.Errorf("got %v, want %v", got1, want1)
	}
}

func TestDifferentEvents(t *testing.T) {
	b := bus.NewBus()
	netmon := b.Client("netmon")
	backend := b.Client("ipnlocal")

	deltaCon := bus.Subscribe[ChangeDeltaEvent](backend)
	routeCon := bus.Subscribe[RouteUpdateEvent](backend)

	want1 := ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"}
	want2 := RouteUpdateEvent{Added: []string{"10.0.0.0/8"}}

	bus.Publish(netmon, want1)
	bus.Publish(netmon, want2)

	got1 := <-deltaCon.Queue
	if !reflect.DeepEqual(got1, want1) {
		t.Errorf("got %v, want %v", got1, want1)
	}

	got2 := <-routeCon.Queue
	if !reflect.DeepEqual(got2, want2) {
		t.Errorf("got %v, want %v", got2, want2)
	}
}

func TestTypeIsolation(t *testing.T) {
	b := bus.NewBus()
	netmon := b.Client("netmon")
	backend := b.Client("ipnlocal")

	deltaCon := bus.Subscribe[ChangeDeltaEvent](backend)
	routeCon := bus.Subscribe[RouteUpdateEvent](backend)

	want := ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"}

	bus.Publish(netmon, want)

	got1 := <-deltaCon.Queue
	if !reflect.DeepEqual(got1, want) {
		t.Errorf("got %v, want %v", got1, want)
	}

	select {
	case got2 := <-routeCon.Queue:
		t.Errorf("got %v, want nothing", got2) //todo: fix after delivery is async
	default:
	}
}

func TestFanOut(t *testing.T) {
	b := bus.NewBus()
	netmon := b.Client("netmon")

	var consumers []*bus.Subscriber[ChangeDeltaEvent]

	for i := range 2 {
		client := b.Client(fmt.Sprintf("ipnlocal-%d", i))
		consumers = append(consumers, bus.Subscribe[ChangeDeltaEvent](client))
	}

	want := ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"}

	bus.Publish(netmon, want)

	for _, consumer := range consumers {
		got := <-consumer.Queue
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestNoConsNoop(t *testing.T) {
	b := bus.NewBus()
	netmon := b.Client("netmon")

	bus.Publish(netmon, ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"})
}
