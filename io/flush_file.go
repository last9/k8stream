package io

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

type FileSink struct {
	Dir string `json:"file_sink_dir" validate:"required"`
}

func (f *FileSink) LoadConfig(b json.RawMessage) error {
	if err := LoadConfig(b, f); err != nil {
		return err
	}

	fileInfo, err := os.Stat(f.Dir)
	if os.IsNotExist(err) {
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", f.Dir)
	}

	if err = unix.Access(f.Dir, unix.W_OK); err != nil {
		return fmt.Errorf("unable to access %s (%s)", f.Dir, err.Error())
	}

	return nil
}

func (f *FileSink) Flush(uuid, filename string, d []byte) error {
	fname := filepath.Join(f.Dir, fmt.Sprintf("%v_%v.log", uuid, filename))
	return ioutil.WriteFile(fname, d, 0644)
}
