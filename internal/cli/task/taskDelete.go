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
	Example: `$ todo task get 62arz`,
	RunE:    taskDelete,
	Args:    cobra.ExactArgs(1),
}

func taskDelete(_ *cobra.Command, args []string) error {
	id := args[0]

	cl.State.Fmt.Print("Deleting Task")
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
