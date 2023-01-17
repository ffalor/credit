package tui

import (
	"fmt"
	"io"

	"github.com/76creates/stickers"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ffalor/credit/pkg/util/types"
)

// keyMap is used to track key bindings
type keyMap struct {
	Submit key.Binding
	Enter  key.Binding
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Help   key.Binding
	Tab    key.Binding
	Delete key.Binding
	Quit   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Enter, k.Tab, k.Delete, k.Up},          // first column
		{k.Down, k.Left, k.Right, k.Help, k.Quit}, // second column
	}
}

var keys = keyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch focus"),
	),
	Submit: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "submit"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Delete: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "delete"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter", "select"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}

// left side list of issues that you can select
// the right side you can edit the issue title and description
// tab is a toggle between the two sides

// sessionState is used to track which model is focused
type sessionState uint

const (
	issueListView sessionState = iota
	issueSummaryInputView
	issueDescriptionInputView
	// issue list width and height
	defaultHeight = 14
	defaultWidth  = 20
)

const (
	// flexBox cell layout for main flexbox
	issueListCell = iota
	issueEditorCell
)

const (
	// flexbox cell layout for issue editor
	issueSummaryCell = iota
	issueDescriptionCell
)

const (
	// flexbox row layout for issue editor
	issueSummaryRow = iota
	issueDescriptionRow
)

type issueItem struct {
	id          string
	summary     string
	description string
	repoName    string
	selected    bool
}

func (i issueItem) FilterValue() string {
	return i.summary
}

type issueItemDelegate struct{}

func (d issueItemDelegate) Height() int  { return 1 }
func (d issueItemDelegate) Spacing() int { return 0 }
func (d issueItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (d issueItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	chosen := " "

	i, ok := listItem.(issueItem)
	if !ok {
		return
	}

	if i.selected {
		chosen = "✓"
	}

	str := fmt.Sprintf("%s %s", chosenItemStyle.Render(chosen), i.summary)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprint(w, fn(str))
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	keys                  keyMap
	help                  help.Model
	inputStyle            lipgloss.Style
	lastKey               string
	quitting              bool
	selected              map[string]struct{}
	issues                map[string]types.Issue
	mergedPrs             map[string]types.MergedPr
	epic                  string
	issueSummaryInput     textinput.Model
	issueDescriptionInput textarea.Model
	issueList             list.Model // left side list of issues
	focusedView           sessionState
	mainFlexBox           *stickers.FlexBox // main flexbox includes issue list and issue editor
}

func InitialModel(mergedPrs map[string]types.MergedPr, issues map[string]types.Issue) (model, error) {

	choices := []list.Item{}

	for _, pr := range mergedPrs {
		choices = append(choices, issueItem{
			id:          pr.Id,
			summary:     pr.Title,
			description: pr.Body,
			repoName:    pr.RepoName,
		})
	}

	for _, issue := range issues {
		choices = append(choices, issueItem{
			id:          issue.Id,
			summary:     issue.Title,
			description: issue.Body,
			repoName:    issue.RepoName,
		})
	}

	l := list.New(choices, issueItemDelegate{}, defaultWidth, defaultWidth)
	l.Title = "Unassigned Issues"
	l.Styles.PaginationStyle = paginationStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.DisableQuitKeybindings()
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.Enter,
			keys.Delete,
			keys.Quit,
			keys.Tab,
		}
	}

	selectedItem, ok := l.Items()[0].(issueItem)

	if !ok {
		return model{}, fmt.Errorf("could not cast first choice to issueItem")
	}

	issueSummaryInput := textinput.New()
	issueSummaryInput.Placeholder = "Issue Summary"
	issueSummaryInput.Prompt = "Issue Summary: "
	issueSummaryInput.PromptStyle.BorderStyle(lipgloss.NormalBorder())
	issueSummaryInput.SetValue(selectedItem.summary)

	issueDescriptionInput := textarea.New()
	issueDescriptionInput.Placeholder = "Issue Description"
	issueDescriptionInput.ShowLineNumbers = false
	issueDescriptionInput.SetValue(selectedItem.description)

	mainFlexBox := stickers.NewFlexBox(0, 0)
	mainFlexBoxRows := []*stickers.FlexBoxRow{
		mainFlexBox.NewRow().AddCells(
			[]*stickers.FlexBoxCell{
				stickers.NewFlexBoxCell(1, 6),
				stickers.NewFlexBoxCell(2, 6),
			},
		),
	}
	mainFlexBox.AddRows(mainFlexBoxRows)

	return model{
		selected:              make(map[string]struct{}),
		focusedView:           0,
		issueList:             l,
		keys:                  keys,
		help:                  l.Help,
		inputStyle:            lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75B7")),
		issues:                issues,
		mergedPrs:             mergedPrs,
		issueSummaryInput:     issueSummaryInput,
		issueDescriptionInput: issueDescriptionInput,
		mainFlexBox:           mainFlexBox,
	}, nil
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// if m.epic == "" {
	// 	prompt := &survey.Input{
	// 		Message: "Please enter the epic number: ",
	// 	}
	// 	survey.AskOne(prompt, &m.epic)
	// }

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can it can gracefully truncate
		// its view as needed.
		m.help.Width = msg.Width
		m.issueList.SetWidth(msg.Width)
		m.mainFlexBox.SetWidth(msg.Width)
		m.mainFlexBox.SetHeight(msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up, m.keys.Down):
			// for some reason .Index() returns the previous or future index depending on the direction
			idx := m.issueList.Index()

			if key.Matches(msg, m.keys.Up) {
				idx = idx - 1
				if idx < 0 {
					idx = 0
				}
			} else {
				idx = idx + 1
				if idx > len(m.issueList.Items())-1 {
					idx = len(m.issueList.Items()) - 1
				}
			}

			selectedItem, ok := m.issueList.Items()[idx].(issueItem)
			if ok {
				m.issueSummaryInput.SetValue(selectedItem.summary)
				m.issueDescriptionInput.SetValue(selectedItem.description)
			}

		case key.Matches(msg, m.keys.Enter):
			if m.focusedView == issueListView {
				selectedItem := m.issueList.Items()[m.issueList.Index()]
				if selectedItem == nil {
					break
				}
				issue, ok := selectedItem.(issueItem)
				if !ok {
					break
				}
				issue.selected = !issue.selected
				m.issueList.SetItem(m.issueList.Index(), issue)
			}
		case key.Matches(msg, m.keys.Tab):
			m.nextView()
		case key.Matches(msg, m.keys.Delete):
			if m.focusedView == issueListView {
				m.issueList.RemoveItem(m.issueList.Index())
			}
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		}
	}

	// only update issueList if it's the focused view
	if m.focusedView == issueListView {
		newListModel, newListCmd := m.issueList.Update(msg)
		m.issueList = newListModel
		cmds = append(cmds, newListCmd)
	}

	newSummaryInputModel, newSummaryInputCmd := m.issueSummaryInput.Update(msg)
	m.issueSummaryInput = newSummaryInputModel

	newDescriptionInputModel, newDescriptionInputCmd := m.issueDescriptionInput.Update(msg)
	m.issueDescriptionInput = newDescriptionInputModel

	cmds = append(cmds, newSummaryInputCmd, newDescriptionInputCmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	//helpView := m.help.View(m.keys)

	if m.quitting {
		return "Bye!\n"
	}

	mainFlexBoxRow, ok := m.mainFlexBox.GetRow(0)
	if !ok {
		return "Unable to get main row"
	}

	issueEditorCellView := fmt.Sprintf("%s\n\n\n%s", m.issueSummaryInput.View(), m.issueDescriptionInput.View())
	switch m.focusedView {
	case issueListView:
		mainFlexBoxRow.Cell(issueListCell).SetStyle(focusedModelStyle).SetContent(m.issueList.View())
		mainFlexBoxRow.Cell(issueEditorCell).SetStyle(modelStyle).SetContent(issueEditorCellView)
	case issueSummaryInputView, issueDescriptionInputView:
		mainFlexBoxRow.Cell(issueListCell).SetStyle(modelStyle).SetContent(m.issueList.View())
		mainFlexBoxRow.Cell(issueEditorCell).SetStyle(focusedModelStyle).SetContent(issueEditorCellView)
	}

	return m.mainFlexBox.Render()
}

func (m *model) nextView() {
	currentView := m.focusedView
	m.focusedView = (currentView + 1) % 3

	// if we're not on the issue list view, then ensure the item we were on is updated
	if !(currentView == issueListView) {
		selectedItem := m.issueList.Items()[m.issueList.Index()]
		if selectedItem != nil {
			issue, ok := selectedItem.(issueItem)
			if ok {
				issue.summary = m.issueSummaryInput.Value()
				issue.description = m.issueDescriptionInput.Value()
				m.issueList.SetItem(m.issueList.Index(), issue)
			}
		}
	}

	switch m.focusedView {
	case issueListView:
		m.issueSummaryInput.Blur()
		m.issueDescriptionInput.Blur()
	case issueSummaryInputView:
		m.issueSummaryInput.Focus()
		m.issueDescriptionInput.Blur()
	case issueDescriptionInputView:
		m.issueSummaryInput.Blur()
		m.issueDescriptionInput.Focus()
	}
}
