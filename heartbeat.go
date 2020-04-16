package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const defaultHeartbeatInterval = 30
const defaultHeartbeatTimeout = 300 // milliseconds

type interruptFn func(chan struct{}, chan struct{}, *sync.WaitGroup)

func StartHeartbeat(uid, hook string, interval, timeout int,
	stopCh chan struct{}, flusherStopCh chan struct{}, wg *sync.WaitGroup,
	fn interruptFn) error {
	if hook == "" {
		return fmt.Errorf("empty heartbeat hook")
	}

	u, err := url.Parse(hook)
	if err != nil {
		return fmt.Errorf("invalid hearbeat hook: %w", err)
	}
	q := u.Query()
	q.Set("uid", uid)
	q.Set("version", VERSION)
	u.RawQuery = q.Encode()

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
			client := http.Client{Timeout: time.Duration(timeout) * time.Millisecond}

			resp, err := client.Get(u.String())
			if err != nil {
				log.Print("error while sending heartbeat: %w", err)
			}

			if resp.StatusCode == http.StatusUpgradeRequired {
				fn(stopCh, flusherStopCh, wg)
			}
		}
	}()

	return nil
}
