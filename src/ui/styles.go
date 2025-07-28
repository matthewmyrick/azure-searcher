package ui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	SelectedStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#874BFD")).
		Foreground(lipgloss.Color("#FFFFFF"))

	ResourceGroupStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)

	ResourceStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	CacheStatusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))
)