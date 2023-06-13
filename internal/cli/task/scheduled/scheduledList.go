package scheduled

import (
	"context"
	"fmt"
	"strings"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/proto"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var CmdScheduledTaskList = &cobra.Command{
	Use:     "list",
	Short:   "List all scheduled tasks",
	Example: `$ todo scheduled list`,
	RunE:    scheduledList,
}

func init() {
	CmdScheduled.AddCommand(CmdScheduledTaskList)
}

func scheduledList(_ *cobra.Command, _ []string) error {
	cl.State.Fmt.Print("Collecting Tasks")

	conn, err := cl.State.Connect()
	if err != nil {
		cl.State.Fmt.PrintErr(err)
		cl.State.Fmt.Finish()
		return err
	}

	client := proto.NewTodoClient(conn)

	resp, err := client.ListScheduledTasks(context.Background(), &proto.ListScheduledTasksRequest{
		Offset: 0,
		Limit:  0,
	})
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not list task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}
	cl.State.Fmt.Finish()

	data := [][]string{}
	for _, task := range resp.ScheduledTasks {
		data = append(data, []string{
			task.Id, task.Title, task.Expression,
		})
	}

	cl.State.Fmt.Println(formatTable(data, !cl.State.Config.NoColor))
	cl.State.Fmt.Finish()

	return nil
}

func formatTable(data [][]string, color bool) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetHeader([]string{"ID", "Title", "Expression"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(true)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(false)
	table.SetRowSeparator("―")
	table.SetRowLine(false)
	table.SetColumnSeparator("")
	table.SetCenterSeparator("")

	if color {
		table.SetHeaderColor(
			tablewriter.Color(tablewriter.FgBlueColor),
			tablewriter.Color(tablewriter.FgBlueColor),
			tablewriter.Color(tablewriter.FgBlueColor),
		)
		table.SetColumnColor(
			tablewriter.Color(tablewriter.FgYellowColor),
			tablewriter.Color(0),
			tablewriter.Color(0),
		)
	}

	table.AppendBulk(data)

	table.Render()
	return tableString.String()
}
