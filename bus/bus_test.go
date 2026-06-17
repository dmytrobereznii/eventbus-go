package bus_test

import (
	"eventbus/bus"
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

	deltaSub := bus.Subscribe[ChangeDeltaEvent](b)

	want1 := ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"}

	bus.Publish(b, want1)

	got1 := <-deltaSub.Queue
	if !reflect.DeepEqual(got1, want1) {
		t.Errorf("got %v, want %v", got1, want1)
	}
}

func TestDifferentEvents(t *testing.T) {
	b := bus.NewBus()

	deltaCon := bus.Subscribe[ChangeDeltaEvent](b)
	routeCon := bus.Subscribe[RouteUpdateEvent](b)

	want1 := ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"}
	want2 := RouteUpdateEvent{Added: []string{"10.0.0.0/8"}}

	bus.Publish(b, want1)
	bus.Publish(b, want2)

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

	deltaCon := bus.Subscribe[ChangeDeltaEvent](b)
	routeCon := bus.Subscribe[RouteUpdateEvent](b)

	want := ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"}

	bus.Publish(b, want)

	got1 := <-deltaCon.Queue
	if !reflect.DeepEqual(got1, want) {
		t.Errorf("got %v, want %v", got1, want)
	}

	select {
	case got2 := <-routeCon.Queue:
		t.Errorf("got %v, want nothing", got2)
	default:
	}
}

func TestFanOut(t *testing.T) {
	b := bus.NewBus()

	var consumers []*bus.Consumer[ChangeDeltaEvent]

	for range 2 {
		consumers = append(consumers, bus.Subscribe[ChangeDeltaEvent](b))
	}

	want := ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"}

	bus.Publish(b, want)

	for _, consumer := range consumers {
		got := <-consumer.Queue
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestNoConsNoop(t *testing.T) {
	b := bus.NewBus()

	bus.Publish(b, ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"})
}
