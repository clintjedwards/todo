package task

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/internal/cli/format"
	"github.com/clintjedwards/todo/proto"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdTaskGet = &cobra.Command{
	Use:     "get <id>",
	Short:   "Describe a task",
	Example: `$ todo 62arz`,
	RunE:    taskGet,
	Args:    cobra.ExactArgs(1),
}

func taskGet(_ *cobra.Command, args []string) error {
	id := args[0]

	cl.State.Fmt.Print("Getting Task Details")

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

	cl.State.Fmt.Println(formatTaskInfo(resp.Task))
	cl.State.Fmt.Finish()
	return nil
}

type data struct {
	ID          string
	Title       string
	Description string
	State       string
	Created     string
	Modified    string
	Parent      string
}

func formatTaskInfo(task *proto.Task) string {
	data := data{
		ID:          color.MagentaString(task.Id),
		Title:       color.BlueString(task.Title),
		Description: task.Description,
		State: format.ColorizeTaskState(
			format.NormalizeEnumValue(task.State.String(), "Unknown"),
		),
		Created:  format.UnixMilli(task.Created, "Unknown", cl.State.Config.Detail),
		Modified: format.UnixMilli(task.Modified, "Unknown", cl.State.Config.Detail),
		Parent:   task.Parent,
	}

	const formatTmpl = `Task [{{.ID}}] :: {{.Title}} :: {{.State}}

  {{if .Description}}{{.Description}}{{- end}}

Created {{.Created}}`

	var tpl bytes.Buffer
	t := template.Must(template.New("tmp").Parse(formatTmpl))
	_ = t.Execute(&tpl, data)
	return tpl.String()
}
