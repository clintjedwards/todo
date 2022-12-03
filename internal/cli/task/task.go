package task

import (
	"github.com/spf13/cobra"
)

var CmdTask = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:  `Manage tasks.`,
}
