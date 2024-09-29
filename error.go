package main

import tea "github.com/charmbracelet/bubbletea"

type errorMsg struct {
	err error
}

func ErrorMsg(err error) tea.Cmd {
	return func() tea.Msg {
		return errorMsg{err}
	}
}
