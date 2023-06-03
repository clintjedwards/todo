package task

import (
	"context"
	"fmt"
	"strings"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/proto"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdTaskDelete = &cobra.Command{
	Use:     "delete <id>",
	Short:   "Delete a new task",
	Example: `$ todo delete 62arz`,
	RunE:    taskDelete,
	Args:    cobra.ExactArgs(1),
}

func init() {
	CmdTaskDelete.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func taskDelete(cmd *cobra.Command, args []string) error {
	id := args[0]

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not delete tasks: %v", err))
		cl.State.Fmt.Finish()
		return err
	}

	cl.State.Fmt.Print("Deleting Task")
	if !force {
		cl.State.Fmt.Finish()

		var input string

		for {
			fmt.Printf("%s\n", color.YellowString("[Caution] Deleting a task will also delete all it's children permanently."))
			fmt.Print("Please type the ID of the task to confirm: ")
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

	resp, err := client.DeleteTask(context.Background(), &proto.DeleteTaskRequest{
		Id: id,
	})
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not delete task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}
	cl.State.Fmt.PrintSuccess(fmt.Sprintf("Deleted tasks: %q", resp.Ids))
	cl.State.Fmt.Finish()
	return nil
}
