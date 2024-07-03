package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	filepicker "github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func initialModel() tea.Model {
	return model{}
}

type mediaType int

type model struct {
	// Add your model fields here
	list         list.Model
	outputFormat string
	filepicker   filepicker.Model
	selectedFile string
	outputFile   string
	quitting     bool
	err          error
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

const (
	Audio mediaType = 0
	Video mediaType = 1
)

func (m model) Init() tea.Cmd {
	// Add your initialization logic here
	return m.filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Add your update logic here
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.selectedFile != "" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.outputFormat = string(i)
				}
				m.list, cmd = m.list.Update(msg)
				return m, cmd
			}
		}

	case clearErrorMsg:
		m.err = nil
	}

	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selectedFile = path
	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	if m.selectedFile != "" && m.outputFormat == "" {
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, cmd
}

func (m model) View() string {
	// Add your view logic here
	// var style = lipgloss.NewStyle().
	// 	Bold(true).
	// 	Foreground(lipgloss.Color("#E06C75"))
	if m.quitting {
		return ""
	}
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile) + "\n")
	}

	if m.outputFormat != "" {
		s.WriteString("\n  Output format: " + m.outputFormat + "\n")
	}

	if m.selectedFile == "" {
		s.WriteString("\n\n" + m.filepicker.View() + "\n")
	}

	if m.selectedFile != "" && m.outputFormat == "" {
		s.WriteString("\n\n" + m.list.View() + "\n")
	}

	return s.String()
}

func main() {

	items := []list.Item{
		item(".mp3"),
		item(".wav"),
		item(".mp4"),
		item(".mov"),
		item(".avi"),
	}

	fp := filepicker.New()
	fp.AllowedTypes = []string{".mp3", ".wav", ".mp4", ".mov", ".avi"}
	fp.CurrentDirectory, _ = os.UserHomeDir()

	const defaultWidth = 40

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Which format do you want?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{
		list:       l,
		filepicker: fp,
	}

	teaModel, _ := tea.NewProgram(&m, tea.WithAltScreen()).Run()
	tmModel := teaModel.(model)
	fmt.Println("\n  You selected: " + m.filepicker.Styles.Selected.Render(tmModel.selectedFile) + "\n")

	f, err := tea.LogToFile("debug.txt", "debug")
	if err != nil {
		log.Fatalf("err: %w", err)
	}

	defer f.Close()

	if err != nil {
		log.Fatal(err)
	}
}
