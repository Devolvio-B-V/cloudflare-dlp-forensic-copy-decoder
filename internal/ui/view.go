package ui

import (
	"fmt"
	"strings"

	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/internal/decoder"
)

// View renders the current state
func (m Model) View() string {
	switch m.mode {
	case ModeFileSelection:
		return m.viewFileSelection()
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

func (m Model) viewFileSelection() string {
	var b strings.Builder

	b.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	b.WriteString("║        Cloudflare DLP Forensic Copy Decoder - TUI Mode         ║\n")
	b.WriteString("╚════════════════════════════════════════════════════════════════╝\n\n")

	if m.inputPath != "" {
		b.WriteString(fmt.Sprintf("Input file: %s\n\n", m.inputPath))
		b.WriteString("Press [Enter] to decode, [q] to quit\n")
	} else {
		b.WriteString("No input file specified.\n")
		b.WriteString("Usage: decoder --tui <input.log.gz>\n\n")
		b.WriteString("Press [q] to quit\n")
	}

	if m.statusMessage != "" {
		b.WriteString("\n")
		b.WriteString(m.statusMessage)
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) viewDecoding() string {
	var b strings.Builder

	b.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	b.WriteString("║                        Decoding File...                        ║\n")
	b.WriteString("╚════════════════════════════════════════════════════════════════╝\n\n")

	b.WriteString(fmt.Sprintf("Processing: %s\n\n", m.inputPath))
	b.WriteString("Please wait...\n")

	return b.String()
}

func (m Model) viewPreview() string {
	var b strings.Builder

	b.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	b.WriteString("║                      Decoded Payload Preview                   ║\n")
	b.WriteString("╚════════════════════════════════════════════════════════════════╝\n\n")

	// Show metadata
	b.WriteString(fmt.Sprintf("File: %s\n", truncate(m.inputPath, 60)))
	b.WriteString(fmt.Sprintf("Content-Type: %s\n", m.result.ContentType))

	if m.result.IsJSON {
		b.WriteString("Format: JSON\n")
	} else if m.result.IsText {
		b.WriteString("Format: Plain Text\n")
	}
	b.WriteString("\n")

	// Show preview
	b.WriteString("──────────────────────────────────────────────────────────────────\n")

	payload := string(m.result.Payload)
	previewHeight := m.height - 15
	if previewHeight < 5 {
		previewHeight = 5
	}

	// Show first N lines
	lines := strings.Split(payload, "\n")
	displayLines := previewHeight
	if len(lines) < displayLines {
		displayLines = len(lines)
	}

	for i := 0; i < displayLines; i++ {
		line := lines[i]
		if len(line) > m.width-2 {
			line = line[:m.width-5] + "..."
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	if len(lines) > displayLines {
		b.WriteString(fmt.Sprintf("\n... (%d more lines)\n", len(lines)-displayLines))
	}

	b.WriteString("──────────────────────────────────────────────────────────────────\n\n")

	// Show controls
	b.WriteString("Controls:\n")
	b.WriteString("  [s/e] Save/Export  [r] Toggle Raw  [o] Open New File  [q] Quit\n")

	if m.statusMessage != "" {
		b.WriteString("\n")
		b.WriteString(m.statusMessage)
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) viewExport() string {
	var b strings.Builder

	b.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	b.WriteString("║                         Export Payload                         ║\n")
	b.WriteString("╚════════════════════════════════════════════════════════════════╝\n\n")

	defaultPath := decoder.GetOutputFilename(m.inputPath, m.result.IsJSON)
	b.WriteString(fmt.Sprintf("Default path: %s\n", defaultPath))
	b.WriteString("Custom path: ")

	if m.exportPath != "" {
		b.WriteString(m.exportPath)
	} else {
		b.WriteString("_")
	}
	b.WriteString("\n\n")

	b.WriteString("Press [Enter] to export, [Esc] to cancel\n")

	if m.statusMessage != "" {
		b.WriteString("\n")
		b.WriteString(m.statusMessage)
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) viewError() string {
	var b strings.Builder

	b.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	b.WriteString("║                            ERROR                               ║\n")
	b.WriteString("╚════════════════════════════════════════════════════════════════╝\n\n")

	b.WriteString("An error occurred:\n\n")

	errMsg := wrapText(m.err.Error(), m.width-4)
	b.WriteString(errMsg)
	b.WriteString("\n\n")

	b.WriteString("Press [o] to try another file, [q] to quit\n")

	return b.String()
}
