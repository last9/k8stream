package io

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/go-playground/validator.v9"
)

func ReadConfig(f *os.File) (json.RawMessage, error) {
	return ioutil.ReadAll(f)
}

func loadEnvConfig(key string, cfg interface{}) error {
	return envconfig.Process(key, cfg)
}

type Config struct {
	Debug             bool            `json:"debug"`
	UID               string          `json:"uid" validate:"required"`
	BatchSize         int             `json:"batch_size"`
	BatchInterval     int             `json:"batch_interval"`
	Sink              string          `json:"sink" validate:"required"`
	Raw               json.RawMessage `json:"-"`
	HeartbeatHook     string          `json:"heartbeat_hook"`
	HeartbeatInterval int             `json:"heartbeat_interval"`
	HeartbeatTimeout  int             `json:"heartbeat_timeout_ms"`
}

func (c Config) Log(msg string, args ...interface{}) {
	if !c.Debug {
		return
	}

	log.Printf(msg, args...)
}

var validate *validator.Validate
var vOnce sync.Once

func Validator() *validator.Validate {
	vOnce.Do(func() {
		validate = validator.New()
	})

	return validate
}

func LoadConfig(b json.RawMessage, i interface{}) error {
	v := Validator()
	if err := json.Unmarshal(b, i); err != nil {
		return err
	}

	if err := v.Struct(i); err != nil {
		return err
	}

	if ic, ok := i.(*Config); ok {
		ic.Raw = b
	}

	return nil
}
