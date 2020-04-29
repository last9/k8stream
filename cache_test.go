package main

import (
	"log"
	"strconv"
	"sync"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

type testItem struct {
	Foo string
	Id  int
}

func TestCache(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	c, err := newCache()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Internal tests", func(t *testing.T) {
		t.Run("Set a key", func(t *testing.T) {
			var wg sync.WaitGroup
			for ix := 0; ix < 4; ix++ {
				wg.Add(1)
				go func(val int) {
					defer wg.Done()
					if err := c.Set("table", strconv.Itoa(val), val); err != nil {
						t.Fatal(err)

					}
				}(ix)
			}

			wg.Wait()

			// Should run the next few gets in Parallel
			t.Parallel()

			for ix := 0; ix < 4; ix++ {
				t.Run("Get a key", func(t *testing.T) {
					var val int
					r, err := c.Get("table", strconv.Itoa(ix))
					if err != nil {
						t.Fatal(err)
					}

					if !r.Exists() {
						t.Fatal("Value", ix, "is missing")
					}

					if err := r.Unmarshal(&val); err != nil {
						t.Fatal(err)
					} else {
						assert.Equal(t, ix, val)
					}

				})
			}
		})
	})

	t.Run("Insert a random struct", func(t *testing.T) {
		if err := c.Set("test", "uid1", &testItem{Foo: "foo", Id: 1}); err != nil {
			t.Fatal(err)
		}

		t.Run("Fetch the struct", func(t *testing.T) {
			var ret testItem
			res, err := c.Get("test", "uid1")
			if err != nil {
				t.Fatal(err)
			}

			if err := res.Unmarshal(&ret); err != nil {
				t.Fatal(err)
			} else {
				assert.Equal(t, ret.Foo, "foo")
				assert.Equal(t, ret.Id, 1)
			}
		})
	})
}
