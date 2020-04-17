package io

import "encoding/json"

type MemSink struct {
	uuid    string
	Records map[string][]byte
	OnFetch func(string)
	batch   string
}

func (m *MemSink) LoadConfig(_ json.RawMessage) error {
	return nil
}

func (m *MemSink) Flush(uuid, ident string, d []byte) error {
	defer m.OnFetch(uuid + "/" + ident)
	m.batch = ident
	m.uuid = uuid
	m.Records[ident] = d
	return nil
}
