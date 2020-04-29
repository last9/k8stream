package main

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/last9/k8stream/io"
)

const (
	lineBreak = "\n"
)

// Ingester returns a send-only message chan. It also spins a goroutine that runs
// a forever loop of listening to messages and flush them to disk till the buffer
// overflows the batchSize or the lease if past the batchInterval. While a batch
// is being flushed, the channels stop listening.
func startIngester(f io.Flusher, cfg *L9K8streamConfig, db Cachier) chan<- interface{} {
	msgChan := make(chan interface{}, cfg.BatchSize)
	go func() {
		for {
			if err := doBatch(f, msgChan, db, cfg); err != nil {
				log.Println(err)
			}
		}
	}()

	return msgChan
}

func doBatch(
	f io.Flusher, msgChan <-chan interface{},
	db Cachier, cfg *L9K8streamConfig,
) error {
	batch, batchIdent := io.Batch(msgChan, &cfg.Config)
	if len(batch) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, v := range batch {
		bytes, err := json.Marshal(v)
		if err != nil {
			return err
		}

		buf.Write(bytes)
		buf.Write([]byte(lineBreak))
	}

	if err := f.Flush(cfg.UID, batchIdent, buf.Bytes()); err != nil {
		return err
	}

	if db != nil {
		for _, v := range batch {
			e := v.(*L9Event)
			db.ExpireSet(eventCacheTable, e.ID, true, objectCacheExpiry)
		}
	}

	return nil
}
