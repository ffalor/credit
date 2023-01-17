package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	normalItem = lipgloss.NewStyle().
			Padding(0, 0, 0, 1)
	focusedItemStyle  = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("170"))
	repoNameStyle     = lipgloss.NewStyle().PaddingBottom(1).Foreground(lipgloss.Color("69"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("#04B575"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	modelStyle        = lipgloss.NewStyle().
				BorderStyle(lipgloss.HiddenBorder())
	focusedModelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(""))
	titleStyle     = lipgloss.NewStyle().MarginLeft(2).Background(lipgloss.Color("69"))
	statusBarStyle = list.DefaultStyles().StatusBar.MarginLeft(2)
)
