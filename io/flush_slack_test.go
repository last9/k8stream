package io

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSlack(t *testing.T) {
	expectedCount := 4
	actual := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := ioutil.ReadAll(r.Body)
		actual++
		// TODO: validate that p unmarshals in a template struct
		// TODO: also validate all fields
		log.Println(string(p))
		w.Write(p)
	}))

	t.Run("should send a post call to server", func(t *testing.T) {
		s := Slack{
			Url: s.URL,
		}

		var buf bytes.Buffer
		for ix := 0; ix < 4; ix++ {
			msg := &SourceMsg{
				Message:   fmt.Sprintf("%d", ix),
				Reason:    "test reason",
				Timestamp: time.Now().Unix(),
				Pod:       map[string]interface{}{"key1": "val1"},
				Services:  []string{"service1"},
			}
			b, err := json.Marshal(msg)
			if err != nil {
				t.Fatal(err)
			}
			buf.Write(b)
			buf.Write([]byte("\n"))
		}

		if err := s.Flush("", "", buf.Bytes()); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectedCount, actual)
	})
}
