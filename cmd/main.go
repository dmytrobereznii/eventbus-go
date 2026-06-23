package main

import (
	"eventbus/bus"
	"fmt"
)

type ChangeDeltaEvent struct {
	NewDefaultRoute string
}

type RouteUpdateEvent struct {
	Added   []string
	Removed []string
}

func main() {
	b := bus.NewBus()
	c := b.Client("client")

	deltaSub := bus.Subscribe[ChangeDeltaEvent](c)
	routeSub := bus.Subscribe[RouteUpdateEvent](c)

	go func() {
		bus.Publish(c, ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"})
		bus.Publish(c, RouteUpdateEvent{Added: []string{"10.0.0.0/8"}})
	}()

	v1 := <-deltaSub.Queue
	fmt.Printf("new default route: %v\n", v1)

	v2 := <-routeSub.Queue
	fmt.Printf("new route: %v\n", v2)

	b.Close()
}
