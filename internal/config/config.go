// Config controls the overall configuration of the application.
//
// It is generated by first attempting to read a configuration file and then overwriting those values
// with anything found in environment variables. Environment variables always come last and have the highest priority.
// As per (https://12factor.net/config).
//
// All environment variables are prefixed with "TODO". Ex: TODO_DEBUG=true
//
// You can print out a current description of current environment variable configuration by using the cli command:
//
//	`todo service printenv`
//
// Note: Even though this package uses the envconfig package it is incorrect to use the 'default' struct tags as that
// will cause incorrect overwriting of user defined configurations.
//
// Note: Because of the idiosyncrasies of how hcl conversion works certain advanced types like `time.Duration` need to
// have a sister variable that we read in through hcl via another type and convert to the actual wanted type.
package config

import (
	"errors"
	"log"
	"os"
	"time"
)

func mustParseDuration(duration string) time.Duration {
	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		log.Fatalf("could not parse duration %q; %v", duration, err)
	}

	return parsedDuration
}

// searchFilePaths will search each path given in order for a file
//
//	and return the first path that exists.
func searchFilePaths(paths ...string) string {
	for _, path := range paths {
		if path == "" {
			continue
		}

		if stat, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			continue
		} else {
			if stat.IsDir() {
				continue
			}
			return path
		}
	}

	return ""
}