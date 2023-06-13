package cli

import (
	"fmt"
	"strings"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/internal/cli/service"
	"github.com/clintjedwards/todo/internal/cli/task"
	"github.com/clintjedwards/todo/internal/cli/task/scheduled"
	"github.com/clintjedwards/todo/internal/config"
	"github.com/spf13/cobra"
)

var appVersion = "0.0.dev_000000"

// RootCmd is the base of the cli
var RootCmd = &cobra.Command{
	Use:   "todo",
	Short: "Yet another simple todo app",
	Long: `Yet another simple todo app

### Environment Variables supported:

	` + strings.Join(config.GetCLIEnvVars(), "\n"),
	Version: " ", // We leave this added but empty so that the rootcmd will supply the -v flag
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		cl.InitState(cmd)
	},
}

func init() {
	RootCmd.SetVersionTemplate(humanizeVersion(appVersion))
	RootCmd.AddCommand(service.CmdService)
	RootCmd.AddCommand(task.CmdTaskCreate)
	RootCmd.AddCommand(task.CmdTaskDelete)
	RootCmd.AddCommand(task.CmdTaskGet)
	RootCmd.AddCommand(task.CmdTaskList)
	RootCmd.AddCommand(task.CmdTaskComplete)
	RootCmd.AddCommand(task.CmdTaskUpdate)
	RootCmd.AddCommand(task.CmdTaskSchedule)
	RootCmd.AddCommand(scheduled.CmdScheduled)

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
