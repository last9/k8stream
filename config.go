package main

import (
	"github.com/last9/k8stream/io"
)

const (
	DEFAULT_RESYNC_INTERVAL = 120
)

type L9K8streamConfig struct {
	io.Config      `json:"config" validate:"required"`
	KubeConfig     string `json:"kubeconfig"`
	ResyncInterval int    `json:"resync_interval"`
}

func setDefaults(c *L9K8streamConfig) {
	if c.ResyncInterval == 0 {
		c.ResyncInterval = DEFAULT_RESYNC_INTERVAL
	}
}
