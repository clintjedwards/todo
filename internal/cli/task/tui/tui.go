package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/proto"
	"github.com/spf13/cobra"
	"github.com/tidwall/btree"
)

var CmdTaskTUI = &cobra.Command{
	Use:     "tui",
	Short:   "Run terminal interface",
	Example: `$ todo tui`,
	RunE:    taskTUI,
}

type taskNode struct {
	task     *proto.Task
	children *btree.BTree
}

func compareTaskNode(a, b any) bool {
	task1, task2 := a.(*taskNode), b.(*taskNode)
	return task1.task.Id < task2.task.Id
}

func newTaskNode(task *proto.Task) *taskNode {
	return &taskNode{
		task:     task,
		children: btree.New(compareTaskNode),
	}
}

type view interface {
	display(m model) string
	handleKey(m *model, msg tea.Msg) (tea.Model, tea.Cmd)
}

type model struct {
	currentView  view
	previousView view
}

func newModel() (*model, error) {
	newListView, err := newListView()
	if err != nil {
		panic(err)
	}

	return &model{
		currentView:  newListView,
		previousView: newListView,
	}, nil
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "c":
			_, isCreateView := m.currentView.(*createView)
			if isCreateView {
				break
			}

			listView := m.currentView.(*listView)
			taskList := listView.taskList
			taskMap := map[string]string{}

			for _, id := range taskList {
				taskMap[id] = listView.taskTree.Get(id).(*taskNode).task.Title
			}

			newCreateView, _ := newCreateView(taskMap)
			m.previousView = m.currentView
			m.currentView = newCreateView
			return m, nil
		case "?":
			_, isHelpView := m.currentView.(*helpView)
			if isHelpView {
				savedView := m.currentView
				m.currentView = m.previousView
				m.previousView = savedView
				break
			}

			newHelpView, _ := newHelpView()
			m.previousView = m.currentView
			m.currentView = newHelpView
			return m, nil
		}
	}

	return m.currentView.handleKey(m, msg)
}

func (m *model) View() string {
	return m.currentView.display(*m)
}

func taskTUI(_ *cobra.Command, _ []string) error {
	cl.State.Fmt.Finish()

	newModel, err := newModel()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	p := tea.NewProgram(newModel)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}
