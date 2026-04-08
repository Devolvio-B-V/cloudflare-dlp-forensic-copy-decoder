// Package ui provides the terminal user interface for the decoder
package ui

import (
	"fmt"

	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/internal/decoder"
	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/pkg/utils"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Mode represents the current UI mode
type Mode int

const (
	// ModeFileBrowser is the UI mode for browsing files.
	ModeFileBrowser Mode = iota
	// ModeDecoding is the UI mode while decoding a selected file.
	ModeDecoding
	// ModePreview is the UI mode showing decoded content preview.
	ModePreview
	// ModeExport is the UI mode used when exporting decoded payloads.
	ModeExport
	// ModeError is the UI mode shown when an error occurs.
	ModeError
)

// Model represents the application state
type Model struct {
	mode             Mode
	inputPath        string
	results          []*decoder.DecodeResult
	currentResultIdx int
	result           *decoder.DecodeResult
	err              error
	showRaw          bool
	exportPath       string
	statusMessage    string
	width            int
	height           int
	fileBrowser      *FileBrowser
	viewport         viewport.Model
	ready            bool
}

// NewModel creates a new TUI model
func NewModel(inputPath string) Model {
	var fb *FileBrowser
	mode := ModeFileBrowser

	// If an input path is provided, skip file browser
	if inputPath != "" {
		mode = ModeDecoding
	} else {
		// Create file browser starting at current directory
		var err error
		fb, err = NewFileBrowser("")
		if err != nil {
			// Fall back to current directory on error
			fb, _ = NewFileBrowser(".")
		}
	}

	return Model{
		mode:        mode,
		inputPath:   inputPath,
		width:       80,
		height:      24,
		fileBrowser: fb,
		viewport:    viewport.New(80, 20),
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

		// Update viewport size
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-10)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 10
		}

		// Adjust file browser view offset for new height
		if m.fileBrowser != nil && m.mode == ModeFileBrowser {
			maxVisible := m.height - 12
			if maxVisible < 5 {
				maxVisible = 5
			}
			m.fileBrowser.AdjustViewOffset(maxVisible)
		}

		return m, nil
	case decodeSuccessMsg:
		m.results = msg.results
		m.currentResultIdx = 0
		m.result = m.results[0]
		m.mode = ModePreview
		if len(m.results) == 1 {
			m.statusMessage = "Decoded successfully"
		} else {
			m.statusMessage = fmt.Sprintf("Decoded successfully (%d payloads)", len(m.results))
		}

		// Set viewport content
		content := string(m.result.Payload)
		m.viewport.SetContent(content)

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

	// Update viewport if in preview mode
	if m.mode == ModePreview {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeFileBrowser:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.fileBrowser != nil {
				m.fileBrowser.MoveDown()
				// Adjust view offset after moving
				maxVisible := m.height - 12
				if maxVisible < 5 {
					maxVisible = 5
				}
				m.fileBrowser.AdjustViewOffset(maxVisible)
			}
		case "k", "up":
			if m.fileBrowser != nil {
				m.fileBrowser.MoveUp()
				// Adjust view offset after moving
				maxVisible := m.height - 12
				if maxVisible < 5 {
					maxVisible = 5
				}
				m.fileBrowser.AdjustViewOffset(maxVisible)
			}
		case "enter":
			if m.fileBrowser != nil {
				path, isFile, err := m.fileBrowser.Enter()
				if err != nil {
					m.statusMessage = fmt.Sprintf("Error: %s", err.Error())
				} else if isFile {
					// File selected, start decoding
					m.inputPath = path
					m.mode = ModeDecoding
					return m, m.decode()
				}
			}
		case "h":
			// Go to home directory
			if m.fileBrowser != nil {
				if err := m.fileBrowser.GoToHomeDir(); err != nil {
					m.statusMessage = fmt.Sprintf("Error: %s", err.Error())
				}
			}
		case "g":
			// Go to top
			if m.fileBrowser != nil {
				m.fileBrowser.GotoTop()
			}
		case "G":
			// Go to bottom
			if m.fileBrowser != nil {
				m.fileBrowser.GotoBottom()
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
		case "o", "b":
			// Go back to file browser
			m.mode = ModeFileBrowser
			m.statusMessage = "Select a file to decode"
			m.results = nil
			m.currentResultIdx = 0
			m.result = nil
		case "j", "down":
			m.viewport.ScrollDown(1)
		case "k", "up":
			m.viewport.ScrollUp(1)
		case "d", "ctrl+d":
			m.viewport.HalfPageDown()
		case "u", "ctrl+u":
			m.viewport.HalfPageUp()
		case "g":
			m.viewport.GotoTop()
		case "G":
			m.viewport.GotoBottom()
		case "n":
			// Navigate to next payload (multi-payload files)
			if len(m.results) > 1 && m.currentResultIdx < len(m.results)-1 {
				m.currentResultIdx++
				m.result = m.results[m.currentResultIdx]
				m.viewport.SetContent(string(m.result.Payload))
				m.viewport.GotoTop()
			}
		case "p":
			// Navigate to previous payload (multi-payload files)
			if len(m.results) > 1 && m.currentResultIdx > 0 {
				m.currentResultIdx--
				m.result = m.results[m.currentResultIdx]
				m.viewport.SetContent(string(m.result.Payload))
				m.viewport.GotoTop()
			}
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
		case "o", "b":
			m.mode = ModeFileBrowser
			m.err = nil
			m.results = nil
			m.result = nil
			m.statusMessage = "Select a file to decode"
		}
	}

	return m, nil
}

// Messages

type decodeSuccessMsg struct {
	results []*decoder.DecodeResult
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

		// Decode the file (supports single and multi-payload files)
		opts := decoder.DecodeOptions{
			TryText: false,
			Verbose: true,
		}
		results, err := decoder.DecodeLogFileMulti(reader, opts)
		if err != nil {
			return decodeErrorMsg{err: err}
		}

		return decodeSuccessMsg{results: results}
	}
}

func (m Model) exportFile() tea.Cmd {
	return func() tea.Msg {
		path := m.exportPath
		if path == "" {
			// Use default path; use numbered names for multi-payload files.
			if len(m.results) > 1 {
				path = decoder.GetOutputFilenameN(m.inputPath, m.result.IsJSON, m.currentResultIdx+1)
			} else {
				path = decoder.GetOutputFilename(m.inputPath, m.result.IsJSON)
			}
		}

		// Write the payload
		if err := utils.WriteFile(path, m.result.Payload, true); err != nil {
			return exportErrorMsg{err: err}
		}

		// Also write the log JSON
		var logPath string
		if len(m.results) > 1 {
			logPath = decoder.GetLogJSONFilenameN(m.inputPath, m.currentResultIdx+1)
		} else {
			logPath = decoder.GetLogJSONFilename(m.inputPath)
		}
		if err := utils.WriteFile(logPath, []byte(m.result.LogJSON), true); err != nil {
			return exportErrorMsg{err: err}
		}

		return exportSuccessMsg{path: path}
	}
}
