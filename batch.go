package main

import (
	"github.com/dgraph-io/ristretto"
	"sync"
)

func NewBatch(
	uuid string, size, interval int, f Flusher, c *ristretto.Cache, stopCh <-chan struct{}, wg *sync.WaitGroup,
) chan<- *L9Event {
	eventChan := make(chan *L9Event, size)
	go func() {
		for {
			if isEmpty := flush(eventChan, interval, uuid, c, f); isEmpty {
				select {
				case <-stopCh:
					wg.Done()
					return
				}
			}
		}
	}()

	return eventChan
}
