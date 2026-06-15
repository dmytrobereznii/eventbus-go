# eventbus-go

A learning project that rebuilds Tailscale's [`util/eventbus`](https://github.com/tailscale/tailscale/tree/main/util/eventbus) from scratch in Go, stage by stage — from a naive `map[string][]func(any)` bus to the real typed, client-based design with async dispatch, ordering guarantees, and clean shutdown.