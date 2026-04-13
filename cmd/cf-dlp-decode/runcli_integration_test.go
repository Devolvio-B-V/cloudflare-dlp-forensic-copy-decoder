package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// helper to create gzipped log file bytes
func makeGzippedLog(t *testing.T, payload []byte, contentType string, contentEncoding string, gzipPayload bool) []byte {
	t.Helper()

	pl := payload
	if gzipPayload {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		if _, err := gw.Write(pl); err != nil {
			t.Fatalf("failed to gzip payload: %v", err)
		}
		if err := gw.Close(); err != nil {
			t.Fatalf("failed to close gzip writer: %v", err)
		}
		pl = buf.Bytes()
	}

	entry := struct {
		Payload string            `json:"Payload"`
		Headers map[string]string `json:"Headers"`
	}{
		Payload: base64.StdEncoding.EncodeToString(pl),
		Headers: map[string]string{"content-type": contentType},
	}
	if contentEncoding != "" {
		entry.Headers["content-encoding"] = contentEncoding
	}

	logJSON, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal entry: %v", err)
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(logJSON); err != nil {
		t.Fatalf("failed to gzip log: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	return buf.Bytes()
}

func TestRunCLI_WritesFiles(t *testing.T) {
	tmp := t.TempDir()

	payload := []byte("hello-from-runcli")
	data := makeGzippedLog(t, payload, "text/plain", "", false)

	inPath := filepath.Join(tmp, "input.log.gz")
	if err := os.WriteFile(inPath, data, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	outPath := filepath.Join(tmp, "out.payload.txt")

	if err := runCLI(inPath, outPath, false, true, false); err != nil {
		t.Fatalf("runCLI failed: %v", err)
	}

	// Check files exist
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected payload file at %s, missing: %v", outPath, err)
	}
	logPath := filepath.Join(tmp, "input.log.json")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected log JSON at %s, missing: %v", logPath, err)
	}
}

func TestRunCLI_TryTextFlag(t *testing.T) {
	tmp := t.TempDir()

	payload := []byte("binary-but-we-force-text")
	data := makeGzippedLog(t, payload, "application/octet-stream", "", false)

	inPath := filepath.Join(tmp, "input2.log.gz")
	if err := os.WriteFile(inPath, data, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	outPath := filepath.Join(tmp, "out2.payload.txt")

	if err := runCLI(inPath, outPath, true, true, false); err != nil {
		t.Fatalf("runCLI with tryText failed: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected payload file at %s, missing: %v", outPath, err)
	}
}

func TestRunCLI_NoOverwriteFails(t *testing.T) {
	tmp := t.TempDir()

	payload := []byte("hello")
	data := makeGzippedLog(t, payload, "text/plain", "", false)

	inPath := filepath.Join(tmp, "input3.log.gz")
	if err := os.WriteFile(inPath, data, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	outPath := filepath.Join(tmp, "out3.payload.txt")
	// Create existing file
	if err := os.WriteFile(outPath, []byte("exists"), 0644); err != nil {
		t.Fatalf("failed to create existing out file: %v", err)
	}

	if err := runCLI(inPath, outPath, false, false, false); err == nil {
		t.Fatalf("expected error when overwrite=false and output exists")
	}
}

// makePlainJSONLog creates an uncompressed JSON log file (as Cloudflare sometimes provides)
func makePlainJSONLog(t *testing.T, payload []byte, contentType string) []byte {
	t.Helper()

	entry := struct {
		Payload string            `json:"Payload"`
		Headers map[string]string `json:"Headers"`
	}{
		Payload: base64.StdEncoding.EncodeToString(payload),
		Headers: map[string]string{"content-type": contentType},
	}

	logJSON, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal entry: %v", err)
	}

	return logJSON
}

func TestRunCLI_PlainJSONInput_WritesFiles(t *testing.T) {
	tmp := t.TempDir()

	payload := []byte("hello-from-json-input")
	data := makePlainJSONLog(t, payload, "text/plain")

	inPath := filepath.Join(tmp, "20260407T141820Z_abc123.json")
	if err := os.WriteFile(inPath, data, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	outPath := filepath.Join(tmp, "out.payload.txt")

	if err := runCLI(inPath, outPath, false, true, false); err != nil {
		t.Fatalf("runCLI with plain JSON input failed: %v", err)
	}

	// Check payload file exists and has correct content
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected payload file at %s, missing: %v", outPath, err)
	}
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(content) != string(payload) {
		t.Errorf("expected payload %q, got %q", string(payload), string(content))
	}

	// Check log JSON file exists
	logPath := filepath.Join(tmp, "20260407T141820Z_abc123.log.json")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected log JSON at %s, missing: %v", logPath, err)
	}
}
