package io

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestStartHeartbeat(t *testing.T) {
	uids := make(chan string, 2)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := r.URL.Query().Get("uid")
		if uid != "" {
			uids <- uid
		}

		w.Write([]byte("Ok"))
	}))

	uid := uuid.NewV4().String()
	interval := 1

	assert.Nil(t, StartHeartbeat(uid, s.URL, interval))

	select {
	case received := <-uids:
		assert.Equal(t, uid, received)
		return
	case <-time.After(time.Duration(interval+3) * time.Second):
		t.Error("no heartbeat in expected interval")
	}
}
