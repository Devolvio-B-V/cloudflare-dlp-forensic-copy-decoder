// Package ui provides the terminal user interface for the decoder
package ui

import (
	"fmt"
	"strings"

	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/internal/decoder"
	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/pkg/utils"
	tea "github.com/charmbracelet/bubbletea"
)

// Mode represents the current UI mode
type Mode int

const (
	ModeFileSelection Mode = iota
	ModeDecoding
	ModePreview
	ModeExport
	ModeError
)

// Model represents the application state
type Model struct {
	mode          Mode
	inputPath     string
	result        *decoder.DecodeResult
	err           error
	showRaw       bool
	exportPath    string
	statusMessage string
	width         int
	height        int
}

// NewModel creates a new TUI model
func NewModel(inputPath string) Model {
	return Model{
		mode:      ModeFileSelection,
		inputPath: inputPath,
		width:     80,
		height:    24,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// If input path is provided, start decoding immediately
	if m.inputPath != "" {
		return m.decode()
	}
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case decodeSuccessMsg:
		m.result = msg.result
		m.mode = ModePreview
		m.statusMessage = "Decoded successfully"
		return m, nil
	case decodeErrorMsg:
		m.err = msg.err
		m.mode = ModeError
		return m, nil
	case exportSuccessMsg:
		m.mode = ModePreview
		m.exportPath = ""
		m.statusMessage = fmt.Sprintf("Exported to: %s", msg.path)
		return m, nil
	case exportErrorMsg:
		m.statusMessage = fmt.Sprintf("Export failed: %s", msg.err.Error())
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeFileSelection:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.inputPath != "" {
				m.mode = ModeDecoding
				return m, m.decode()
			}
		}

	case ModePreview:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.showRaw = !m.showRaw
			m.statusMessage = fmt.Sprintf("Raw mode: %v", m.showRaw)
		case "s", "e":
			m.mode = ModeExport
			m.statusMessage = "Enter export path (or press enter to use default)"
		case "o":
			m.mode = ModeFileSelection
			m.statusMessage = "Enter new file path"
			m.result = nil
		}

	case ModeExport:
		switch msg.String() {
		case "esc":
			m.mode = ModePreview
			m.exportPath = ""
			m.statusMessage = ""
		case "enter":
			return m, m.exportFile()
		case "backspace":
			if len(m.exportPath) > 0 {
				m.exportPath = m.exportPath[:len(m.exportPath)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.exportPath += msg.String()
			}
		}

	case ModeError:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "o":
			m.mode = ModeFileSelection
			m.err = nil
			m.statusMessage = "Enter new file path"
		}
	}

	return m, nil
}

// Messages

type decodeSuccessMsg struct {
	result *decoder.DecodeResult
}

type decodeErrorMsg struct {
	err error
}

type exportSuccessMsg struct {
	path string
}

type exportErrorMsg struct {
	err error
}

// Commands

func (m Model) decode() tea.Cmd {
	return func() tea.Msg {
		// Read input file
		reader, err := utils.ReadInput(m.inputPath)
		if err != nil {
			return decodeErrorMsg{err: err}
		}

		// Decode the file
		opts := decoder.DecodeOptions{
			TryText: false,
			Verbose: true,
		}
		result, err := decoder.DecodeLogFile(reader, opts)
		if err != nil {
			return decodeErrorMsg{err: err}
		}

		return decodeSuccessMsg{result: result}
	}
}

func (m Model) exportFile() tea.Cmd {
	return func() tea.Msg {
		path := m.exportPath
		if path == "" {
			// Use default path
			path = decoder.GetOutputFilename(m.inputPath, m.result.IsJSON)
		}

		// Write the payload
		if err := utils.WriteFile(path, m.result.Payload, true); err != nil {
			return exportErrorMsg{err: err}
		}

		// Also write the log JSON
		logPath := decoder.GetLogJSONFilename(m.inputPath)
		if err := utils.WriteFile(logPath, []byte(m.result.LogJSON), true); err != nil {
			return exportErrorMsg{err: err}
		}

		return exportSuccessMsg{path: path}
	}
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func wrapText(text string, width int) string {
	if width <= 0 {
		width = 80
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if len(line) <= width {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Wrap long lines
		for len(line) > width {
			result.WriteString(line[:width])
			result.WriteString("\n")
			line = line[width:]
		}
		if len(line) > 0 {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}
