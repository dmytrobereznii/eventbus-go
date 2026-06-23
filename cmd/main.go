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
	netmon := b.Client("netmon")    // publishes network changes
	backend := b.Client("ipnlocal") // subscribes to network changes

	deltaSub := bus.Subscribe[ChangeDeltaEvent](backend)
	routeSub := bus.Subscribe[RouteUpdateEvent](backend)

	go func() {
		bus.Publish(netmon, ChangeDeltaEvent{NewDefaultRoute: "192.168.1.1"})
		bus.Publish(netmon, RouteUpdateEvent{Added: []string{"10.0.0.0/8"}})
	}()

	v1 := <-deltaSub.Queue
	fmt.Printf("new default route: %v\n", v1)

	v2 := <-routeSub.Queue
	fmt.Printf("new route: %v\n", v2)

	b.Close()
}
