package io

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	f, err := os.Open("testdata/sample-config.json")
	if err != nil {
		t.Fatal(f)
	}

	b, err := ReadConfig(f)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Validate", func(t *testing.T) {
		t.Run("Config", func(t *testing.T) {
			s := &Config{}
			if err := LoadConfig(b, s); err != nil {
				t.Fatal(err)
			}

			assert.NotEmpty(t, s.Raw)
		})

		t.Run("S3Sink", func(t *testing.T) {
			s := &S3Sink{}
			if err := LoadConfig(b, s); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, s.Prefix, "local/test-upload")
			assert.Equal(t, s.Profile, "last9data")
		})

		t.Run("FileSink", func(t *testing.T) {
			s := &FileSink{}
			if err := LoadConfig(b, s); err != nil {
				t.Fatal(err)
			}
		})
	})
}
