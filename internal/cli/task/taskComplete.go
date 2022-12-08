package task

import (
	"context"
	"fmt"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/proto"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdTaskComplete = &cobra.Command{
	Use:     "complete <id>",
	Short:   "Mark a task as complete",
	Example: `$ todo complete 62arz`,
	RunE:    taskComplete,
	Args:    cobra.ExactArgs(1),
}

func taskComplete(_ *cobra.Command, args []string) error {
	id := args[0]

	cl.State.Fmt.Print("Completing Task")

	conn, err := cl.State.Connect()
	if err != nil {
		cl.State.Fmt.PrintErr(err)
		cl.State.Fmt.Finish()
		return err
	}

	client := proto.NewTodoClient(conn)

	resp, err := client.GetTask(context.Background(), &proto.GetTaskRequest{
		Id: id,
	})
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not get task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}

	_, err = client.UpdateTask(context.Background(), &proto.UpdateTaskRequest{
		Id:          id,
		Title:       resp.Task.Title,
		Description: resp.Task.Description,
		Parent:      resp.Task.Parent,
		State:       proto.UpdateTaskRequest_COMPLETED,
	})
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not complete task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}
	cl.State.Fmt.PrintSuccess(fmt.Sprintf("Completed task: %s", color.MagentaString(id)))
	cl.State.Fmt.Finish()
	return nil
}
