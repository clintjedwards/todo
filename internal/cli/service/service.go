package service

import (
	"github.com/spf13/cobra"
)

var CmdService = &cobra.Command{
	Use:   "service",
	Short: "Manages service related commands for Todo.",
}
