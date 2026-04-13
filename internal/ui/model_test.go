package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/internal/decoder"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel_WithInputPath_SetsDecodingMode(t *testing.T) {
	m := NewModel("some/path.log.gz")
	if m.mode != ModeDecoding {
		t.Fatalf("expected mode %v, got %v", ModeDecoding, m.mode)
	}
	if m.Init() == nil {
		t.Fatalf("expected non-nil Init() cmd when inputPath provided")
	}
}

func TestHandleKeyPress_PreviewToggleRawAndBack(t *testing.T) {
	m := Model{
		mode:    ModePreview,
		showRaw: false,
		result:  &decoder.DecodeResult{},
	}

	// Toggle raw mode with 'r'
	mm, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	mmModel := mm.(Model)
	if !mmModel.showRaw {
		t.Fatalf("expected showRaw to be true after pressing 'r')")
	}
	m = mmModel

	// Go back to file browser with 'o'
	mm, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	mmModel2 := mm.(Model)
	if mmModel2.mode != ModeFileBrowser {
		t.Fatalf("expected mode %v after pressing 'o', got %v", ModeFileBrowser, mmModel2.mode)
	}
	if mmModel2.result != nil {
		t.Fatalf("expected result to be nil after returning to file browser")
	}
}

func TestHandleKeyPress_FileBrowser_MoveDown(t *testing.T) {
	fb := &FileBrowser{
		entries: []FileEntry{
			{Name: "a.log.gz", Path: "a.log.gz", IsDir: false},
			{Name: "b.txt", Path: "b.txt", IsDir: false},
		},
		selectedIndex: 0,
	}

	m := Model{
		mode:        ModeFileBrowser,
		fileBrowser: fb,
		height:      24,
	}

	// Move down
	mm, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	_ = mm.(Model)
	if fb.selectedIndex != 1 {
		t.Fatalf("expected selectedIndex 1 after 'j', got %d", fb.selectedIndex)
	}
}

func TestFileBrowser_Enter_SelectsFile(t *testing.T) {
	fb := &FileBrowser{
		entries: []FileEntry{
			{Name: "a.log.gz", Path: "a.log.gz", IsDir: false},
			{Name: "b.txt", Path: "b.txt", IsDir: false},
		},
		selectedIndex: 0,
	}

	path, isFile, err := fb.Enter()
	if err != nil {
		t.Fatalf("Enter returned error: %v", err)
	}
	if !isFile {
		t.Fatalf("expected isFile true for .log.gz selection")
	}
	if path != "a.log.gz" {
		t.Fatalf("expected path a.log.gz, got %s", path)
	}
}

func TestModel_ExportFile_WritesFiles(t *testing.T) {
	tmp := t.TempDir()

	m := Model{
		mode:      ModeExport,
		inputPath: filepath.Join(tmp, "input.log.gz"),
		result: &decoder.DecodeResult{
			Payload: []byte("payload-data"),
			LogJSON: "{}",
			IsJSON:  false,
		},
		exportPath: filepath.Join(tmp, "exported.payload.txt"),
	}

	// Run export command directly
	msg := m.exportFile()()
	switch v := msg.(type) {
	case exportSuccessMsg:
		// ok
	case exportErrorMsg:
		t.Fatalf("export failed: %v", v.err)
	default:
		t.Fatalf("unexpected msg type: %T", msg)
	}

	// Check files exist
	if _, err := os.Stat(m.exportPath); err != nil {
		t.Fatalf("expected exported payload at %s, missing: %v", m.exportPath, err)
	}
	logPath := decoder.GetLogJSONFilename(m.inputPath)
	if _, err := os.Stat(logPath); err == nil {
		_ = logPath
	}
}

func TestPreview_ViewportCommands_NoPanic(t *testing.T) {
	m := Model{
		mode:     ModePreview,
		result:   &decoder.DecodeResult{Payload: []byte("x")},
		viewport: viewport.Model{},
	}

	keys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("j")},
		{Type: tea.KeyRunes, Runes: []rune("k")},
		{Type: tea.KeyRunes, Runes: []rune("d")},
		{Type: tea.KeyRunes, Runes: []rune("u")},
		{Type: tea.KeyRunes, Runes: []rune("g")},
		{Type: tea.KeyRunes, Runes: []rune("G")},
	}

	for _, k := range keys {
		mm, _ := m.handleKeyPress(k)
		_ = mm.(Model)
	}
}

func TestPreview_MultiPayload_Navigation(t *testing.T) {
	res1 := &decoder.DecodeResult{Payload: []byte("payload-one"), IsText: true}
	res2 := &decoder.DecodeResult{Payload: []byte("payload-two"), IsText: true}
	res3 := &decoder.DecodeResult{Payload: []byte("payload-three"), IsText: true}

	m := Model{
		mode:             ModePreview,
		results:          []*decoder.DecodeResult{res1, res2, res3},
		currentResultIdx: 0,
		result:           res1,
		viewport:         viewport.Model{},
	}

	// Navigate forward with 'n'
	mm, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = mm.(Model)
	if m.currentResultIdx != 1 {
		t.Fatalf("expected currentResultIdx 1 after 'n', got %d", m.currentResultIdx)
	}
	if string(m.result.Payload) != "payload-two" {
		t.Fatalf("expected payload-two, got %q", string(m.result.Payload))
	}

	// Navigate forward again
	mm, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = mm.(Model)
	if m.currentResultIdx != 2 {
		t.Fatalf("expected currentResultIdx 2, got %d", m.currentResultIdx)
	}

	// Should not go past the last result
	mm, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = mm.(Model)
	if m.currentResultIdx != 2 {
		t.Fatalf("expected currentResultIdx to stay at 2, got %d", m.currentResultIdx)
	}

	// Navigate backward with 'p'
	mm, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = mm.(Model)
	if m.currentResultIdx != 1 {
		t.Fatalf("expected currentResultIdx 1 after 'p', got %d", m.currentResultIdx)
	}
	if string(m.result.Payload) != "payload-two" {
		t.Fatalf("expected payload-two, got %q", string(m.result.Payload))
	}
}
