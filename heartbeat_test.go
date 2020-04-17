package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"
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

	uid := string(uuid.NewUUID())
	interval := 1

	assert.Nil(t, StartHeartbeat(uid, s.URL, interval, 0))

	select {
	case received := <-uids:
		assert.Equal(t, uid, received)
		return
	case <-time.After(time.Duration(interval+3) * time.Second):
		t.Error("no heartbeat in expected interval")
	}

	t.Run("upgrade should send SIGQUIT to main process", func(t *testing.T) {
		sigCh := make(chan os.Signal)
		signal.Notify(sigCh, syscall.SIGQUIT)

		if err := StartHeartbeat(upgradeUid, s.URL, interval, 0); err != nil {
			t.Fatal(err)
		}

		select {
		case <-time.After(3 * time.Second):
			t.Fatal("should have received a SIGQUIT")
		case <-sigCh:
			return
		}

		t.Fatal("should have received a SIGQUIT")
	})

}
