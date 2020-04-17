package io

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Event struct {
	ID string
}

func TestBatch(t *testing.T) {
	c := &Config{
		UID:           "mock-uuid",
		BatchSize:     5,
		BatchInterval: 2,
		Sink:          "memory",
	}

	t.Run("Batch", func(t *testing.T) {
		ch := make(chan interface{}, c.BatchSize)

		// Expected 3 batches.
		t.Run("Send and Receive Events", func(t *testing.T) {
			go func() {
				for i := 0; i <= 13; i++ {
					ch <- &Event{ID: string(i)}
				}
			}()

			batchLen := []int{}
			idents := []string{}

			for ix := 0; ix < 3; ix++ {
				b, ident := Batch(ch, c)
				batchLen = append(batchLen, len(b))
				idents = append(idents, ident)
			}

			assert.ElementsMatch(t, batchLen, []int{5, 5, 4})

			t.Run("Another batch get should return 0 after timeout", func(t *testing.T) {
				b, _ := Batch(ch, c)
				assert.Equal(t, len(b), 0)
			})
		})
	})
}

func TestMain(m *testing.M) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	os.Exit(m.Run())
}
