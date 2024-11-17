package lists

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
)

type ServerSelectedMsg struct {
	Hostname string
}

type ServersListModel struct {
	panel string

	filterInput textinput.Model

	matchesIndex int

	servers []string
	matches fuzzy.Matches
}

func InitServersListModel() ServersListModel {
	filterInput := textinput.New()
	filterInput.Prompt = "> "
	filterInput.Placeholder = "host.example.com"
	filterInput.CharLimit = 64
	filterInput.Focus()

	return ServersListModel{
		panel:       "filter",
		filterInput: filterInput,
		servers:     []string{},
	}
}

func (m ServersListModel) SetServers(servers []string) ServersListModel {
	m.panel = "filter"
	m.servers = servers
	m.matchesIndex = 0
	m.filterInput.Focus()

	if m.filterInput.Value() != "" {
		m.matches = fuzzy.Find(m.filterInput.Value(), m.servers)
	}

	return m
}

func (m ServersListModel) Update(msg tea.Msg) (ServersListModel, tea.Cmd) {
	var cmd tea.Cmd

	if m.panel == "filter" {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter", "tab":
				if m.filterInput.Value() != "" {
					if len(m.matches) == 0 {
						return m, tea.Quit
					}

					if len(m.matches) == 1 {
						m.panel = "empty"

						return m, func() tea.Msg { return ServerSelectedMsg{m.matches[0].Str} }
					}

					m.panel = "list"
					m.filterInput.Blur()

					return m, nil
				}
			}
		}

		m.filterInput, cmd = m.filterInput.Update(msg)

		if m.filterInput.Value() != "" {
			m.matches = fuzzy.Find(m.filterInput.Value(), m.servers)
		}

		return m, cmd
	}

	if m.panel == "list" {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "down", "tab":
				m.matchesIndex += 1
				if m.matchesIndex >= len(m.matches) {
					m.matchesIndex = 0
				}
			case "up":
				m.matchesIndex -= 1
				if m.matchesIndex < 0 {
					m.matchesIndex = len(m.matches) - 1
				}
			case "enter":
				m.panel = "empty"

				return m, func() tea.Msg { return ServerSelectedMsg{m.matches[m.matchesIndex].Str} }
			}
		}
	}

	return m, nil
}

func contains(needle int, haystack []int) bool {
	for _, i := range haystack {
		if needle == i {
			return true
		}
	}
	return false
}

func (m ServersListModel) View() string {
	if m.panel == "empty" {
		return ""
	}

	builder := strings.Builder{}

	builder.WriteString(m.filterInput.View())
	builder.WriteString("\n\n")

	if m.panel == "filter" {
		if len(m.matches) == 0 && m.filterInput.Value() != "" {
			builder.WriteString("No matches found")
			builder.WriteRune('\n')
		} else {
			if len(m.matches) != 0 {
				limit := min(len(m.matches), 10)

				for i, match := range m.matches[:limit] {
					word := strings.Builder{}

					for j := 0; j < len(match.Str); j++ {
						if contains(j, match.MatchedIndexes) {
							word.WriteString(foundItemStyle.Render(string(match.Str[j])))
						} else {
							word.WriteString(normalItemStyle.Render(string(match.Str[j])))
						}
					}

					builder.WriteString(itemStyle.Render(word.String()))

					if i != limit {
						builder.WriteRune('\n')
					}
				}
			} else {
				limit := min(len(m.servers), 10)

				for i, server := range m.servers[:limit] {
					builder.WriteString(itemStyle.Render(normalItemStyle.Render(server)))

					if i != 9 {
						builder.WriteRune('\n')
					}
				}
			}
		}
	}

	if m.panel == "list" {
		limit := min(len(m.matches), 10)
		from := 0
		if m.matchesIndex > (limit / 2) {
			from = m.matchesIndex - (limit / 2)
			from = min(from, len(m.matches)-limit)
		}

		for i, match := range m.matches[from : from+limit] {
			word := strings.Builder{}

			for j := 0; j < len(match.Str); j++ {
				if contains(j, match.MatchedIndexes) {
					word.WriteString(foundItemStyle.Render(string(match.Str[j])))
				} else {
					if m.matchesIndex == from+i {
						word.WriteString(string(match.Str[j]))
					} else {
						word.WriteString(normalItemStyle.Render(string(match.Str[j])))
					}
				}
			}

			if m.matchesIndex == from+i {
				builder.WriteString("> " + word.String())
			} else {
				builder.WriteString(itemStyle.Render(word.String()))
			}

			if i != limit {
				builder.WriteRune('\n')
			}
		}
	}

	return builder.String()
}
