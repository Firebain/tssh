package lists

import "github.com/charmbracelet/lipgloss"

var (
	itemStyle       = lipgloss.NewStyle().PaddingLeft(2)
	normalItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#696969"))
	foundItemStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#C3E88D"))
)
