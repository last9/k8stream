package main

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"testing"

	"gopkg.in/go-playground/assert.v1"
	v1 "k8s.io/api/core/v1"
)

type events struct {
	Items []*v1.Event `json:"items"`
}

func TestMakeL9Event(t *testing.T) {
	e := &events{}
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

	kc, err := newK8sClient("testdata/kubeconfig")
	if err != nil {
		t.Fatal(err)
	}

	mcache, err := cacheClient()
	if err != nil {
		t.Fatal(err)
	}
	mcache.Set("814a4994-e977-4e07-be69-6a464b2169c9", nil, 0)

	var wg sync.WaitGroup
	ch := make(chan *L9Event)
	go func() {
		defer wg.Done()
		x := <-ch
		assert.Equal(t, "Scheduled", x.Reason)
	}()

	wg.Add(1)
	h := &Handler{kc, ch, mcache}
	h.OnAdd(e.Items[0])
	wg.Wait()
}
