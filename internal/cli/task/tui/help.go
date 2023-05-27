package tui

import (
	"fmt"
	"strings"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpView struct{}

func (h *helpView) display(m model) string {
	symbolColor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D3D3D3"))

	descColor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#808080"))

	menuTable := [][]string{
		{
			symbolColor.Render("↑/k"), descColor.Render("move up"),
		},
		{
			symbolColor.Render("↓/j"), descColor.Render("move down"),
		},
		{},
		{
			symbolColor.Render("?"), descColor.Render("toggle help"),
		},
		{
			symbolColor.Render("q/ctrl+c"), descColor.Render("quit"),
		},
	}

	helpMenu := "  Help\n\n"
	helpMenu += generateGenericTable(menuTable, " ", 4)

	return helpMenu
}

func (h *helpView) handleKey(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "?":
			listView, err := newListView()
			if err != nil {
				panic(err)
			}
			m.currentView = listView
		}
	}

	return m, nil
}

func newHelpView() (*helpView, error) {
	return &helpView{}, nil
}

func generateGenericTable(data [][]string, sep string, indent int) string {
	tableString := &strings.Builder{}
	table := tabwriter.NewWriter(tableString, 0, 2, 1, ' ', tabwriter.TabIndent)

	for _, item := range data {
		fmttedRow := ""

		for i := 1; i < indent; i++ {
			fmttedRow += " "
		}

		fmttedRow += strings.Join(item, fmt.Sprintf("\t%s ", sep))
		fmt.Fprintln(table, fmttedRow)
	}
	table.Flush()
	return tableString.String()
}
