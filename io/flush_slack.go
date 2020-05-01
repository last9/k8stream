package io

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Slack struct {
	Url string `json:"slack_url", validate:"required"`
}

func (s *Slack) LoadConfig(b json.RawMessage) error {
	return LoadConfig(b, s)
}

func (s *Slack) Flush(uuid, ident string, msg []byte) error {
	scanner := bufio.NewScanner(bytes.NewBuffer(msg))

	// read each line: which is a json object
	for scanner.Scan() {
		var obj SourceMsg
		if err := json.Unmarshal(scanner.Bytes(), &obj); err != nil {
			return err
		}

		t, err := NewTemplate(&obj)
		if err != nil {
			return err
		}

		b, err := json.Marshal(t)
		if err != nil {
			return err
		}

		// call slack webhook
		if err := PostMessage(s.Url, b); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func PostMessage(url string, msg []byte) error {
	if _, err := http.Post(url, "application/json", bytes.NewReader(msg)); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

type SourceMsg struct {
	ID                 string                 `json:"id"`
	Timestamp          int64                  `json:"timestamp"`
	Message            string                 `json:"message"`
	Namespace          string                 `json:"namespace"`
	Reason             string                 `json:"reason"`
	ReferenceUID       string                 `json:"reference_uid"`
	ReferenceNamespace string                 `json:"reference_namespace"`
	ReferenceName      string                 `json:"reference_name"`
	ReferenceKind      string                 `json:"reference_kind"`
	ReferenceVersion   string                 `json:"reference_version"`
	ObjectUid          string                 `json:"object_uid"`
	Labels             map[string]string      `json:"labels"`
	Annotations        map[string]string      `json:"annotations"`
	Address            []string               `json:"address"`
	Pod                map[string]interface{} `json:"pod"`
	Services           []string               `json:"services"`
}

// Template for the message format that is sent on slack.
type Tmpl struct {
	Timestamp        string          `json:"timestamp"`
	EventReason      string          `json:"event_reason"`
	AffectedServices []string        `json:"affected_services"`
	ObjectDetails    json.RawMessage `json:"object_details"`
	Details          json.RawMessage `json:"details"`
}

func NewTemplate(obj *SourceMsg) (*Tmpl, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	p, err := json.Marshal(obj.Pod)
	if err != nil {
		return nil, err
	}

	return &Tmpl{
		Details:          b,
		ObjectDetails:    p,
		Timestamp:        time.Unix(obj.Timestamp, 0).Format(time.RFC3339),
		AffectedServices: obj.Services,
		EventReason:      obj.Reason,
	}, nil
}
