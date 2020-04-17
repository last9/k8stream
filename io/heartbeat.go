package io

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const defaultHeartbeatInterval = 30

func StartHeartbeat(uid, hook string, interval int) error {
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

	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	go func() {
		for {
			<-ticker.C
			q := u.Query()
			q.Set("uid", uid)
			u.RawQuery = q.Encode()

			resp, err := http.Get(u.String())
			if err != nil {
				log.Print("error while sending heartbeat: %w", err)
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				respBody, _ := ioutil.ReadAll(resp.Body)

				log.Printf("error while sending heartbeat: %d %s", resp.StatusCode, string(respBody))
			}
		}
	}()

	return nil
}
