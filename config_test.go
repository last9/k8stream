package main

import (
	"os"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestConfig(t *testing.T) {
	f, err := os.Open("testdata/sample-config.json")
	if err != nil {
		t.Fatal(f)
	}

	b, err := readConfig(f)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Validate", func(t *testing.T) {
		t.Run("L9K8streamConfig", func(t *testing.T) {
			s := &L9K8streamConfig{}
			if err := loadConfig(b, s); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("S3Sink", func(t *testing.T) {
			s := &S3Sink{}
			if err := loadConfig(b, s); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, s.Prefix, "local/test-upload")
			assert.Equal(t, s.Profile, "last9data")
		})

		t.Run("FileSink", func(t *testing.T) {
			s := &FileSink{}
			if err := loadConfig(b, s); err != nil {
				t.Fatal(err)
			}
		})
	})
}
