package task

import (
	"context"
	"fmt"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/proto"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdTaskUpdate = &cobra.Command{
	Use:     "update <id>",
	Short:   "Update the details of a task",
	Example: `$ todo update 62arz -d "example description"`,
	RunE:    taskUpdate,
	Args:    cobra.ExactArgs(1),
}

func init() {
	CmdTaskUpdate.Flags().StringP("description", "d", "", "Description about task")
	CmdTaskUpdate.Flags().StringP("parent", "p", "", "Link this task as the child of another task")
	CmdTaskUpdate.Flags().StringP("title", "t", "", "Task title")
	CmdTaskUpdate.Flags().StringP("state", "s", "", "Manipulate task state")
}

func taskUpdate(cmd *cobra.Command, args []string) error {
	id := args[0]

	cl.State.Fmt.Print("Updating Task")

	description := ""
	parent := ""
	title := ""
	state := ""

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not update task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}

	parent, err = cmd.Flags().GetString("parent")
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not update task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}

	title, err = cmd.Flags().GetString("title")
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not update task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}

	state, err = cmd.Flags().GetString("state")
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not update task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}

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

	if description == "" {
		description = resp.Task.Description
	}

	if title == "" {
		title = resp.Task.Title
	}

	if parent == "" {
		parent = resp.Task.Parent
	}

	if state == "" {
		state = resp.Task.State.String()
	}

	_, err = client.UpdateTask(context.Background(), &proto.UpdateTaskRequest{
		Id:          id,
		Title:       title,
		Description: description,
		Parent:      parent,
		State:       proto.UpdateTaskRequest_TaskState(proto.UpdateTaskRequest_TaskState_value[state]),
	})
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not update task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}
	cl.State.Fmt.PrintSuccess(fmt.Sprintf("Updated task: %s", color.MagentaString(id)))
	cl.State.Fmt.Finish()
	return nil
}
