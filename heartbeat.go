package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"syscall"
	"time"
)

const (
	defaultHeartbeatInterval = 30
	defaultHeartbeatTimeout  = 300
)

func StartHeartbeat(uid, hook string, interval, timeout int) error {
	if hook == "" {
		return nil
	}
	u, err := url.Parse(hook)
	if err != nil {
		return fmt.Errorf("invalid hearbeat hook: %w", err)
	}

	if interval == 0 {
		interval = defaultHeartbeatInterval
	}
	if timeout == 0 {
		timeout = defaultHeartbeatTimeout
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	go func() {
		for {
			<-ticker.C
			q := u.Query()
			q.Set("uid", uid)
			q.Set("version", VERSION)
			u.RawQuery = q.Encode()

			client := http.Client{
				Timeout: time.Duration(timeout) * time.Millisecond,
			}
			resp, err := client.Get(u.String())
			if err != nil {
				log.Print("error while sending heartbeat: %w", err)
				continue
			}

			if resp.StatusCode == http.StatusUpgradeRequired {
				syscall.Kill(syscall.Getpid(), syscall.SIGQUIT)
				return
			}
		}
	}()

	return nil
}
