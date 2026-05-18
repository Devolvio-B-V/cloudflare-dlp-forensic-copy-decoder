package ui

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/internal/decoder"
	tea "github.com/charmbracelet/bubbletea"
)

func makeGzippedLogBytes(t *testing.T, payload []byte, contentType string, gzipPayload bool) []byte {
	t.Helper()

	pl := payload
	if gzipPayload {
		var b bytes.Buffer
		gw := gzip.NewWriter(&b)
		if _, err := gw.Write(pl); err != nil {
			t.Fatalf("gzip payload write: %v", err)
		}
		if err := gw.Close(); err != nil {
			t.Fatalf("gzip payload close: %v", err)
		}
		pl = b.Bytes()
	}

	entry := decoder.LogEntry{
		Payload: base64.StdEncoding.EncodeToString(pl),
		Headers: map[string]string{"content-type": contentType},
	}

	j, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal entry: %v", err)
	}

	var out bytes.Buffer
	gw := gzip.NewWriter(&out)
	if _, err := gw.Write(j); err != nil {
		t.Fatalf("gzip write log: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close log: %v", err)
	}
	return out.Bytes()
}

func TestModel_DecodeCommand_Succeeds(t *testing.T) {
	data := makeGzippedLogBytes(t, []byte("hello-model-decode"), "text/plain", false)

	// Write to temp file
	tmp := t.TempDir()
	inPath := tmp + "/input.log.gz"
	if err := os.WriteFile(inPath, data, 0644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	m := NewModel(inPath)

	// call decode cmd directly
	cmd := m.decode()
	msg := cmd()
	switch v := msg.(type) {
	case decodeSuccessMsg:
		if len(v.results) == 0 {
			t.Fatalf("expected results in decodeSuccessMsg")
		}
	case decodeErrorMsg:
		t.Fatalf("decode failed: %v", v.err)
	default:
		t.Fatalf("unexpected msg type: %T", msg)
	}
}

func TestModel_Update_WindowSizeAndDecodeMsg(t *testing.T) {
	m := NewModel("")
	// Window size message should set ready and viewport dimensions
	mIface, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	mm := mIface.(Model)
	if mm.width != 100 || mm.height != 40 {
		t.Fatalf("expected width/height set, got %d/%d", mm.width, mm.height)
	}

	// simulate decode success
	res := &decoder.DecodeResult{Payload: []byte("x"), IsText: true}
	m2Iface, _ := mm.Update(decodeSuccessMsg{results: []*decoder.DecodeResult{res}})
	mm2 := m2Iface.(Model)
	if mm2.mode != ModePreview {
		t.Fatalf("expected ModePreview, got %v", mm2.mode)
	}
}

func TestHandleKeyPress_ErrorRetryWithTryText_Succeeds(t *testing.T) {
	data := makeGzippedLogBytes(t, []byte("hello-retry"), "application/octet-stream", false)
	tmp := t.TempDir()
	inPath := tmp + "/input.log.gz"
	if err := os.WriteFile(inPath, data, 0644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	m := Model{
		mode:      ModeError,
		inputPath: inPath,
		err:       decoder.ErrUnsupportedContentType,
	}

	updated, cmd := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	mm := updated.(Model)
	if mm.mode != ModeDecoding {
		t.Fatalf("expected mode %v after retry, got %v", ModeDecoding, mm.mode)
	}
	if !mm.tryText {
		t.Fatalf("expected tryText to be enabled after pressing 't'")
	}
	if cmd == nil {
		t.Fatalf("expected decode command on retry")
	}

	msg := cmd()
	if _, ok := msg.(decodeSuccessMsg); !ok {
		t.Fatalf("expected decodeSuccessMsg after retry, got %T", msg)
	}
}
