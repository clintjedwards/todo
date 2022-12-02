package config

import "time"

type API struct {
	// DevMode turns on humanized debug messages, extra debug logging for the web server and other
	// convenient features for development. Usually turned on along side LogLevel=debug.
	DevMode bool `hcl:"dev_mode,optional"`
	// Log level affects the entire application's logs including launched triggers.
	LogLevel string  `split_words:"true" hcl:"log_level,optional"`
	Server   *Server `hcl:"server,block"`
}

func DefaultAPIConfig() *API {
	return &API{
		DevMode:  false,
		LogLevel: "debug",
		Server:   DefaultServerConfig(),
	}
}

// Server represents lower level HTTP/GRPC server settings.
type Server struct {
	// URL for the server to bind to. Ex: localhost:8080
	Host string `hcl:"host,optional"`

	// How long the GRPC service should wait on in-progress connections before hard closing everything out.
	ShutdownTimeout time.Duration `split_words:"true"`

	// ShutdownTimeoutHCL is the HCL compatible counter part to ShutdownTimeout. It allows the parsing of a string
	// to a time.Duration since HCL does not support parsing directly into a time.Duration.
	ShutdownTimeoutHCL string `ignored:"true" hcl:"shutdown_timeout,optional"`

	// Path to the sqlite database.
	StoragePath string `split_words:"true" hcl:"storage_path,optional"`

	// The total amount of results the database will attempt to pass back when a limit is not explicitly given.
	StorageResultsLimit int `split_words:"true" hcl:"storage_results_limit,optional"`

	TLSCertPath string `split_words:"true" hcl:"tls_cert_path,optional"`
	TLSKeyPath  string `split_words:"true" hcl:"tls_key_path,optional"`
}

// DefaultServerConfig returns a pre-populated configuration struct that is used as the base for super imposing user configuration
// settings.
func DefaultServerConfig() *Server {
	return &Server{
		Host:                "localhost:8080",
		StoragePath:         "/tmp/todo.db",
		StorageResultsLimit: 200,
	}
}
