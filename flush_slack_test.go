package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSlack(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := ioutil.ReadAll(r.Body)
		// TODO: validate that p unmarshals in a template struct
		// TODO: also validate all fields
		w.Write(p)
	}))

	t.Run("should send a post call to server", func(t *testing.T) {
		s := Slack{
			Url: s.URL,
		}

		var buf bytes.Buffer
		msg := &L9Event{
			Message:       "test",
			Reason:        "test reason",
			ReferenceKind: "Pod",
			ReferenceName: "pod name",
			Timestamp:     time.Now().Unix(),
			Pod:           map[string]interface{}{"key1": "val1"},
			Services:      []string{"service1"},
		}
		b, err := json.Marshal(msg)
		if err != nil {
			t.Fatal(err)
		}
		buf.Write(b)
		buf.Write([]byte("\n"))

		if err := s.Flush("", "", buf.Bytes()); err != nil {
			t.Fatal(err)
		}

		// TODO: add an assert
	})
}
