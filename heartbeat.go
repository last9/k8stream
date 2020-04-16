package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const defaultHeartbeatInterval = 30
const defaultHeartbeatTimeout = 300 // milliseconds

func StartHeartbeat(uid, version, hook string, interval, timeout int) <-chan error {
	errCh := make(chan error)

	if hook == "" {
		errCh <- fmt.Errorf("empty heartbeat hook")
	}

	u, err := url.Parse(hook)
	if err != nil {
		errCh <- fmt.Errorf("invalid hearbeat hook: %w", err)
	}
	q := u.Query()
	q.Set("uid", uid)
	q.Set("version", version)
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

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				respBody, _ := ioutil.ReadAll(resp.Body)
				log.Printf("error while sending heartbeat: %d %s", resp.StatusCode, string(respBody))

				if resp.StatusCode == http.StatusUpgradeRequired {
					log.Println("upgrade required for k8stream.")
					errCh <- fmt.Errorf("upgrade required for k8stream")
				}
			}
		}
	}()

	return errCh
}
