package config

import (
	"testing"

	"github.com/fatih/structs"
)

// Simply test for panics, the reflect code here will panic if the API struct has any
// pointers with zero values.
func TestGetEnvvarsFromStruct(t *testing.T) {
	api := API{
		Development: &Development{},
		Server:      &Server{},
	}
	fields := structs.Fields(api)
	getEnvVarsFromStruct("TODO_", fields)
}
