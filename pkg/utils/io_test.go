package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile_ReadFileContent_FileExistsAndOverwrite(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "out.txt")

	data := []byte("hello world")
	if err := WriteFile(out, data, false); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Ensure file exists and content matches
	got, err := ReadFileContent(out)
	if err != nil {
		t.Fatalf("ReadFileContent failed: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("content mismatch: got %q want %q", string(got), string(data))
	}

	// Attempt to write without overwrite should fail
	if err := WriteFile(out, []byte("x"), false); err == nil {
		t.Fatalf("expected error when writing existing file without overwrite")
	}

	// Overwrite should succeed
	if err := WriteFile(out, []byte("new"), true); err != nil {
		t.Fatalf("WriteFile overwrite failed: %v", err)
	}
	got, err = ReadFileContent(out)
	if err != nil {
		t.Fatalf("ReadFileContent failed: %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("expected overwritten content, got %q", string(got))
	}
}

func TestReadInput_FileAndStdin(t *testing.T) {
	tmp := t.TempDir()
	inPath := filepath.Join(tmp, "in.txt")
	if err := os.WriteFile(inPath, []byte("abc"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	r, err := ReadInput(inPath)
	if err != nil {
		t.Fatalf("ReadInput(file) failed: %v", err)
	}

	// should be an *os.File for files
	if _, ok := r.(*os.File); !ok {
		t.Fatalf("expected *os.File from ReadInput(file), got %T", r)
	}

	// For '-' should return os.Stdin
	r2, err := ReadInput("-")
	if err != nil {
		t.Fatalf("ReadInput(-) failed: %v", err)
	}
	if r2 != os.Stdin {
		t.Fatalf("expected os.Stdin for ReadInput('-'), got %T", r2)
	}
}

func TestWriteAndRead_Integration_WithGzipPayload(t *testing.T) {
	// Ensure WriteFile can write a JSON/log and ReadFileContent can read it back
	tmp := t.TempDir()
	inPath := filepath.Join(tmp, "input.log.gz")

	// create a gzipped log with base64 payload
	payload := []byte("payload-data")
	encoded := base64.StdEncoding.EncodeToString(payload)
	entry := struct {
		Payload string            `json:"Payload"`
		Headers map[string]string `json:"Headers"`
	}{Payload: encoded, Headers: map[string]string{"content-type": "text/plain"}}

	j, _ := json.Marshal(entry)
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(j); err != nil {
		t.Fatalf("gzip write failed: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close failed: %v", err)
	}

	if err := WriteFile(inPath, buf.Bytes(), false); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	got, err := ReadFileContent(inPath)
	if err != nil {
		t.Fatalf("ReadFileContent failed: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected content in written gz file")
	}
}
