package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/go-playground/validator.v9"
)

func readConfig(f *os.File) (json.RawMessage, error) {
	return ioutil.ReadAll(f)
}

func loadEnvConfig(key string, cfg interface{}) error {
	return envconfig.Process(key, cfg)
}

type L9K8streamConfig struct {
	KubeConfig    string `json:"kubeconfig" validate:"required"`
	UID           string `json:"uid" validate:"required"`
	BatchSize     int    `json:"batch_size"`
	BatchInterval int    `json:"batch_interval"`
	Sink          string `json:"sink" validate:"required"`
}

var validate *validator.Validate
var vOnce sync.Once

func Validator() *validator.Validate {
	vOnce.Do(func() {
		validate = validator.New()
	})

	return validate
}

func loadConfig(b json.RawMessage, i interface{}) error {
	v := Validator()
	if err := json.Unmarshal(b, i); err != nil {
		return err
	}

	if err := v.Struct(i); err != nil {
		return err
	}

	return nil
}
