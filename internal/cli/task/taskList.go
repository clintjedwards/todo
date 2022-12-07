package task

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/proto"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var cmdTaskList = &cobra.Command{
	Use:     "list",
	Short:   "List all tasks",
	Example: `$ todo task list`,
	RunE:    taskList,
}

func init() {
	cmdTaskList.Flags().BoolP("show-completed", "c", false, "Show tasks which have been completed")
	CmdTask.AddCommand(cmdTaskList)
}

func taskList(cmd *cobra.Command, _ []string) error {
	cl.State.Fmt.Print("Collecting Tasks")

	showCompleted, err := cmd.Flags().GetBool("show-completed")
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not list tasks: %v", err))
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

	resp, err := client.ListTasks(context.Background(), &proto.ListTasksRequest{
		Offset:           0,
		Limit:            0,
		ExcludeCompleted: !showCompleted,
	})
	if err != nil {
		cl.State.Fmt.PrintErr(fmt.Sprintf("could not list task: %v", err))
		cl.State.Fmt.Finish()
		return err
	}
	cl.State.Fmt.Finish()

	fmt.Println(stringifyTasks(resp.Tasks))

	return nil
}

type taskNode struct {
	task     *proto.Task
	children map[string]struct{}
}

// Returns a mapping of task nodes to their children, if any.
// This allows us the ability to stringify the tasks.
// Also returns the top-lvl keys in alphabetically sorted order by title.
func toTaskTree(tasks []*proto.Task) (tree map[string]taskNode, keys []string) {
	taskMap := map[string]taskNode{}
	keys = []string{}

	for _, task := range tasks {
		_, exists := taskMap[task.Id]
		if !exists {
			taskMap[task.Id] = taskNode{
				task:     task,
				children: map[string]struct{}{},
			}
			keys = append(keys, task.Id)
		}

		if task.Parent == "" {
			continue
		}

		// Does the parent exist?
		parentTaskNode, exists := taskMap[task.Parent]
		if !exists {
			keys = append(keys, task.Parent)

			// If it doesn't we want to declare a new child map.
			parentTaskNode = taskNode{
				task:     task,
				children: map[string]struct{}{},
			}
		}

		// Add this task as a child of the mentioned parent.
		parentTaskNode.children[task.Id] = struct{}{}

		// Finally, add the updated child map back to the parent.
		taskMap[task.Parent] = parentTaskNode
	}

	sort.Strings(keys)
	return taskMap, keys
}

func stringifyTask(task *proto.Task) string {
	taskStr := fmt.Sprintf("[%s] %s", color.MagentaString(task.Id), color.BlueString(task.Title))

	faint := color.New(color.Faint).SprintfFunc()
	if task.State == proto.Task_COMPLETED {
		taskStr = faint("%s", taskStr)
	}

	return taskStr
}

func stringifyTasks(tasks []*proto.Task) string {
	taskTree, keys := toTaskTree(tasks)

	// We opt to not use strings.Builder here because to do fancy styling we
	// sometimes want to remove strings and replacement them with others.
	// strings.Builder does not easily allow this, but keeping the strings in
	// a normal slice does.
	var sb []string

	// Used to track the very first node printed. We use this to round the corners properly
	// of the very first node and no others.
	firstNode := true

	for index, taskID := range keys {
		task := taskTree[taskID].task

		// Skip any child nodes since we want to start building the strings only on the parents.
		if task.Parent != "" {
			continue
		}

		// Recursively process all children of that top level task.
		stringifyTaskTreeBranch(&sb, taskTree, taskID, 0, firstNode)

		if firstNode {
			firstNode = false
		}

		// When we're done printing a branch get some space in-between before printing another.
		if index != len(keys)-1 {
			sb = append(sb, "┊\n")
		}
	}

	if len(sb) > 1 {
		// Lastly if this is the very last node, go ahead and round the corner off.
		lastNodePrinted := sb[len(sb)-1]
		sb[len(sb)-1] = strings.Replace(lastNodePrinted, "├", "└", 1)
	}

	return strings.Join(sb, "")
}

func stringifyTaskTreeBranch(sb *[]string, taskTree map[string]taskNode, id string, lvl int, firstNode bool) {
	taskString := ""

	if firstNode {
		taskString += "┌─"
	} else {
		taskString += "├─"
	}

	if lvl > 0 {
		taskString += strings.Repeat("─", lvl)
	}
	taskString += " "

	task := taskTree[id].task
	children := taskTree[id].children
	taskString += stringifyTask(task) + "\n"

	*sb = append(*sb, taskString)

	if len(children) == 0 {
		return
	}

	for childID := range children {
		stringifyTaskTreeBranch(sb, taskTree, childID, lvl+1, false)
	}
}
