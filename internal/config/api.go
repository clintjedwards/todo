package config

import (
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/kelseyhightower/envconfig"
)

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

// convertDurationFromHCL attempts to move the string value of a duration written in HCL to
// the real time.Duration type. This is needed due to advanced types like time.Duration being not handled particularly
// well during HCL parsing: https://github.com/hashicorp/hcl/issues/202
func (c *API) convertDurationFromHCL() {
	if c.Server != nil && c.Server.ShutdownTimeoutHCL != "" {
		c.Server.ShutdownTimeout = mustParseDuration(c.Server.ShutdownTimeoutHCL)
	}
}

// FromEnv parses environment variables into the config object based on envconfig name
func (c *API) FromEnv() error {
	err := envconfig.Process("todo", c)
	if err != nil {
		return err
	}

	return nil
}

// FromBytes attempts to parse a given HCL configuration.
func (c *API) FromBytes(content []byte) error {
	err := hclsimple.Decode("config.hcl", content, nil, c)
	if err != nil {
		return err
	}

	c.convertDurationFromHCL()

	return nil
}

func (c *API) FromFile(path string) error {
	err := hclsimple.DecodeFile(path, nil, c)
	if err != nil {
		return err
	}

	c.convertDurationFromHCL()

	return nil
}

// Get the final configuration for the server.
// This involves correctly finding and ordering different possible paths for the configuration file.
//
// 1) The function is intended to be called with paths gleaned from the -config flag
// 2) Then combine that with possible other config locations that the user might store a config file.
// 3) Then try to see if the user has set an envvar for the config file, which overrides
// all previous config file paths.
// 4) Finally, pass back whatever is deemed the final config path from that process.
//
// We then use that path data to find the config file and read it in via HCL parsers. Once that is finished
// we then take any configuration from the environment and superimpose that on top of the final config struct.
func InitAPIConfig(userDefinedPath string) (*API, error) {
	// First we initiate the default values for the config.
	config := DefaultAPIConfig()

	possibleConfigPaths := []string{userDefinedPath, "/etc/todo/todo.hcl"}

	path := searchFilePaths(possibleConfigPaths...)

	// envVars top all other entries so if its not empty we just insert it over the current path
	// regardless of if we found one.
	envPath := os.Getenv("TODO_CONFIG_PATH")
	if envPath != "" {
		path = envPath
	}

	if path != "" {
		err := config.FromFile(path)
		if err != nil {
			return nil, err
		}
	}

	err := config.FromEnv()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func PrintAPIEnvs() error {
	var config API
	err := envconfig.Usage("todo", &config)
	if err != nil {
		return err
	}
	fmt.Println("TODO_CONFIG_PATH")

	return nil
}
