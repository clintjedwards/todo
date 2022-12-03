package service

import (
	"fmt"
	"os"

	_ "embed"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/spf13/cobra"
)

var cmdServiceInitConfig = &cobra.Command{
	Use:   "init-config",
	Short: "Create example todo config file.",
	Long: `Create example todo config file.

This file can be used as a example starting point and be customized further. This file should not
be used to run production versions of Todo as it is inherently insecure.

The default filename is example.todo.hcl, but can be renamed via flags.`,
	Example: `$ todo service init-config
$ todo service init-config -f myServer.hcl`,
	RunE: serviceInitConfig,
}

//go:embed sampleAPIConfig.hcl
var content string

func init() {
	cmdServiceInitConfig.Flags().StringP("filepath", "f", "./example.todo.hcl", "path to file")
	CmdService.AddCommand(cmdServiceInitConfig)
}

func serviceInitConfig(cmd *cobra.Command, _ []string) error {
	filepath, _ := cmd.Flags().GetString("filepath")

	cl.State.Fmt.Print("Creating service config file")

	err := createServiceConfigFile(filepath)
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not create service config file: %v", err))
		cl.State.Fmt.Finish()
		return err
	}

	cl.State.Fmt.PrintSuccess(fmt.Sprintf("Created service config file: %s", filepath))
	cl.State.Fmt.Finish()
	return nil
}

func createServiceConfigFile(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}
