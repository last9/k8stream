package io

import (
	"encoding/json"
	"log"
)

type Flusher interface {
	Flush(uuid, ident string, d []byte) error
	LoadConfig(json.RawMessage) error
}

func GetFlusher(conf *Config) (Flusher, error) {
	var f Flusher
	switch conf.Sink {
	case "s3":
		f = &S3Sink{}
	case "file":
		f = &FileSink{}
	case "memory":
		f = &MemSink{Records: map[string][]byte{}, OnFetch: func(id string) {
			log.Println("Flushing", id)
		}}
	}

	if err := f.LoadConfig(conf.Raw); err != nil {
		return nil, err
	}

	return f, nil
}
