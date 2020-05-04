package io

import (
	"strconv"
	"time"
)

func BatchNumber() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

// Listen to an Interface channel and return a buffer batch on
// Either a timeout happens
// OR buffer is filled to a size.
func Batch(ch <-chan interface{}, c *Config) (batch []interface{}, ident string) {
	batch = make([]interface{}, c.BatchSize)

	var ix int

	defer func() {
		batch = batch[:ix]
		ident = BatchNumber()
		return
	}()

	for ; ix < c.BatchSize; ix++ {
		select {
		case <-time.After(time.Duration(c.BatchInterval) * time.Second):
			c.Log("Flushing batch for Timeout %v", c.BatchInterval)
			return
		case x := <-ch:
			batch[ix] = x
		}
	}

	return
}
