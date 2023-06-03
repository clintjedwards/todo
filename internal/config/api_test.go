package config

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestInitAPIConfigAgainstSampleOverwriteWithEnvs(t *testing.T) {
	_ = os.Setenv("TODO_SERVER__HOST", "localhost:8080")
	_ = os.Setenv("TODO_SERVER__SHUTDOWN_TIMEOUT", "15s")
	_ = os.Setenv("TODO_SERVER__TLS_CERT_PATH", "./test")
	_ = os.Setenv("TODO_SERVER__TLS_KEY_PATH", "./localhost.key")

	defer os.Unsetenv("TODO_SERVER__HOST")
	defer os.Unsetenv("TODO_SERVER__SHUTDOWN_TIMEOUT")
	defer os.Unsetenv("TODO_SERVER__TLS_CERT_PATH")
	defer os.Unsetenv("TODO_SERVER__TLS_KEY_PATH")

	config, err := InitAPIConfig("../cli/service/sampleConfig.hcl", true, false, false)
	if err != nil {
		t.Fatal(err)
	}

	expected := API{
		LogLevel:    "info",
		Development: &Development{},
		Server: &Server{
			Host:                "localhost:8080",
			ShutdownTimeout:     time.Second * 15,
			TLSCertPath:         "./test",
			TLSKeyPath:          "./localhost.key",
			StoragePath:         "/tmp/todo.db",
			StorageResultsLimit: 200,
		},
	}

	diff := cmp.Diff(expected, *config)
	if diff != "" {
		t.Errorf("result is different than expected(-want +got):\n%s", diff)
	}
}
