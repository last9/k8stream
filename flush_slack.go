package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/last9/k8stream/io"
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
	Url string `json:"slack_url", validate:"required"`
}

func (s *Slack) LoadConfig(b json.RawMessage) error {
	return io.LoadConfig(b, s)
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

		events.Messages = append(events.Messages, &obj)
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
	resp, err := http.Post(url, "application/json", bytes.NewReader(msg))
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Flushed to slack ", resp.StatusCode)
	return nil
}
