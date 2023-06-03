package task

import (
	"context"
	"fmt"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/proto"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdTaskSchedule = &cobra.Command{
	Use:   "schedule <title> <expression>",
	Short: "Schedule a new task",
	Long: `Todo allows you to schedule a task that reoccurs. Useful for tasks that need to be done on some sort of schedule.
For example, if I need to re-lube my bike chain every month I can schedule a task with the expression: "0 0 1 * * *".

The syntax for the expression is a subset of the cron syntax; you can find more about it here: https://github.com/clintjedwards/avail.

Scheduled tasks will automatically be created for you on the timeline that you set.`,
	Example: `$ todo schedule "New Task" "0 0 1 * * *"
$ todo schedule "New Task" "* * * * * *" --description="my new task"
`,
	RunE: taskSchedule,
	Args: cobra.ExactArgs(2),
}

func init() {
	CmdTaskSchedule.Flags().StringP("description", "d", "", "Description about task")
	CmdTaskSchedule.Flags().StringP("parent", "p", "", "Link this task as the child of another task")
}

func taskSchedule(cmd *cobra.Command, args []string) error {
	title := args[0]
	expression := args[1]

	cl.State.Fmt.Print("Creating Task")

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not schedule task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}

	parent, err := cmd.Flags().GetString("parent")
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not schedule task: %v", err))
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

	resp, err := client.CreateScheduledTask(context.Background(), &proto.CreateScheduledTaskRequest{
		Title:       title,
		Description: description,
		Parent:      parent,
		Expression:  expression,
	})
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not schedule task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}
	cl.State.Fmt.PrintSuccess(fmt.Sprintf("Scheduled task: [%s] %s", color.MagentaString(resp.Id), "\""+color.BlueString(title)+"\""))
	cl.State.Fmt.Finish()
	return nil
}
