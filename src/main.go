package main

import (
	"fmt"
	"os"
	"strings"

	list "github.com/charmbracelet/bubbles/list"
	textArea "github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

const FOCUSABLES = 3

// Global request data structure
type RequestData struct {
	Method  string
	URL     string
	Params  string
	Auth    string
	Headers string
	Body    string
}

type model struct {
	width, height  int
	showMethods    bool
	methods        list.Model
	urlInput       textArea.Model
	selectedMethod string
	focusIndex     int
	requestData    *RequestData
	activeTab      int // 0=Params 1=Auth 2=Headers 3=Body 4=Response
	insertMode     bool
	paramsInput    textArea.Model
	authInput      textArea.Model
	headersInput   textArea.Model
	bodyInput      textArea.Model
	responseArea   textArea.Model
}

func (m *model) saveTabContent() {
	m.requestData.Params = m.paramsInput.Value()
	m.requestData.Auth = m.authInput.Value()
	m.requestData.Headers = m.headersInput.Value()
	m.requestData.Body = m.bodyInput.Value()
}

type item string

func (i item) FilterValue() string { return string(i) }
func (i item) Title() string       { return string(i) }
func (i item) Description() string { return "" }

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.methods.SetWidth(15)
		m.methods.SetHeight(3)

		m.urlInput.SetWidth(m.width - 25)
		m.urlInput.SetHeight(3)

		m.paramsInput.SetHeight(m.height - 20)
		m.paramsInput.SetWidth(m.width - 12)

		m.authInput.SetHeight(m.height - 20)
		m.authInput.SetWidth(m.width - 12)

		m.headersInput.SetHeight(m.height - 20)
		m.headersInput.SetWidth(m.width - 12)

		m.bodyInput.SetHeight(m.height - 20)
		m.bodyInput.SetWidth(m.width - 12)

		m.responseArea.SetHeight(m.height - 20)
		m.responseArea.SetWidth(m.width - 12)

	case tea.KeyMsg:

		if m.insertMode && msg.String() == "esc" { // exit insert
			m.insertMode = false
			m.paramsInput.Blur()
			m.authInput.Blur()
			m.headersInput.Blur()
			m.bodyInput.Blur()
			m.responseArea.Blur()
			return m, nil
		}

		if m.insertMode { // handle insert mode
			switch m.activeTab {
			case 0:
				m.paramsInput, cmd = m.paramsInput.Update(msg)
			case 1:
				m.authInput, cmd = m.authInput.Update(msg)
			case 2:
				m.headersInput, cmd = m.headersInput.Update(msg)
			case 3:
				m.bodyInput, cmd = m.bodyInput.Update(msg)
			case 4:
				m.responseArea, cmd = m.responseArea.Update(msg)
			}
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.focusIndex != 1 {
				return m, tea.Quit
			}
		case "i":
			if m.focusIndex == 2 {
				m.insertMode = true
				switch m.activeTab {
				case 0:
					m.paramsInput.Focus()
				case 1:
					m.authInput.Focus()
				case 2:
					m.headersInput.Focus()
				case 3:
					m.bodyInput.Focus()
				case 4:
					m.responseArea.Focus()
				}
			}
		case "m":
			if m.focusIndex != 1 {
				m.focusIndex = 0
				m.showMethods = true
			}
		case "u":
			if m.showMethods {
				m.showMethods = false
			}
			if m.focusIndex != 1 {
				m.focusIndex = 1
				m.urlInput.Focus()
			}
		case "tab":
			if m.showMethods {
				m.showMethods = false
			} else {
				if m.focusIndex == 2 {
					m.saveTabContent()
					m.activeTab = (m.activeTab + 1) % 5 // cycle between tabs
				} else {
					m.focusIndex = (m.focusIndex + 1) % FOCUSABLES
					if m.focusIndex == 0 {
						m.urlInput.Blur()
						m.showMethods = true
						m.requestData.URL = m.urlInput.Value()
					} else if m.focusIndex == 1 {
						m.urlInput.Focus()
						m.showMethods = false
					} else if m.focusIndex == 2 {
						m.urlInput.Blur()
					}
				}
			}
		case "enter":
			if m.showMethods {
				i, ok := m.methods.SelectedItem().(list.Item)
				if ok {
					m.selectedMethod = i.FilterValue()
					m.requestData.Method = m.selectedMethod
				}
				m.showMethods = false
				m.focusIndex = 1
				m.urlInput.Focus()
			}
		case "e":
			if m.focusIndex != 1 {
				m.saveTabContent()

				response, err := executeHttpRequest(m.requestData)
				if err != nil {
					m.responseArea.SetValue(fmt.Sprintf("Error: %v", err))
				} else {
					m.responseArea.SetValue(response)
				}
				m.activeTab = 4
			}
		}
	}

	if m.showMethods {
		m.methods, cmd = m.methods.Update(msg)
	} else if m.focusIndex == 1 {
		m.urlInput, cmd = m.urlInput.Update(msg)
		m.requestData.URL = m.urlInput.Value()
	}

	return m, cmd
}

func (m model) View() string {
	if m.width < 40 || m.height < 10 {
		return "Window too small -_-"
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(0, 1).
		Render("Term-API")

	// horizontal line
	lineWidth := m.width - 4
	hline := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#84FFFF")).
		Render(strings.Repeat("─", lineWidth))

	header := lipgloss.JoinVertical(lipgloss.Left, title, hline)

	// HTTP Method Selector
	methodBorderColor := lipgloss.Color("#84FFFF")
	if m.focusIndex == 0 {
		methodBorderColor = lipgloss.Color("205") // Focused color
	}

	methodBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(methodBorderColor).
		Width(15).
		Height(3).
		Render(fmt.Sprintf(" %s ", m.selectedMethod))

	if m.showMethods {
		methodBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Render(m.methods.View())
	}

	//  URL Input
	urlBorderColor := lipgloss.Color("#84FFFF")
	if m.focusIndex == 1 {
		urlBorderColor = lipgloss.Color("205") // Focused color
	}

	urlBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(urlBorderColor).
		Width(m.width - 25).
		Height(3).
		Render(m.urlInput.View())

	requestBar := lipgloss.JoinHorizontal(lipgloss.Top, methodBox, urlBox)

	// Tabs
	tabs := []string{"Params", "Auth", "Headers", "Body", "Response"}
	var tabStrings []string

	for i, tab := range tabs {
		style := lipgloss.NewStyle().Padding(0, 2)
		if i == m.activeTab {
			style = style.
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Underline(true)
		} else {
			style = style.Foreground(lipgloss.Color("#84FFFF"))
		}
		tabStrings = append(tabStrings, style.Render(tab))

	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabStrings...)

	tabBorderColor := lipgloss.Color("#84FFFF")
	if m.focusIndex == 2 {
		tabBorderColor = lipgloss.Color("205")
	}

	tabContent := ""
	modeIndicator := ""
	if m.insertMode {
		modeIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Render("--INSERT--")
	} else {
		modeIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7d7d7dff")).
			Bold(true).
			Render("--NORMAL--")
	}

	switch m.activeTab {
	case 0:
		tabContent = lipgloss.JoinVertical(lipgloss.Left,
			"Query Parameters:",
			"",
			m.paramsInput.View(),
			"",
			modeIndicator,
		)
	case 1:
		tabContent = lipgloss.JoinVertical(lipgloss.Left,
			"Authentication:",
			"",
			m.authInput.View(),
			"",
			modeIndicator,
		)
	case 2:
		tabContent = lipgloss.JoinVertical(lipgloss.Left,
			"Headers:",
			"",
			m.headersInput.View(),
			"",
			modeIndicator,
		)
	case 3:
		tabContent = lipgloss.JoinVertical(lipgloss.Left,
			"Body:",
			"",
			m.bodyInput.View(),
			"",
			modeIndicator,
		)
	case 4:
		tabContent = lipgloss.JoinVertical(lipgloss.Left,
			"Response:",
			"",
			m.responseArea.View(),
			"",
			modeIndicator,
		)
	}

	tabBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tabBorderColor).
		Width(m.width-8).
		Height(m.height-14).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, tabBar, "", tabContent))

	// Debug Info
	// debugInfo := lipgloss.NewStyle().
	// 	Foreground(lipgloss.Color("241")).
	// 	Padding(1, 1).
	// 	Render(fmt.Sprintf("DEBUG | Method: %s | URL: %s", m.requestData.Method, m.requestData.URL))

	// Combine Everything
	content := lipgloss.JoinVertical(lipgloss.Left, header, requestBar, "", tabBox)

	outer := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#84FFFF")).
		Width(m.width-4).
		Height(m.height-4).
		Margin(0, 2).
		Render(content)

	return outer
}

func main() {
	items := []list.Item{
		item("GET"),
		item("POST"),
		item("PUT"),
		item("DELETE"),
		item("PATCH"),
	}

	methodsList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	methodsList.Title = "Method"
	methodsList.SetShowHelp(false)
	methodsList.SetShowStatusBar(false)

	ti := textArea.New()
	ti.Placeholder = "Enter URL..."
	ti.ShowLineNumbers = false
	// ti.SetHeight(1)
	ti.Focus()

	paramsInput := textArea.New()
	paramsInput.Placeholder = "key=value"

	authInput := textArea.New()
	authInput.Placeholder = "Bearer token or username:password"

	headersInput := textArea.New()
	headersInput.Placeholder = "Content-Type: application/json"

	bodyInput := textArea.New()
	bodyInput.Placeholder = "Request body..."

	responseArea := textArea.New()
	responseArea.Placeholder = "Response from the request shows here."

	requestData := &RequestData{
		Method:  "GET",
		URL:     "",
		Params:  "",
		Auth:    "",
		Headers: "",
		Body:    "",
	}

	initialModel := model{
		methods:        methodsList,
		urlInput:       ti,
		selectedMethod: "GET",
		showMethods:    false,
		focusIndex:     1,
		requestData:    requestData,
		activeTab:      0,
		insertMode:     false,
		paramsInput:    paramsInput,
		authInput:      authInput,
		headersInput:   headersInput,
		bodyInput:      bodyInput,
		responseArea:   responseArea,
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running: ", err)
		os.Exit(1)
	}
}
