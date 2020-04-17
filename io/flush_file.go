package io

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type FileSink struct {
	Dir string `json:"file_sink_dir" validate:"required"`
}

func (f *FileSink) LoadConfig(b json.RawMessage) error {
	// TODO: Check if Dir actually exists and is writable.
	return LoadConfig(b, f)
}

func (f *FileSink) Flush(uuid, filename string, d []byte) error {
	fname := filepath.Join(f.Dir, fmt.Sprintf("%v_%v.log", uuid, filename))
	return ioutil.WriteFile(fname, d, 0644)
}
