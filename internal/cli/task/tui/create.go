package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/selection"
)

const (
	titleFocus       int = 0
	descriptionFocus int = 1
	parentsFocus     int = 2
)

type task struct {
	ID    string
	Title string
}

type createView struct {
	title        textinput.Model
	description  textarea.Model
	parents      *selection.Selection[task]
	currentFocus int
}

func (v *createView) display(m model) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#787878")).AlignHorizontal(lipgloss.Center).PaddingLeft(2).MarginBottom(1)
	bodyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D3D3D3")).PaddingLeft(2)

	header := headerStyle.Render("Create")

	output := header + "\n"
	output += bodyStyle.Render(v.title.View()) + "\n"
	output += bodyStyle.Render("Description: ") + "\n"
	output += bodyStyle.Render(v.description.View()) + "\n"

	return output
}

func (v *createView) handleKey(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if v.currentFocus+1 > 2 {
				v.currentFocus = 0
			} else {
				v.currentFocus += 1
			}

			switch v.currentFocus {
			case 0:
				v.title.Focus()
				v.description.Blur()
				var cmd tea.Cmd
				v.title, cmd = v.title.Update(msg)
				return m, cmd
			case 1:
				v.title.Blur()
				v.description.Focus()
				var cmd tea.Cmd
				v.description, cmd = v.description.Update(msg)
				return m, cmd

			}
		}
	}

	switch v.currentFocus {
	case 0:
		var cmd tea.Cmd
		v.title, cmd = v.title.Update(msg)
		return m, cmd
	case 1:
		var cmd tea.Cmd
		v.description, cmd = v.description.Update(msg)
		return m, cmd
	}

	return m, nil
}

func newCreateView(tasks map[string]string) (*createView, error) {
	titleInput := textinput.New()
	titleInput.Prompt = "Title: "
	titleInput.Focus()

	descInput := textarea.New()
	descInput.Blur()
	descInput.ShowLineNumbers = false

	choices := []task{}

	for id, title := range tasks {
		choices = append(choices, task{ID: id, Title: title})
	}

	parentSelect := selection.New("Choose a parent", choices)
	parentSelect.FilterPrompt = "Filter:"
	parentSelect.FilterPlaceholder = "Type to filter"
	parentSelect.PageSize = 3
	parentSelect.Filter = func(filter string, choice *selection.Choice[task]) bool {
		return strings.HasPrefix(choice.Value.Title, filter)
	}

	return &createView{
		title:        titleInput,
		description:  descInput,
		parents:      parentSelect,
		currentFocus: 0,
	}, nil
}

type parentSelection struct {
	parentList []string
	parents    map[string]string
	cursor     int
}

func newParentSelectionModal(parents map[string]string) *parentSelection {
	parentList := []string{}

	for id := range parents {
		parentList = append(parentList, id)
	}

	sort.Strings(parentList)

	return &parentSelection{
		parentList: parentList,
		parents:    parents,
		cursor:     0,
	}
}

func (p *parentSelection) View() string {
	output := ""

	for _, id := range p.parentList {
		output += fmt.Sprintf("[%s] %s", id, p.parents[id])
	}

	return output
}

func (p *parentSelection) Update(msg tea.Msg) (parentSelection, tea.Cmd) {
	return parentSelection{}, nil
}
