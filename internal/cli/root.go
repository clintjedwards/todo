package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var appVersion = "0.0.dev_000000"

// RootCmd is the base of the cli
var RootCmd = &cobra.Command{
	Use:     "todo",
	Short:   "",
	Long:    ``,
	Version: " ", // We leave this added but empty so that the rootcmd will supply the -v flag
}

func init() {
	RootCmd.SetVersionTemplate(humanizeVersion(appVersion))
	RootCmd.AddCommand(nil)

	RootCmd.PersistentFlags().String("config", "", "configuration file path")
	RootCmd.PersistentFlags().Bool("no-color", false, "disable color output")
	RootCmd.PersistentFlags().String("host", "", "specify the URL of the server to communicate to")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}

func humanizeVersion(version string) string {
	semver, hash, err := strings.Cut(version, "_")
	if !err {
		return ""
	}
	return fmt.Sprintf("todo %s [%s]\n", semver, hash)
}
