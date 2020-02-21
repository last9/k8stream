package main

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"gopkg.in/go-playground/assert.v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type events struct {
	Items []*v1.Event `json:"items"`
}

func TestMakeL9Event(t *testing.T) {
	var wg sync.WaitGroup
	e := &events{}
	ch := make(chan *L9Event)
	mCache, err := cacheClient()
	if err != nil {
		t.Fatal(err)
	}

	key := "814a4994-e977-4e07-be69-6a464b2169c9"

	t.Run("Wait till you get the key", func(t *testing.T) {
		var ok bool
		for i := 0; i < 10; i++ {
			mCache.Set(key, &unstructured.Unstructured{}, 1)
			if _, ok = mCache.Get(key); ok {
				t.Log(i)
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		assert.Equal(t, true, ok)
	})

	t.Run("Make Last9 Event", func(t *testing.T) {
		b, err := ioutil.ReadFile("testdata/events.log")
		if err != nil {
			t.Fatal(err)
		}

		if err := json.Unmarshal(b, e); err != nil {
			t.Fatal(err)
		}

		ev := makeL9Event(
			e.Items[0], nil, []string{"127.0.0.1"},
		)

		assert.Equal(t, ev.ID, "19b4506f-95f4-4dd0-8d2d-bf7647997877")
		assert.Equal(t, ev.Address[0], "127.0.0.1")
	})

	t.Run("Receive event over Handler", func(t *testing.T) {
		go func() {
			defer wg.Done()
			x := <-ch
			assert.Equal(t, "Scheduled", x.Reason)
		}()

		wg.Add(1)
		h := &Handler{&kubernetesClient{}, ch, mCache}
		h.OnAdd(e.Items[0])
		wg.Wait()
	})
}
