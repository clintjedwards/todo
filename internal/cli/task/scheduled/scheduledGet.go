package scheduled

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/proto"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdScheduledTaskGet = &cobra.Command{
	Use:     "get <id>",
	Short:   "Describe a scheduled task",
	Example: `$ todo scheduled get 62arz`,
	RunE:    scheduledtaskGet,
	Args:    cobra.ExactArgs(1),
}

func init() {
	CmdScheduled.AddCommand(CmdScheduledTaskGet)
}

func scheduledtaskGet(_ *cobra.Command, args []string) error {
	id := args[0]

	cl.State.Fmt.Print("Getting Scheduled Task Details")

	conn, err := cl.State.Connect()
	if err != nil {
		cl.State.Fmt.PrintErr(err)
		cl.State.Fmt.Finish()
		return err
	}

	client := proto.NewTodoClient(conn)

	resp, err := client.GetScheduledTask(context.Background(), &proto.GetScheduledTaskRequest{
		Id: id,
	})
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not get scheduled task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}

	cl.State.Fmt.Println(formatScheduledTaskInfo(resp.ScheduledTask))
	cl.State.Fmt.Finish()
	return nil
}

type data struct {
	ID          string
	Title       string
	Description string
	Parent      string
	Expression  string
}

func formatScheduledTaskInfo(scheduledtask *proto.ScheduledTask) string {
	data := data{
		ID:          color.MagentaString(scheduledtask.Id),
		Title:       color.BlueString(scheduledtask.Title),
		Description: scheduledtask.Description,
		Parent:      scheduledtask.Parent,
		Expression:  scheduledtask.Expression,
	}

	const formatTmpl = `ScheduledTask [{{.ID}}] :: {{.Title}} :: {{.Expression}}

  {{if .Description}}{{.Description}}{{- end}}

{{if .Parent}}Parent: {{.Parent}}{{- end}}`

	var tpl bytes.Buffer
	t := template.Must(template.New("tmp").Parse(formatTmpl))
	_ = t.Execute(&tpl, data)
	return tpl.String()
}
