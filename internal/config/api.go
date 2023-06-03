package config

import (
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/knadh/koanf/parsers/hcl"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type API struct {
	// Log level affects the entire application's logs including launched extensions.
	LogLevel string `koanf:"log_level"`

	Development *Development `koanf:"development"`
	Server      *Server      `koanf:"server"`
}

func DefaultAPIConfig() *API {
	return &API{
		LogLevel:    "info",
		Development: DefaultDevelopmentConfig(),
		Server:      DefaultServerConfig(),
	}
}

type Development struct {
	PrettyLogging   bool `koanf:"pretty_logging"`
	UseLocalhostTLS bool `koanf:"use_localhost_tls"`
}

func DefaultDevelopmentConfig() *Development {
	return &Development{
		PrettyLogging:   false,
		UseLocalhostTLS: false,
	}
}

func FullDevelopmentConfig() *Development {
	return &Development{
		PrettyLogging:   true,
		UseLocalhostTLS: true,
	}
}

// Server represents lower level HTTP/GRPC server settings.
type Server struct {
	// URL for the server to bind to. Ex: localhost:8080
	Host string `koanf:"host"`

	// How long the GRPC service should wait on in-progress connections before hard closing everything out.
	ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`

	// Path to the sqlite database.
	StoragePath string `koanf:"storage_path"`

	// The total amount of results the database will attempt to pass back when a limit is not explicitly given.
	StorageResultsLimit int `koanf:"storage_results_limit"`

	TLSCertPath string `koanf:"tls_cert_path"`
	TLSKeyPath  string `koanf:"tls_key_path"`
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

// Get the final configuration for the server.
// This involves correctly finding and ordering different possible paths for the configuration file:
//
//  1. The function is intended to be called with paths gleaned from the -config flag in the cli.
//  2. If the user does not use the -config path of the path does not exist,
//     then we default to a few hard coded config path locations.
//  3. Then try to see if the user has set an envvar for the config file, which overrides
//     all previous config file paths.
//  4. Finally, whatever configuration file path is found first is the processed.
//
// Whether or not we use the configuration file we then search the environment for all environment variables:
//   - Environment variables are loaded after the config file and therefore overwrite any conflicting keys.
//   - All configuration that goes into a configuration file can also be used as an environment variable.
func InitAPIConfig(userDefinedPath string, loadDefaults, validate, devMode bool) (*API, error) {
	var config *API

	// First we initiate the default values for the config.
	if loadDefaults {
		config = DefaultAPIConfig()
	}

	if devMode {
		config.Development = FullDevelopmentConfig()
	}

	possibleConfigPaths := []string{userDefinedPath, "/etc/todo/todo.hcl"}

	path := searchFilePaths(possibleConfigPaths...)

	// envVars top all other entries so if its not empty we just insert it over the current path
	// regardless of if we found one.
	envPath := os.Getenv("TODO_CONFIG_PATH")
	if envPath != "" {
		path = envPath
	}

	configParser := koanf.New(".")

	if path != "" {
		err := configParser.Load(file.Provider(path), hcl.Parser(true))
		if err != nil {
			return nil, err
		}
	}

	err := configParser.Load(env.Provider("TODO_", "__", func(s string) string {
		newStr := strings.TrimPrefix(s, "TODO_")
		newStr = strings.ToLower(newStr)
		return newStr
	}), nil)
	if err != nil {
		return nil, err
	}

	err = configParser.Unmarshal("", &config)
	if err != nil {
		return nil, err
	}

	if validate {
		err = config.validate()
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func (c *API) validate() error {
	return nil
}

func GetAPIEnvVars() []string {
	api := API{
		Development: &Development{},
		Server:      &Server{},
	}
	fields := structs.Fields(api)

	vars := getEnvVarsFromStruct("TODO_", fields)
	sort.Strings(vars)
	return vars
}
