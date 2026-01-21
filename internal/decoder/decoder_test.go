package decoder

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

// Helper function to create a gzipped log file for testing
func createTestLogFile(t *testing.T, payload string, contentType string, contentEncoding string, gzipPayload bool) []byte {
	t.Helper()

	// Encode payload
	payloadBytes := []byte(payload)
	if gzipPayload {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		if _, err := gw.Write(payloadBytes); err != nil {
			t.Fatalf("failed to gzip payload: %v", err)
		}
		if err := gw.Close(); err != nil {
			t.Fatalf("failed to close gzip writer: %v", err)
		}
		payloadBytes = buf.Bytes()
	}

	// Create log entry
	entry := LogEntry{
		Payload: base64.StdEncoding.EncodeToString(payloadBytes),
		Headers: map[string]string{
			"content-type": contentType,
		},
	}
	if contentEncoding != "" {
		entry.Headers["content-encoding"] = contentEncoding
	}

	// Marshal to JSON
	logJSON, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal log entry: %v", err)
	}

	// Gzip the log file
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(logJSON); err != nil {
		t.Fatalf("failed to gzip log file: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	return buf.Bytes()
}

func TestDecodeLogFile_JSON(t *testing.T) {
	payload := `{"key": "value", "number": 42}`
	logData := createTestLogFile(t, payload, "application/json", "", false)

	opts := DecodeOptions{}
	result, err := DecodeLogFile(bytes.NewReader(logData), opts)
	if err != nil {
		t.Fatalf("DecodeLogFile failed: %v", err)
	}

	if !result.IsJSON {
		t.Error("Expected IsJSON to be true")
	}
	if result.IsText {
		t.Error("Expected IsText to be false")
	}

	// Verify the payload is valid JSON and matches
	var decoded map[string]interface{}
	if err := json.Unmarshal(result.Payload, &decoded); err != nil {
		t.Fatalf("Decoded payload is not valid JSON: %v", err)
	}

	if decoded["key"] != "value" {
		t.Errorf("Expected key=value, got %v", decoded["key"])
	}
	if decoded["number"].(float64) != 42 {
		t.Errorf("Expected number=42, got %v", decoded["number"])
	}
}

func TestDecodeLogFile_PlainText(t *testing.T) {
	payload := "Hello, World! This is plain text."
	logData := createTestLogFile(t, payload, "text/plain", "", false)

	opts := DecodeOptions{}
	result, err := DecodeLogFile(bytes.NewReader(logData), opts)
	if err != nil {
		t.Fatalf("DecodeLogFile failed: %v", err)
	}

	if result.IsJSON {
		t.Error("Expected IsJSON to be false")
	}
	if !result.IsText {
		t.Error("Expected IsText to be true")
	}

	if string(result.Payload) != payload {
		t.Errorf("Expected payload %q, got %q", payload, string(result.Payload))
	}
}

func TestDecodeLogFile_GzippedJSON(t *testing.T) {
	payload := `{"gzipped": true, "data": "compressed"}`
	logData := createTestLogFile(t, payload, "application/json", "gzip", true)

	opts := DecodeOptions{}
	result, err := DecodeLogFile(bytes.NewReader(logData), opts)
	if err != nil {
		t.Fatalf("DecodeLogFile failed: %v", err)
	}

	if !result.IsJSON {
		t.Error("Expected IsJSON to be true")
	}

	// Verify the payload is valid JSON and matches
	var decoded map[string]interface{}
	if err := json.Unmarshal(result.Payload, &decoded); err != nil {
		t.Fatalf("Decoded payload is not valid JSON: %v", err)
	}

	if decoded["gzipped"] != true {
		t.Errorf("Expected gzipped=true, got %v", decoded["gzipped"])
	}
}

func TestDecodeLogFile_GzippedText(t *testing.T) {
	payload := "This text is gzipped!"
	logData := createTestLogFile(t, payload, "text/plain", "gzip", true)

	opts := DecodeOptions{}
	result, err := DecodeLogFile(bytes.NewReader(logData), opts)
	if err != nil {
		t.Fatalf("DecodeLogFile failed: %v", err)
	}

	if !result.IsText {
		t.Error("Expected IsText to be true")
	}

	if string(result.Payload) != payload {
		t.Errorf("Expected payload %q, got %q", payload, string(result.Payload))
	}
}

func TestDecodeLogFile_FormData(t *testing.T) {
	payload := "field1=value1&field2=value2"
	logData := createTestLogFile(t, payload, "multipart/form-data", "", false)

	opts := DecodeOptions{}
	result, err := DecodeLogFile(bytes.NewReader(logData), opts)
	if err != nil {
		t.Fatalf("DecodeLogFile failed: %v", err)
	}

	if !result.IsText {
		t.Error("Expected IsText to be true")
	}

	if string(result.Payload) != payload {
		t.Errorf("Expected payload %q, got %q", payload, string(result.Payload))
	}
}

func TestDecodeLogFile_UnsupportedContentType(t *testing.T) {
	payload := "some binary data"
	logData := createTestLogFile(t, payload, "application/octet-stream", "", false)

	opts := DecodeOptions{}
	_, err := DecodeLogFile(bytes.NewReader(logData), opts)
	if err == nil {
		t.Error("Expected error for unsupported content type")
	}
	if !strings.Contains(err.Error(), "content-type not supported") {
		t.Errorf("Expected content-type error, got: %v", err)
	}
}

func TestDecodeLogFile_UnsupportedContentTypeWithTryText(t *testing.T) {
	payload := "some binary data"
	logData := createTestLogFile(t, payload, "application/octet-stream", "", false)

	opts := DecodeOptions{TryText: true}
	result, err := DecodeLogFile(bytes.NewReader(logData), opts)
	if err != nil {
		t.Fatalf("DecodeLogFile failed with TryText: %v", err)
	}

	if !result.IsText {
		t.Error("Expected IsText to be true with TryText")
	}

	if string(result.Payload) != payload {
		t.Errorf("Expected payload %q, got %q", payload, string(result.Payload))
	}
}

func TestDecodeLogFile_EmptyPayload(t *testing.T) {
	entry := LogEntry{
		Payload: "",
		Headers: map[string]string{
			"content-type": "application/json",
		},
	}

	logJSON, _ := json.Marshal(entry)
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write(logJSON)
	gw.Close()

	opts := DecodeOptions{}
	_, err := DecodeLogFile(bytes.NewReader(buf.Bytes()), opts)
	if err == nil {
		t.Error("Expected error for empty payload")
	}
	if !strings.Contains(err.Error(), "no .Payload found") {
		t.Errorf("Expected empty payload error, got: %v", err)
	}
}

func TestDecodeLogFile_InvalidBase64(t *testing.T) {
	entry := LogEntry{
		Payload: "not-valid-base64!@#$",
		Headers: map[string]string{
			"content-type": "application/json",
		},
	}

	logJSON, _ := json.Marshal(entry)
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write(logJSON)
	gw.Close()

	opts := DecodeOptions{}
	_, err := DecodeLogFile(bytes.NewReader(buf.Bytes()), opts)
	if err == nil {
		t.Error("Expected error for invalid base64")
	}
	if !strings.Contains(err.Error(), "base64 decode failed") {
		t.Errorf("Expected base64 error, got: %v", err)
	}
}

func TestDecodeLogFile_InvalidJSON(t *testing.T) {
	payload := `{"invalid": json}`
	logData := createTestLogFile(t, payload, "application/json", "", false)

	opts := DecodeOptions{}
	_, err := DecodeLogFile(bytes.NewReader(logData), opts)
	if err == nil {
		t.Error("Expected error for invalid JSON payload")
	}
	if !strings.Contains(err.Error(), "json decode failed") {
		t.Errorf("Expected JSON error, got: %v", err)
	}
}

func TestGetOutputFilename(t *testing.T) {
	tests := []struct {
		input    string
		isJSON   bool
		expected string
	}{
		{"test.log.gz", true, "test.payload.json"},
		{"test.log.gz", false, "test.payload.txt"},
		{"path/to/test.log.gz", true, "path/to/test.payload.json"},
		{"test", true, "test.payload.json"},
	}

	for _, tt := range tests {
		result := GetOutputFilename(tt.input, tt.isJSON)
		if result != tt.expected {
			t.Errorf("GetOutputFilename(%q, %v) = %q, want %q", tt.input, tt.isJSON, result, tt.expected)
		}
	}
}

func TestGetLogJSONFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test.log.gz", "test.log.json"},
		{"path/to/test.log.gz", "path/to/test.log.json"},
		{"test", "test.log.json"},
	}

	for _, tt := range tests {
		result := GetLogJSONFilename(tt.input)
		if result != tt.expected {
			t.Errorf("GetLogJSONFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetHeader_CaseInsensitive(t *testing.T) {
	headers := map[string]string{
		"Content-Type":     "application/json",
		"content-encoding": "gzip",
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"content-type", "application/json"},
		{"Content-Type", "application/json"},
		{"CONTENT-TYPE", "application/json"},
		{"content-encoding", "gzip"},
		{"Content-Encoding", "gzip"},
		{"missing", ""},
	}

	for _, tt := range tests {
		result := getHeader(headers, tt.key)
		if result != tt.expected {
			t.Errorf("getHeader(%q) = %q, want %q", tt.key, result, tt.expected)
		}
	}
}
