package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/proto"
	"github.com/fatih/color"
	"github.com/tidwall/btree"
)

type listView struct {
	taskTree     *btree.BTree
	taskList     []string // A flattened version of the Btree so that we can traverse through it easily for highlighting.
	currentTask  string
	currentIndex int
}

func (l *listView) display(m model) string {
	if len(l.taskList) == 0 {
		return "No current tasks"
	}

	return l.stringifyTasks()
}

func (l *listView) handleKey(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			if l.currentIndex+1 > len(l.taskList)-1 {
				return m, nil
			}

			l.currentIndex += 1
		case "up", "k":
			if l.currentIndex-1 < 0 {
				return m, nil
			}

			l.currentIndex -= 1
		}
	}

	return m, nil
}

func newListView() (*listView, error) {
	conn, err := cl.State.Connect()
	if err != nil {
		_ = tea.Quit()
		return nil, err
	}

	client := proto.NewTodoClient(conn)

	resp, err := client.ListTasks(context.Background(), &proto.ListTasksRequest{
		Offset:           0,
		Limit:            0,
		ExcludeCompleted: true,
	})
	if err != nil {
		_ = tea.Quit()
		return nil, err
	}

	listTree := btree.New(compareTaskNode)

	// This map contains nodes in the list tree for easy access to them while creating the tree.
	listTreeNodeMap := map[string]*taskNode{}

	for _, task := range resp.Tasks {
		insertNodeToBTree(task, resp.Tasks, *listTree, listTreeNodeMap)
	}

	taskList := []string{}

	flattenTree(listTree, &taskList)

	return &listView{
		taskTree:     listTree,
		taskList:     taskList,
		currentIndex: 0,
	}, nil
}

func flattenTree(tree *btree.BTree, list *[]string) {
	tree.Descend(nil, func(item any) bool {
		task := item.(*taskNode)
		*list = append(*list, task.task.Id)
		flattenTree(task.children, list)
		return true
	})
}

func insertNodeToBTree(task *proto.Task, taskMap map[string]*proto.Task, listTree btree.BTree, listTreeNodeMap map[string]*taskNode) {
	// If this node was already processed then skip it.
	_, exists := listTreeNodeMap[task.Id]
	if exists {
		return
	}

	// If it doesn't exist within the map, then just create a new node.
	newNode := newTaskNode(task)

	// Check if the node has a parent, if it doesn't just insert it into data structure and exit.
	if task.Parent == "" {
		listTree.Set(newNode)
		listTreeNodeMap[task.Id] = newNode
		return
	}

	// If it does we can add this node to the parent and call it a day.
	parent, exists := listTreeNodeMap[task.Parent]
	if exists {
		parent.children.Set(newNode)
		listTreeNodeMap[task.Id] = newNode
		return
	}

	// If it does have a parent but the parent isn't created yet then we can just do that first
	protoParentNode, exists := taskMap[task.Parent]
	if !exists {
		// If the parent doesn't even exist in the main list we can just skip out here.
		return
	}

	insertNodeToBTree(protoParentNode, taskMap, listTree, listTreeNodeMap)
	insertNodeToBTree(task, taskMap, listTree, listTreeNodeMap)
}

func (l *listView) stringifyTasks() string {
	// We opt to not use strings.Builder here because to do fancy styling we
	// sometimes want to remove strings and replacement them with others.
	// strings.Builder does not easily allow this, but keeping the strings in
	// a normal slice does.
	var sb []string

	// Used to track the very first node printed. We use this to round the corners properly
	// of the very first node and no others.
	firstNode := true

	l.taskTree.Descend(nil, func(item any) bool {
		taskNode := item.(*taskNode)

		// To space out the branches that don't relate to each other
		// before we start printing a new branch we print a spacer.
		if !firstNode {
			sb = append(sb, "┊\n")
		}

		// Recursively process each task.
		l.stringifyTaskTreeBranch(&sb, taskNode, 0, firstNode)

		firstNode = false

		return true
	})

	if len(sb) > 1 {
		// Lastly if this is the very last node round the corner off.
		lastNodePrinted := sb[len(sb)-1]
		sb[len(sb)-1] = strings.Replace(lastNodePrinted, "├", "└", 1)
	}

	return strings.Join(sb, "")
}

func stringifyTask(task *proto.Task, highlight bool) string {
	id := color.YellowString(task.Id)
	if task.State == proto.Task_COMPLETED {
		id = color.GreenString(task.Id)
	}

	faint := color.New(color.Faint).SprintFunc()
	underline := color.New(color.Underline).SprintFunc()

	title := task.Title

	if highlight {
		title = underline(color.BlueString(title))
	}

	if task.State == proto.Task_COMPLETED {
		id = faint(id)
		title = faint(title)
	}

	taskStr := fmt.Sprintf("[%s] %s", id, title)

	return taskStr
}

func (l *listView) stringifyTaskTreeBranch(sb *[]string, node *taskNode, lvl int, firstNode bool) {
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

	task := node.task
	children := node.children
	highlight := l.taskList[l.currentIndex] == task.Id

	taskString += stringifyTask(task, highlight) + "\n"

	*sb = append(*sb, taskString)

	if children.Len() == 0 {
		return
	}

	children.Descend(nil, func(item any) bool {
		childTaskNode := item.(*taskNode)
		l.stringifyTaskTreeBranch(sb, childTaskNode, lvl+1, false)
		return true
	})
}
