package task

import (
	"context"
	"fmt"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/proto"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdTaskCreate = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new task",
	Long:  `Create a new task.`,
	Example: `$ todo create "New Task"
$ todo create "New Task" --description="my new task"
`,
	RunE: taskCreate,
	Args: cobra.ExactArgs(1),
}

func init() {
	CmdTaskCreate.Flags().StringP("description", "d", "", "Description about task")
	CmdTaskCreate.Flags().StringP("parent", "p", "", "Link this task as the child of another task")
}

func taskCreate(cmd *cobra.Command, args []string) error {
	title := args[0]

	cl.State.Fmt.Print("Creating Task")

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not create task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}

	parent, err := cmd.Flags().GetString("parent")
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not create task: %v", err))
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

	resp, err := client.CreateTask(context.Background(), &proto.CreateTaskRequest{
		Title:       title,
		Description: description,
		Parent:      parent,
	})
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not create task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}
	cl.State.Fmt.PrintSuccess(fmt.Sprintf("Created task: [%s] %s", color.MagentaString(resp.Id), "\""+color.BlueString(title)+"\""))
	cl.State.Fmt.Finish()
	return nil
}
