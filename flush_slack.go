package main

import "C"
import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/last9/k8stream/io"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

var tmpl = func() *template.Template {
	return template.Must(template.New("slackTmpl").Funcs(
		template.FuncMap{
			"ifComma": func(ix, length int) bool {
				return ix < (length - 1)
			},
			"joinStr": strings.Join,
			"formatTime": func(t int64) string {
				return time.Unix(t, 0).Format(time.RFC850)
			},
		}).Parse(slackTmpl))
}()

type Slack struct {
	Url        string   `json:"slack_url", validate:"required"`
	Namespaces []string `json:"namespaces"`
	Events     []string `json:"events"`
}

func (s *Slack) LoadConfig(b json.RawMessage) error {
	return io.LoadConfig(b, s)
}

func (s *Slack) isEligible(obj *L9Event) bool {
	found := len(s.Namespaces) == 0
	for _, n := range s.Namespaces {
		if obj.Namespace == n || n == "*" {
			found = true
			break
		}
	}
	found = found && len(s.Events) == 0
	for _, n := range s.Events {
		if obj.Reason == n || n == "*" {
			found = true
			break
		}
	}

	return found
}

func (s *Slack) Flush(uuid, ident string, byts []byte) error {
	events := struct {
		Messages []*L9Event
	}{}

	scanner := bufio.NewScanner(bytes.NewBuffer(byts))

	// read each line: which is a json object
	for scanner.Scan() {
		var obj L9Event
		if err := json.Unmarshal(scanner.Bytes(), &obj); err != nil {
			return err
		}

		if s.isEligible(&obj) {
			events.Messages = append(events.Messages, &obj)
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, events); err != nil {
		return err
	}

	if err := PostMessage(s.Url, buf.Bytes()); err != nil {
		return err
	}

	return scanner.Err()
}

func PostMessage(url string, msg []byte) error {
	log.Println(url)
	fmt.Println(string(msg))
	if _, err := http.Post(url, "application/json", bytes.NewReader(msg)); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
