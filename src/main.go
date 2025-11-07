package main

import (
	"fmt"
	"os"
	"strings"

	list "github.com/charmbracelet/bubbles/list"
	textInput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

const FOCUSABLES = 3

// Global request data structure
type RequestData struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

type model struct {
	width, height  int
	showMethods    bool
	methods        list.Model
	urlInput       textInput.Model
	selectedMethod string
	focusIndex     int
	requestData    *RequestData
	activeTab      int // 0=Params 1=Auth 2=Headers 3=Body
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
		m.urlInput.Width = m.width - 25

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.focusIndex != 1 {
				return m, tea.Quit
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
					m.activeTab = (m.activeTab + 1) % 4
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
		case "ctrl+enter":
			fmt.Println()
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
	tabs := []string{"Params", "Auth", "Headers", "Body"}
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
	switch m.activeTab {
	case 0:
		tabContent = "Query Parameters"
	case 1:
		tabContent = "Authentication"
	case 2:
		tabContent = "Headers"
	case 3:
		tabContent = "Body"
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

	ti := textInput.New()
	ti.Placeholder = "Enter URL..."
	ti.Focus()

	requestData := &RequestData{
		Method:  "GET",
		URL:     "",
		Headers: make(map[string]string),
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
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running: ", err)
		os.Exit(1)
	}
}
