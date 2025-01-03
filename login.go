package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/pquerna/otp/totp"
)

var (
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	noStyle      = lipgloss.NewStyle()
)

type SubmitMsg struct{}

type LoginModel struct {
	stage string

	focusIndex    int
	passwordInput textinput.Model
	secretInput   textinput.Model

	secret string
	code   string
}

func InitLoginModel() LoginModel {
	m := LoginModel{
		stage:         "edit",
		passwordInput: textinput.New(),
		secretInput:   textinput.New(),
	}

	m.passwordInput.EchoMode = textinput.EchoPassword
	m.passwordInput.EchoCharacter = '•'

	m.passwordInput.Focus()

	m.passwordInput.PromptStyle = noStyle
	m.passwordInput.TextStyle = noStyle

	m.secretInput.EchoMode = textinput.EchoPassword
	m.secretInput.EchoCharacter = '•'

	m.secretInput.PromptStyle = blurredStyle
	m.secretInput.TextStyle = blurredStyle

	return m
}

func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	err, isErr := msg.(errorMsg)
	if isErr {
		return m, tea.Sequence(
			tea.Println(err.err),
			tea.Quit,
		)
	}

	if m.stage == "edit" {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit

			case "tab", "shift+tab", "enter", "up", "down":
				s := msg.String()

				if s == "enter" && m.focusIndex == 1 {
					m.stage = "submit"

					m.secretInput.PromptStyle = blurredStyle
					m.secretInput.TextStyle = blurredStyle
					m.secretInput.Blur()

					m.passwordInput.PromptStyle = blurredStyle
					m.passwordInput.TextStyle = blurredStyle
					m.passwordInput.Blur()

					secret := m.secretInput.Value()

					if strings.HasPrefix(secret, "data:image/png;base64,") {
						imageRaw, err := base64.StdEncoding.DecodeString(secret[22:])
						if err != nil {
							return m, ErrorMsg(err)
						}

						image, _, err := image.Decode(bytes.NewReader(imageRaw))
						if err != nil {
							return m, ErrorMsg(err)
						}

						bmp, err := gozxing.NewBinaryBitmapFromImage(image)
						if err != nil {
							return m, ErrorMsg(err)
						}

						qrReader := qrcode.NewQRCodeReader()
						result, err := qrReader.Decode(bmp, nil)
						if err != nil {
							return m, ErrorMsg(err)
						}

						otpauth, err := url.Parse(result.GetText())
						if err != nil {
							return m, ErrorMsg(err)
						}

						secret = otpauth.Query().Get("secret")
					}

					if len(secret) == 0 {
						return m, ErrorMsg(errors.New("no secret in url"))
					}

					code, err := totp.GenerateCode(secret, time.Now())
					if err != nil {
						return m, ErrorMsg(err)
					}

					m.secret = secret
					m.code = code

					return m, nil
				}

				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > 1 {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = 1
				}

				if m.focusIndex == 0 {
					cmd := m.passwordInput.Focus()
					m.passwordInput.PromptStyle = noStyle
					m.passwordInput.TextStyle = noStyle

					m.secretInput.Blur()
					m.secretInput.PromptStyle = blurredStyle
					m.secretInput.TextStyle = blurredStyle

					return m, cmd
				}

				if m.focusIndex == 1 {
					cmd := m.secretInput.Focus()
					m.secretInput.PromptStyle = noStyle
					m.secretInput.TextStyle = noStyle

					m.passwordInput.Blur()
					m.passwordInput.PromptStyle = blurredStyle
					m.passwordInput.TextStyle = blurredStyle

					return m, cmd
				}
			}
		}

		var cmd tea.Cmd

		if m.focusIndex == 0 {
			m.passwordInput, cmd = m.passwordInput.Update(msg)

			return m, cmd
		}

		if m.focusIndex == 1 {
			m.secretInput, cmd = m.secretInput.Update(msg)

			return m, cmd
		}
	}

	if m.stage == "submit" {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "r":
				code, err := totp.GenerateCode(m.secret, time.Now())
				if err != nil {
					return m, ErrorMsg(err)
				}

				m.code = code

				return m, nil
			case "enter":
				err := StoreAuth(Auth{
					Password: m.passwordInput.Value(),
					Secret:   m.secret,
				})
				if err != nil {
					return m, ErrorMsg(err)
				}

				m.stage = "success"

				return m, tea.Quit
			}

		}
	}

	return m, nil
}

func (m LoginModel) View() string {
	var b strings.Builder

	if m.stage == "edit" {
		if m.passwordInput.Focused() {
			b.WriteString(noStyle.Render("Password:"))
			b.WriteRune('\n')
		} else {
			b.WriteString(blurredStyle.Render("Password:"))
			b.WriteRune('\n')
		}
		b.WriteString(m.passwordInput.View())

		b.WriteRune('\n')
		b.WriteRune('\n')

		if m.secretInput.Focused() {
			b.WriteString(noStyle.Render("Secret (OTP secret or data:image/png;base64 QRcode string):"))
			b.WriteRune('\n')
		} else {
			b.WriteString(blurredStyle.Render("Secret (OTP secret or data:image/png;base64 QRcode string):"))
			b.WriteRune('\n')
		}
		b.WriteString(m.secretInput.View())
	}

	if m.stage == "submit" {
		b.WriteString(blurredStyle.Render("Password:"))
		b.WriteRune('\n')
		b.WriteString(m.passwordInput.View())

		b.WriteRune('\n')
		b.WriteRune('\n')

		b.WriteString(blurredStyle.Render("Secret (OTP secret or data:image/png;base64 QRcode string):"))
		b.WriteRune('\n')
		b.WriteString(m.secretInput.View())

		b.WriteRune('\n')
		b.WriteRune('\n')

		b.WriteString("OTP code: ")
		b.WriteString(m.code)

		b.WriteRune('\n')
		b.WriteRune('\n')

		b.WriteString(blurredStyle.Render("press r to update otp code"))
	}

	if m.stage == "success" {
		b.WriteString(blurredStyle.Render("Password:"))
		b.WriteRune('\n')
		b.WriteString(m.passwordInput.View())

		b.WriteRune('\n')
		b.WriteRune('\n')

		b.WriteString(blurredStyle.Render("Secret (OTP secret or data:image/png;base64 QRcode string):"))
		b.WriteRune('\n')
		b.WriteString(m.secretInput.View())

		b.WriteRune('\n')
		b.WriteRune('\n')

		b.WriteString(blurredStyle.Render("OTP code: "))
		b.WriteString(blurredStyle.Render(m.code))

		b.WriteRune('\n')
		b.WriteRune('\n')

		b.WriteString("Success!")
		b.WriteRune('\n')
	}

	return b.String()
}
