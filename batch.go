package main

import (
	"github.com/dgraph-io/ristretto"
)

func NewBatch(
	uuid string, size, interval int, f Flusher, c *ristretto.Cache,
) chan<- *L9Event {
	eventChan := make(chan *L9Event, size)
	go func() {
		for {
			flush(eventChan, interval, uuid, c, f)
		}
	}()

	return eventChan
}
