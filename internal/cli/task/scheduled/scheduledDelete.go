package scheduled

import (
	"context"
	"fmt"
	"strings"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/proto"
	"github.com/spf13/cobra"
)

var CmdScheduledTaskDelete = &cobra.Command{
	Use:     "delete <id>",
	Short:   "Delete a new scheduled task",
	Example: `$ todo scheduled delete 62arz`,
	RunE:    scheduledtaskDelete,
	Args:    cobra.ExactArgs(1),
}

func init() {
	CmdScheduled.AddCommand(CmdScheduledTaskDelete)
	CmdScheduledTaskDelete.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func scheduledtaskDelete(cmd *cobra.Command, args []string) error {
	id := args[0]

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not delete scheduled tasks: %v", err))
		cl.State.Fmt.Finish()
		return err
	}

	cl.State.Fmt.Print("Deleting Scheduled Task")
	if !force {
		cl.State.Fmt.Finish()

		var input string

		for {
			fmt.Print("Please type the ID of the scheduled task to confirm: ")
			fmt.Scanln(&input)
			if strings.EqualFold(input, id) {
				break
			}
		}
	}

	cl.State.NewFormatter()

	conn, err := cl.State.Connect()
	if err != nil {
		cl.State.Fmt.PrintErr(err)
		cl.State.Fmt.Finish()
		return err
	}

	client := proto.NewTodoClient(conn)

	resp, err := client.DeleteScheduledTask(context.Background(), &proto.DeleteScheduledTaskRequest{
		Id: id,
	})
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not delete scheduled task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}
	cl.State.Fmt.PrintSuccess(fmt.Sprintf("Deleted scheduled task: %q", resp.Id))
	cl.State.Fmt.Finish()
	return nil
}
