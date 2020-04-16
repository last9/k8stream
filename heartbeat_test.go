package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func TestStartHeartbeat(t *testing.T) {
	uids := make(chan string, 2)
	upgradeUid := "test"
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := r.URL.Query().Get("uid")
		if uid == upgradeUid {
			w.WriteHeader(http.StatusUpgradeRequired)
			return
		}
		if uid != "" {
			uids <- uid
		}

		w.Write([]byte("Ok"))
	}))

	t.Run("Receive heartbeat", func(t *testing.T) {
		uid := string(uuid.NewUUID())
		interval := 1

		if err := StartHeartbeat(uid, s.URL, interval, 0,
			nil, nil, nil,
			func(c chan struct{}, c2 chan struct{}, group *sync.WaitGroup) {}); err != nil {
			t.Errorf("failed to send heartbeat, error occured %s", err.Error())
		}

		select {
		case received := <-uids:
			assert.Equal(t, uid, received)
			return
		case <-time.After(time.Duration(interval+3) * time.Second):
			t.Error("no heartbeat in expected interval")
		}
	})

	t.Run("Error out on 426 upgrade required", func(t *testing.T) {
		ch := make(chan struct{})
		err := StartHeartbeat(upgradeUid, s.URL, 1, 0, ch, nil, nil,
			func(c1 chan struct{}, c2 chan struct{}, group *sync.WaitGroup) {
				c1 <- struct{}{}
			})

		assert.Nil(t, err)
		data := <-ch
		assert.NotNil(t, data)
	})
}
