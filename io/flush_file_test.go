package io

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestLoadConfig(t *testing.T) {
	errorConfigs := []struct {
		Name         string
		Path         string
		ErrorMessage string
		Setup        func(path string) error
	}{
		{
			Name:         "Invalid path",
			Path:         "/tmp/invalid-path",
			ErrorMessage: "no such file or directory",
			Setup: func(path string) error {
				// Remove path if it already exists
				if _, err := os.Stat(path); err == nil {
					if err := os.Remove(path); err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			Name:         "File Path",
			Path:         "/tmp/file-path",
			ErrorMessage: "not a directory",
			Setup: func(path string) error {
				// Create a file
				if _, err := os.Create(path); err != nil {
					t.Fatal(err)
				}
				return nil
			},
		},
		{
			Name:         "Read-only Directory",
			Path:         "/tmp/read-only-dir",
			ErrorMessage: "unable to access",
			Setup: func(path string) error {
				// Remove directory if it already exists
				// and create a read-only directory
				if _, err := os.Stat(path); err == nil {
					if err := os.Remove(path); err != nil {
						return err
					}
				}

				if err := os.Mkdir(path, 0400); err != nil {
					return err
				}

				return nil
			},
		},
	}

	validConfigs := []struct {
		Name  string
		Path  string
		Setup func(path string) error
	}{
		{
			Name: "Valid Directory",
			Path: "/tmp/valid-directory",
			Setup: func(path string) error {
				// Remove directory if exists
				// and create a new directory
				if _, err := os.Stat(path); err == nil {
					if err := os.Remove(path); err != nil {
						return err
					}
				}

				if err := os.Mkdir(path, 0700); err != nil {
					return err
				}

				return nil
			},
		},
	}

	t.Run("Validate", func(t *testing.T) {
		for _, test := range errorConfigs {
			t.Run(test.Name, func(t *testing.T) {
				path := test.Path

				if err := test.Setup(path); err != nil {
					t.Fatal(err)
				}

				rawConfig := json.RawMessage([]byte(fmt.Sprintf(
					`{ "file_sink_dir": "%s" }`, path,
				)))

				fs := &FileSink{}

				err := fs.LoadConfig(rawConfig)
				if err == nil {
					t.Fatal("Expects an error")
				}

				if !strings.Contains(err.Error(), test.ErrorMessage) {
					t.Fatalf("Error should contain the message: %s", test.ErrorMessage)
				}
			})
		}

		for _, test := range validConfigs {
			t.Run(test.Name, func(t *testing.T) {
				path := test.Path

				if err := test.Setup(path); err != nil {
					t.Fatal(err)
				}

				rawConfig := json.RawMessage([]byte(fmt.Sprintf(
					`{ "file_sink_dir": "%s" }`, path,
				)))

				fs := &FileSink{}

				if err := fs.LoadConfig(rawConfig); err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, path, fs.Dir)
			})
		}
	})
}
