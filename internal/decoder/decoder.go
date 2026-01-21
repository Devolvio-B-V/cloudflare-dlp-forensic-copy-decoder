// Package decoder implements the Cloudflare DLP forensic copy decoding logic.
// It handles base64 decoding, gzip decompression, and content-type detection
// to extract payloads from DLP forensic log files.
package decoder

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// LogEntry represents the structure of a DLP forensic log file
type LogEntry struct {
	Payload string            `json:"Payload"`
	Headers map[string]string `json:"Headers"`
}

// DecodeResult contains the decoded payload and metadata
type DecodeResult struct {
	// LogJSON is the pretty-printed log entry
	LogJSON string
	// Payload is the decoded payload content
	Payload []byte
	// ContentType is the detected content type
	ContentType string
	// IsJSON indicates if the payload is JSON
	IsJSON bool
	// IsText indicates if the payload is text
	IsText bool
}

// DecodeOptions configures the decoding behavior
type DecodeOptions struct {
	// TryText forces text decoding even for unsupported content types
	TryText bool
	// Verbose enables detailed error messages
	Verbose bool
}

// DecodeLogFile decodes a gzip-compressed DLP forensic log file.
// It decompresses, parses JSON, and decodes the payload according to content-type headers.
func DecodeLogFile(gzipData io.Reader, opts DecodeOptions) (*DecodeResult, error) {
	// Step 1: Decompress the gzip data
	gzReader, err := gzip.NewReader(gzipData)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	logData, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress log file: %w", err)
	}

	// Step 2: Parse the JSON log entry
	var entry LogEntry
	if err := json.Unmarshal(logData, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse log JSON: %w", err)
	}

	// Step 3: Pretty-print the log JSON
	var prettyBuf bytes.Buffer
	if err := json.Indent(&prettyBuf, logData, "", "  "); err != nil {
		return nil, fmt.Errorf("failed to pretty-print JSON: %w", err)
	}

	// Step 4: Decode the payload
	result := &DecodeResult{
		LogJSON: prettyBuf.String(),
	}

	// Get content type and encoding from headers
	contentType := getHeader(entry.Headers, "content-type")
	contentEncoding := getHeader(entry.Headers, "content-encoding")
	result.ContentType = contentType

	// Decode base64 payload
	if entry.Payload == "" {
		return nil, fmt.Errorf("no .Payload found (or it was empty) in log file")
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(entry.Payload)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed (is .Payload valid base64?): %w", err)
	}

	// Handle gzip-encoded payloads
	if contentEncoding == "gzip" {
		payloadBytes, err = decompressGzip(payloadBytes)
		if err != nil {
			return nil, fmt.Errorf("gzip decode failed (is payload really gzipped?): %w", err)
		}
	}

	// Determine payload type and set flags
	if strings.HasPrefix(contentType, "application/json") {
		result.IsJSON = true
		// Validate and pretty-print JSON
		if err := validateJSON(payloadBytes); err != nil {
			return nil, fmt.Errorf("json decode failed (decoded payload wasn't valid JSON?): %w", err)
		}
		var prettyPayload bytes.Buffer
		if err := json.Indent(&prettyPayload, payloadBytes, "", "  "); err != nil {
			result.Payload = payloadBytes // Use raw if pretty-print fails
		} else {
			result.Payload = prettyPayload.Bytes()
		}
	} else if strings.HasPrefix(contentType, "text/plain") || strings.HasPrefix(contentType, "multipart/form-data") {
		result.IsText = true
		result.Payload = payloadBytes
	} else if opts.TryText {
		// Try text decode with --try-text flag
		result.IsText = true
		result.Payload = payloadBytes
	} else {
		return nil, fmt.Errorf("content-type not supported (got: %s). Use --try-text to force text decode", contentType)
	}

	return result, nil
}

// getHeader retrieves a header value case-insensitively
func getHeader(headers map[string]string, key string) string {
	// Try exact match first
	if val, ok := headers[key]; ok {
		return val
	}
	// Try case-insensitive match
	lowerKey := strings.ToLower(key)
	for k, v := range headers {
		if strings.ToLower(k) == lowerKey {
			return v
		}
	}
	return ""
}

// decompressGzip decompresses gzip-encoded data
func decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// validateJSON checks if data is valid JSON
func validateJSON(data []byte) error {
	var js interface{}
	return json.Unmarshal(data, &js)
}

// GetOutputFilename returns the appropriate output filename based on content type
func GetOutputFilename(inputPath string, isJSON bool) string {
	// Remove .log.gz extension
	baseName := inputPath
	if strings.HasSuffix(baseName, ".log.gz") {
		baseName = strings.TrimSuffix(baseName, ".log.gz")
	}

	// Return appropriate extension
	if isJSON {
		return baseName + ".payload.json"
	}
	return baseName + ".payload.txt"
}

// GetLogJSONFilename returns the log JSON output filename
func GetLogJSONFilename(inputPath string) string {
	baseName := inputPath
	if strings.HasSuffix(baseName, ".log.gz") {
		baseName = strings.TrimSuffix(baseName, ".log.gz")
	}
	return baseName + ".log.json"
}
