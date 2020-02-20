package main

import "github.com/dgraph-io/ristretto"

func cacheClient() (*ristretto.Cache, error) {
	return ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7, // number of keys to track frequency of (10M).
		MaxCost:     10000000,
		BufferItems: 64, // number of keys per Get buffer.
	})
}
