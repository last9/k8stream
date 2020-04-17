package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/golang/protobuf/proto"
	"gopkg.in/go-playground/assert.v1"
)

type MemFlush struct {
	uuid    string
	records map[string][]byte
	onFetch func(string)
	last    string
}

func (m *MemFlush) LoadConfig(_ json.RawMessage) error {
	return nil
}

func (m *MemFlush) Flush(uuid, ident string, d []byte) error {
	defer m.onFetch(uuid)
	m.last = ident
	m.uuid = uuid
	m.records[ident] = d
	return nil
}

func TestBatch(t *testing.T) {
	uuid := "mock-uuid"

	t.Run("Batch", func(t *testing.T) {
		var wg sync.WaitGroup

		f := &MemFlush{
			records: map[string][]byte{},
			onFetch: func(ident string) {
				wg.Done()
			},
		}

		ch := NewBatch(uuid, 5, 2, f, nil)
		assert.Equal(t, 5, cap(ch))

		// Expected 3 batches.
		t.Run("Send and Receive Events", func(t *testing.T) {
			wg.Add(5)
			for i := 0; i <= 20; i++ {
				ch <- &L9Event{ID: string(i)}
			}

			wg.Wait()
			assert.Equal(t, 5, len(f.records))
			assert.Equal(t, uuid, f.uuid)

			t.Run("Protobuf unmarshal output", func(t *testing.T) {
				for ix, b := range f.records {
					ne := &L9EventBatch{}
					if err := proto.Unmarshal(b, ne); err != nil {
						t.Fatal(err)
					}

					if ix == f.last {
						assert.Equal(t, 1, len(ne.Events))
					} else {
						assert.Equal(t, 5, len(ne.Events))
					}
				}
			})

		})

	})
}

func TestMain(m *testing.M) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	os.Exit(m.Run())
}
