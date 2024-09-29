package lists

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type UserSelectedMsg struct {
	User string
}

type UsersListModel struct {
	index int

	users []string
}

func InitUsersListModel() UsersListModel {
	return UsersListModel{
		users: []string{},
	}
}

func (m UsersListModel) SetUsers(users []string) UsersListModel {
	m.users = users

	return m
}

func (m UsersListModel) Update(msg tea.Msg) (UsersListModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "down", "tab":
			m.index += 1
			if m.index >= len(m.users) {
				m.index = 0
			}
		case "up":
			m.index -= 1
			if m.index < 0 {
				m.index = len(m.users) - 1
			}
		case "enter":
			return m, func() tea.Msg { return UserSelectedMsg{m.users[m.index]} }
		}
	}

	return m, nil
}

func (m UsersListModel) View() string {
	builder := strings.Builder{}

	limit := min(len(m.users), 10)
	from := 0
	if m.index > (limit / 2) {
		from = m.index - (limit / 2)
		from = min(from, len(m.users)-limit)
	}

	for i, user := range m.users[from : from+limit] {
		if m.index == from+i {
			builder.WriteString("> " + user)
		} else {
			builder.WriteString(itemStyle.Render(normalItemStyle.Render(user)))
		}

		if i != limit {
			builder.WriteRune('\n')
		}
	}

	return builder.String()
}
