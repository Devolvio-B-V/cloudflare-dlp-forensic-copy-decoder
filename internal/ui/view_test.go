package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/internal/decoder"
	"github.com/charmbracelet/bubbles/viewport"
)

func TestViewFileBrowser_Renders(t *testing.T) {
	tmp := t.TempDir()
	// create a file so file browser isn't empty
	_ = os.WriteFile(filepath.Join(tmp, "a.log.gz"), []byte("x"), 0644)

	fb, err := NewFileBrowser(tmp)
	if err != nil {
		t.Fatalf("NewFileBrowser failed: %v", err)
	}

	m := Model{
		mode:        ModeFileBrowser,
		fileBrowser: fb,
		width:       80,
		height:      24,
	}

	out := m.viewFileBrowser()
	if out == "" {
		t.Fatalf("expected non-empty view output")
	}
}

func TestViewPreviewExportError_Renders(t *testing.T) {
	vp := viewport.New(40, 5)
	vp.SetContent("line1\nline2")

	m := Model{
		mode:      ModePreview,
		inputPath: "input.log.gz",
		result: &decoder.DecodeResult{
			ContentType: "text/plain",
			IsText:      true,
			Payload:     []byte("x"),
		},
		viewport: vp,
		width:    80,
		height:   24,
	}

	if out := m.viewPreview(); out == "" {
		t.Fatalf("expected preview output")
	}

	m.mode = ModeExport
	m.exportPath = ""
	if out := m.viewExport(); out == "" {
		t.Fatalf("expected export view output")
	}

	m.mode = ModeError
	m.err = errors.New("boom")
	if out := m.viewError(); out == "" {
		t.Fatalf("expected error view output")
	}
}

func TestViewError_ShowsTryTextRetryHelp(t *testing.T) {
	m := Model{
		mode:  ModeError,
		width: 80,
		err:   fmt.Errorf("%w (got: application/octet-stream). Use --try-text to force text decode", decoder.ErrUnsupportedContentType),
	}

	out := m.viewError()
	if !strings.Contains(out, "[t]") {
		t.Fatalf("expected retry hint in error view, got: %q", out)
	}
}
