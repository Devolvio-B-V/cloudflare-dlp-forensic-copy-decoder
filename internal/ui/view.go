package ui

import (
	"fmt"
	"strings"

	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/internal/decoder"
	"github.com/charmbracelet/lipgloss"
)

// Define color scheme similar to lazygit
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")). // Blue
			Background(lipgloss.Color("235")). // Dark gray
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")). // Light yellow
			Background(lipgloss.Color("57")). // Purple/blue
			Bold(true)

	directoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")). // Blue
			Bold(true)

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")) // Light gray

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")). // Light yellow
			Background(lipgloss.Color("235")). // Dark gray
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")) // Dark gray

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")) // Green

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Gray
)

// View renders the current state
func (m Model) View() string {
	switch m.mode {
	case ModeFileBrowser:
		return m.viewFileBrowser()
	case ModeDecoding:
		return m.viewDecoding()
	case ModePreview:
		return m.viewPreview()
	case ModeExport:
		return m.viewExport()
	case ModeError:
		return m.viewError()
	default:
		return "Unknown mode"
	}
}

func (m Model) viewFileBrowser() string {
	var b strings.Builder
	
	// Title bar
	title := titleStyle.Width(m.width).Render("  Cloudflare DLP Forensic Copy Decoder - File Browser")
	b.WriteString(title)
	b.WriteString("\n\n")
	
	if m.fileBrowser == nil {
		b.WriteString("Error: File browser not initialized\n")
		return b.String()
	}
	
	// Current directory
	b.WriteString(directoryStyle.Render(fmt.Sprintf("📁 %s", m.fileBrowser.GetCurrentDir())))
	b.WriteString("\n\n")
	
	// File list
	entries := m.fileBrowser.GetEntries()
	selectedIdx := m.fileBrowser.GetSelectedIndex()
	
	// Calculate visible range
	maxVisible := m.height - 12
	if maxVisible < 5 {
		maxVisible = 5
	}
	
	startIdx := m.fileBrowser.GetViewOffset()
	endIdx := startIdx + maxVisible
	if endIdx > len(entries) {
		endIdx = len(entries)
	}
	
	if len(entries) == 0 {
		b.WriteString(helpStyle.Render("  (empty directory)"))
		b.WriteString("\n")
	} else {
		for i := startIdx; i < endIdx; i++ {
			entry := entries[i]
			
			var line string
			if entry.IsDir {
				icon := "📁"
				if entry.Name == ".." {
					icon = "⬆️ "
				}
				line = fmt.Sprintf("  %s %s", icon, entry.Name)
			} else {
				icon := "📄"
				if strings.HasSuffix(entry.Name, ".log.gz") {
					icon = "📦"
				}
				line = fmt.Sprintf("  %s %s", icon, entry.Name)
			}
			
			if i == selectedIdx {
				b.WriteString(selectedStyle.Width(m.width - 2).Render(line))
			} else if entry.IsDir {
				b.WriteString(directoryStyle.Render(line))
			} else {
				b.WriteString(fileStyle.Render(line))
			}
			b.WriteString("\n")
		}
		
		// Show scroll indicator
		if len(entries) > maxVisible {
			b.WriteString(helpStyle.Render(fmt.Sprintf("  (%d-%d of %d)", startIdx+1, endIdx, len(entries))))
			b.WriteString("\n")
		}
	}
	
	b.WriteString("\n")
	
	// Status bar
	statusBar := statusBarStyle.Width(m.width).Render(
		fmt.Sprintf(" %d files/folders │ Select: ↑↓/jk │ Enter: open │ h: home │ g/G: top/bottom │ q: quit", len(entries)),
	)
	b.WriteString(statusBar)
	
	if m.statusMessage != "" {
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(m.statusMessage))
	}
	
	return b.String()
}

func (m Model) viewDecoding() string {
	var b strings.Builder
	
	title := titleStyle.Width(m.width).Render("  Decoding File...")
	b.WriteString(title)
	b.WriteString("\n\n")
	
	b.WriteString(fmt.Sprintf("Processing: %s\n\n", m.inputPath))
	b.WriteString("Please wait...\n")
	
	return b.String()
}

func (m Model) viewPreview() string {
	var b strings.Builder
	
	// Title bar
	title := titleStyle.Width(m.width).Render("  Decoded Payload Preview")
	b.WriteString(title)
	b.WriteString("\n\n")
	
	// Show metadata
	metaLine := fmt.Sprintf("📄 %s │ %s", 
		truncate(m.inputPath, 40),
		m.result.ContentType,
	)
	if m.result.IsJSON {
		metaLine += " │ JSON"
	} else if m.result.IsText {
		metaLine += " │ Text"
	}
	b.WriteString(helpStyle.Render(metaLine))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")
	
	// Show viewport with content
	b.WriteString(m.viewport.View())
	b.WriteString("\n")
	
	b.WriteString(borderStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")
	
	// Status bar with keybindings
	statusBar := statusBarStyle.Width(m.width).Render(
		" ↑↓/jk: scroll │ d/u: page │ g/G: top/bottom │ s: save │ b: back │ q: quit",
	)
	b.WriteString(statusBar)
	
	if m.statusMessage != "" {
		b.WriteString("\n")
		if strings.Contains(m.statusMessage, "success") || strings.Contains(m.statusMessage, "Exported") {
			b.WriteString(successStyle.Render(m.statusMessage))
		} else {
			b.WriteString(helpStyle.Render(m.statusMessage))
		}
	}
	
	return b.String()
}

func (m Model) viewExport() string {
	var b strings.Builder
	
	title := titleStyle.Width(m.width).Render("  Export Payload")
	b.WriteString(title)
	b.WriteString("\n\n")
	
	defaultPath := decoder.GetOutputFilename(m.inputPath, m.result.IsJSON)
	b.WriteString(fmt.Sprintf("Default path: %s\n", helpStyle.Render(defaultPath)))
	b.WriteString("Custom path: ")
	
	if m.exportPath != "" {
		b.WriteString(selectedStyle.Render(m.exportPath))
	} else {
		b.WriteString(helpStyle.Render("_"))
	}
	b.WriteString("\n\n")
	
	b.WriteString(helpStyle.Render("Press [Enter] to export, [Esc] to cancel"))
	b.WriteString("\n")
	
	if m.statusMessage != "" {
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(m.statusMessage))
	}
	
	return b.String()
}

func (m Model) viewError() string {
	var b strings.Builder
	
	title := titleStyle.Width(m.width).Render("  ERROR")
	b.WriteString(title)
	b.WriteString("\n\n")
	
	b.WriteString(errorStyle.Render("An error occurred:"))
	b.WriteString("\n\n")
	
	errMsg := wrapText(m.err.Error(), m.width-4)
	b.WriteString(errMsg)
	b.WriteString("\n\n")
	
	b.WriteString(helpStyle.Render("Press [b] to go back to file browser, [q] to quit"))
	b.WriteString("\n")
	
	return b.String()
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
