package main

import (
	"fmt"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/Firebain/tssh/lists"
	"github.com/creack/pty"
	"github.com/pquerna/otp/totp"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gravitational/teleport/api/client"
)

type LoginSuccess struct {
	cr client.Credentials
}

type CacheEmptyMsg struct{}

type CacheLoadedMsg struct {
	servers *ServersInfo
}

type ServersLoadedMsg struct {
	servers *ServersInfo
}

type UserSelectedMsg struct{}

func RunLoginCmd() tea.Cmd {
	auth, err := GetAuth()
	if err != nil {
		return ErrorMsg(err)
	}

	c := exec.Command("tsh", "login")

	if auth == nil {
		return tea.ExecProcess(c, func(err error) tea.Msg {
			if err != nil {
				return errorMsg{err}
			}

			cr := client.LoadProfile("", "")

			return LoginSuccess{cr}
		})
	} else {
		f, err := pty.Start(c)
		if err != nil {
			return ErrorMsg(err)
		}
		defer f.Close()

		_, err = io.WriteString(f, auth.password+"\n")
		if err != nil {
			return ErrorMsg(err)
		}

		code, err := totp.GenerateCode(auth.secret, time.Now())
		if err != nil {
			return ErrorMsg(err)
		}

		_, err = io.WriteString(f, code+"\n")
		if err != nil {
			return ErrorMsg(err)
		}

		io.Copy(io.Discard, f)

		cr := client.LoadProfile("", "")

		return func() tea.Msg { return LoginSuccess{cr} }
	}
}

func RunConnectCmd(user string, hostname string) tea.Cmd {
	c := exec.Command("tsh", "ssh", user+"@"+hostname)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errorMsg{err}
		}

		return tea.Quit()
	})
}

type AppModel struct {
	cr   client.Credentials
	info *ServersInfo

	panel string

	spinner     spinner.Model
	serversList lists.ServersListModel
	usersList   lists.UsersListModel
}

func InitAppModel() AppModel {
	cr := client.LoadProfile("", "")

	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	return AppModel{
		cr: cr,

		panel: "empty",

		spinner:     s,
		serversList: lists.InitServersListModel(),
		usersList:   lists.InitUsersListModel(),
	}
}

func (m AppModel) Init() tea.Cmd {
	expireAt, canDetectExpire := m.cr.Expiry()
	if !canDetectExpire {
		return tea.Sequence(
			tea.Println("Can't detect profile. Please run 'tsh login'"),
			tea.Quit,
		)
	}

	if expireAt.Before(time.Now()) {
		return RunLoginCmd()
	}

	servers, err := GetServersInfoFromCache()
	if err != nil {
		return func() tea.Msg { return CacheEmptyMsg{} }
	}

	return func() tea.Msg { return CacheLoadedMsg{servers} }
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "ctrl+u":
			m.panel = "user"

			return m, nil
		}
	case errorMsg:
		m.panel = "empty"

		return m, tea.Sequence(
			tea.Println(msg.err),
			tea.Quit,
		)
	case LoginSuccess:
		m.cr = msg.cr

		return m, func() tea.Msg {
			servers, err := GetServersInfoFromCache()
			if err != nil {
				return CacheEmptyMsg{}
			}

			return CacheLoadedMsg{servers}
		}
	case CacheEmptyMsg:
		m.panel = "spiner"

		return m, tea.Batch(
			m.spinner.Tick,
			func() tea.Msg {
				info, err := FetchServersInfo(m.cr)
				if err != nil {
					return errorMsg{err}
				}

				return ServersLoadedMsg{info}
			},
		)
	case CacheLoadedMsg:
		m.info = msg.servers

		names := []string{}
		for _, server := range msg.servers.Servers {
			names = append(names, server.Name)
		}

		m.serversList = m.serversList.SetServers(names)
		m.usersList = m.usersList.SetUsers(msg.servers.Logins)

		if msg.servers.DefaultLogin == "" {
			m.panel = "user"

			return m, nil
		}

		m.panel = "list"

		return m, nil
	case ServersLoadedMsg:
		err := StoreServersInfo(msg.servers)
		if err != nil {
			return m, ErrorMsg(err)
		}

		m.info = msg.servers

		names := []string{}
		for _, server := range msg.servers.Servers {
			names = append(names, server.Name)
		}

		m.serversList = m.serversList.SetServers(names)
		m.usersList = m.usersList.SetUsers(msg.servers.Logins)

		if msg.servers.DefaultLogin == "" {
			m.panel = "user"

			return m, nil
		}

		m.panel = "list"

		return m, nil
	case lists.UserSelectedMsg:
		m.info.DefaultLogin = msg.User

		err := StoreServersInfo(m.info)
		if err != nil {
			return m, ErrorMsg(err)
		}

		m.panel = "list"

		return m, nil
	case lists.ServerSelectedMsg:
		return m, RunConnectCmd(m.info.DefaultLogin, msg.Hostname)
	}

	var cmd tea.Cmd

	if m.panel == "spiner" {
		m.spinner, cmd = m.spinner.Update(msg)

		return m, cmd
	}

	if m.panel == "user" {
		m.usersList, cmd = m.usersList.Update(msg)

		return m, cmd
	}

	if m.panel == "list" {
		m.serversList, cmd = m.serversList.Update(msg)

		return m, cmd
	}

	return m, nil
}

func (m AppModel) View() string {
	if m.panel == "empty" {
		return ""
	}

	if m.panel == "spiner" {
		return fmt.Sprintf("%s Loading servers...\n", m.spinner.View())
	}

	if m.panel == "user" {
		return fmt.Sprintf("Select default user:\n\n%s\n", m.usersList.View())
	}

	if m.panel == "list" {
		return m.serversList.View()
	}

	return ""
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "login" {
		m := InitLoginModel()
		p := tea.NewProgram(m)
		_, err := p.Run()
		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}

		return
	}

	if len(os.Args) == 3 && os.Args[1] == "login" && os.Args[2] == "forget" {
		err := DeleteAuth()

		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}

		fmt.Println("Deleted")

		return
	}

	m := InitAppModel()
	p := tea.NewProgram(m)
	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
