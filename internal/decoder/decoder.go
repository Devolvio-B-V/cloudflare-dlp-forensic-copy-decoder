// Package decoder implements the Cloudflare DLP forensic copy decoding logic.
// It handles base64 decoding, gzip decompression, and content-type detection
// to extract payloads from DLP forensic log files.
package decoder

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Compression magic numbers
const (
	gzipMagic1  = 0x1f
	gzipMagic2  = 0x8b
	zlibMagic1  = 0x78
	zlibMagic2a = 0x01
	zlibMagic2b = 0x9c
	zlibMagic2c = 0xda
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

// DecodeLogFileMulti decodes a gzip-compressed DLP forensic log file that may contain
// one or more concatenated JSON objects. It returns one DecodeResult per JSON object found.
func DecodeLogFileMulti(gzipData io.Reader, opts DecodeOptions) ([]*DecodeResult, error) {
	// Step 1: Decompress the gzip data
	gzReader, err := gzip.NewReader(gzipData)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gzReader.Close() }()

	// Step 2: Stream-parse JSON objects; a single file may contain multiple
	// concatenated top-level objects (one per forensic copy payload).
	dec := json.NewDecoder(gzReader)

	var results []*DecodeResult
	for dec.More() {
		// Capture each top-level object as raw bytes so we can both unmarshal
		// it into a LogEntry and pretty-print it faithfully.
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, fmt.Errorf("failed to parse log JSON: %w", err)
		}

		var entry LogEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			return nil, fmt.Errorf("failed to parse log JSON: %w", err)
		}

		result, err := decodeEntryPayload(entry, raw, opts)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no JSON objects found in log file")
	}

	return results, nil
}

// DecodeLogFile decodes a gzip-compressed DLP forensic log file.
// It decompresses, parses JSON, and decodes the payload according to content-type headers.
// For files that contain multiple concatenated JSON objects, use DecodeLogFileMulti instead.
func DecodeLogFile(gzipData io.Reader, opts DecodeOptions) (*DecodeResult, error) {
	results, err := DecodeLogFileMulti(gzipData, opts)
	if err != nil {
		return nil, err
	}
	return results[0], nil
}

// decodeEntryPayload builds a DecodeResult from a parsed LogEntry and its raw JSON bytes.
func decodeEntryPayload(entry LogEntry, rawJSON []byte, opts DecodeOptions) (*DecodeResult, error) {
	// Pretty-print the log JSON
	var prettyBuf bytes.Buffer
	if err := json.Indent(&prettyBuf, rawJSON, "", "  "); err != nil {
		return nil, fmt.Errorf("failed to pretty-print JSON: %w", err)
	}

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

	// Handle gzip/deflate encoded payloads or try automatic detection
	switch contentEncoding {
	case "gzip":
		payloadBytes, err = decompressGzip(payloadBytes)
		if err != nil {
			return nil, fmt.Errorf("gzip decode failed (is payload really gzipped?): %w", err)
		}
	case "deflate":
		payloadBytes, err = decompressDeflate(payloadBytes)
		if err != nil {
			return nil, fmt.Errorf("deflate decode failed (is payload really deflated?): %w", err)
		}
	case "":
		// No content-encoding header, try to detect compression automatically
		payloadBytes = tryDecompression(payloadBytes)
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
	} else if strings.HasPrefix(contentType, "application/xml") || strings.HasPrefix(contentType, "text/xml") {
		// XML content types
		result.IsText = true
		result.Payload = payloadBytes
	} else if strings.HasPrefix(contentType, "text/html") {
		// HTML content
		result.IsText = true
		result.Payload = payloadBytes
	} else if strings.HasPrefix(contentType, "text/csv") || strings.HasPrefix(contentType, "application/csv") {
		// CSV content
		result.IsText = true
		result.Payload = payloadBytes
	} else if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		// URL-encoded form data
		result.IsText = true
		result.Payload = payloadBytes
	} else if strings.HasPrefix(contentType, "application/javascript") || strings.HasPrefix(contentType, "text/javascript") {
		// JavaScript content
		result.IsText = true
		result.Payload = payloadBytes
	} else if strings.HasPrefix(contentType, "application/typescript") || strings.HasPrefix(contentType, "text/typescript") {
		// TypeScript content
		result.IsText = true
		result.Payload = payloadBytes
	} else if strings.HasPrefix(contentType, "text/plain") || strings.HasPrefix(contentType, "multipart/form-data") {
		result.IsText = true
		result.Payload = payloadBytes
	} else if strings.HasPrefix(contentType, "text/") {
		// Catch-all for any text/* content type
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
	defer func() { _ = reader.Close() }()

	return io.ReadAll(reader)
}

// decompressDeflate decompresses deflate-encoded data
func decompressDeflate(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	return io.ReadAll(reader)
}

// tryDecompression attempts to detect and decompress data automatically
// It tries gzip and deflate in order, returning the original data if neither works
func tryDecompression(data []byte) []byte {
	// Check for gzip magic number
	if len(data) >= 2 && data[0] == gzipMagic1 && data[1] == gzipMagic2 {
		if decompressed, err := decompressGzip(data); err == nil {
			return decompressed
		}
	}

	// Check for zlib/deflate magic number
	if len(data) >= 2 && data[0] == zlibMagic1 &&
		(data[1] == zlibMagic2a || data[1] == zlibMagic2b || data[1] == zlibMagic2c) {
		if decompressed, err := decompressDeflate(data); err == nil {
			return decompressed
		}
	}

	// Return original data if no compression detected or decompression failed
	return data
}

// validateJSON checks if data is valid JSON
func validateJSON(data []byte) error {
	var js interface{}
	return json.Unmarshal(data, &js)
}

// GetOutputFilename returns the appropriate output filename based on content type
func GetOutputFilename(inputPath string, isJSON bool) string {
	// Remove .log.gz extension
	baseName := strings.TrimSuffix(inputPath, ".log.gz")

	// Return appropriate extension
	if isJSON {
		return baseName + ".payload.json"
	}
	return baseName + ".payload.txt"
}

// GetOutputFilenameN returns the numbered output filename for multi-payload files (1-based index).
func GetOutputFilenameN(inputPath string, isJSON bool, n int) string {
	baseName := strings.TrimSuffix(inputPath, ".log.gz")
	if isJSON {
		return fmt.Sprintf("%s.%d.payload.json", baseName, n)
	}
	return fmt.Sprintf("%s.%d.payload.txt", baseName, n)
}

// GetLogJSONFilename returns the log JSON output filename
func GetLogJSONFilename(inputPath string) string {
	baseName := strings.TrimSuffix(inputPath, ".log.gz")
	return baseName + ".log.json"
}

// GetLogJSONFilenameN returns the numbered log JSON output filename for multi-payload files (1-based index).
func GetLogJSONFilenameN(inputPath string, n int) string {
	baseName := strings.TrimSuffix(inputPath, ".log.gz")
	return fmt.Sprintf("%s.%d.log.json", baseName, n)
}
