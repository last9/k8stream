package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto"
	proto "github.com/golang/protobuf/proto"
)

type Flusher interface {
	Flush(uuid, ident string, d []byte) error
	LoadConfig(json.RawMessage) error
}

func getFlusher(conf *L9K8streamConfig, b json.RawMessage) (Flusher, error) {
	var f Flusher
	switch conf.Sink {
	case "s3":
		f = &S3Sink{}
	case "file":
		f = &FileSink{}
	}

	if err := f.LoadConfig(b); err != nil {
		return nil, err
	}

	return f, nil
}

func flushEvents(q <-chan *L9Event, interval int) (
	ne *L9EventBatch, ident string,
) {
	var ix int
	defer func() {
		ne.Events = ne.Events[:ix]
		ident = strconv.FormatInt(time.Now().UnixNano(), 10)
	}()

	size := cap(q)
	ne = &L9EventBatch{Events: make([]*L9Event, size)}

	for ; ix < size; ix++ {
		select {
		case <-time.After(time.Duration(interval) * time.Second):
			return
		case x := <-q:
			ne.Events[ix] = x
		}
	}

	return
}

func flush(
	q <-chan *L9Event, interval int, uid string, cache *ristretto.Cache,
	sink Flusher,
) {
	ne, ident := flushEvents(q, interval)
	if len(ne.Events) == 0 {
		return
	}

	bytes, err := proto.Marshal(ne)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Flushing", len(ne.Events), "items", ident)
	if err := sink.Flush(uid, ident, bytes); err != nil {
		log.Println(err)
		return
	}

	for _, e := range ne.Events {
		cache.Set(e.ID, true, 0)
	}
}
