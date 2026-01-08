# Cloudflare DLP Forensic Copy Decoder

A handy command-line tool that decodes and extracts Cloudflare DLP (Data Loss Prevention) forensic copies from compressed log files.

## Overview

When Cloudflare DLP captures sensitive data, it creates forensic copies stored as compressed, base64-encoded JSON payloads. This tool automates the process of:
- Decompressing the log file
- Pretty-printing the JSON structure
- Decoding the base64-encoded payload
- Handling gzip-compressed payloads
- Extracting the original data in a human-readable format

## Features

- **Automatic Detection**: Intelligently detects content type and encoding from headers
- **Gzip Support**: Handles both plain and gzip-compressed payloads
- **JSON Formatting**: Pretty-prints all JSON output for easy reading
- **Interactive Mode**: Prompts to attempt JSON decoding for unsupported content types
- **Force Decode**: Optional `--try-json` flag to attempt decoding regardless of content type

## Prerequisites

The following tools must be installed and available in your PATH:
- `gzip` - for decompression
- `jq` - for JSON parsing and formatting
- `base64` - for base64 decoding

Most Linux/Unix systems have these pre-installed. On macOS, you may need to install `jq`:
```bash
brew install jq
```

## Installation

1. Clone this repository:
```bash
git clone https://github.com/yourusername/cloudflare-dlp-forensic-copy-decoder.git
cd cloudflare-dlp-forensic-copy-decoder
```

2. Make the script executable:
```bash
chmod +x cf-dlp-decode.sh
```

3. Optionally, add to your PATH or create a symlink:
```bash
sudo ln -s "$(pwd)/cf-dlp-decode.sh" /usr/local/bin/cf-dlp-decode
```

## Usage

### Basic Usage

```bash
./cf-dlp-decode.sh <input.log.gz>
```

### With Force JSON Decode

```bash
./cf-dlp-decode.sh --try-json <input.log.gz>
```

### Examples

**Example 1: Decode a standard DLP forensic copy**
```bash
./cf-dlp-decode.sh captured-data.log.gz
```

This will produce:
- `captured-data.log.json` - The decompressed and formatted log file
- `captured-data.json` - The extracted and decoded payload

**Example 2: Force JSON decoding for non-JSON content types**
```bash
./cf-dlp-decode.sh --try-json suspicious-upload.log.gz
```

**Example 3: View help**
```bash
./cf-dlp-decode.sh --help
```

## How It Works

1. **Decompression**: Extracts the `.log.gz` file to `.log.json`
2. **Formatting**: Pretty-prints the JSON structure using `jq`
3. **Header Analysis**: Reads `content-type` and `content-encoding` headers
4. **Payload Decoding**: 
   - Base64 decodes the `.Payload` field
   - If `content-encoding: gzip`, applies gzip decompression
   - Formats the final JSON output
5. **Output**: Writes the decoded payload to a separate file

### Supported Content Types

The tool automatically decodes payloads with:
- `content-type: application/json*` (any JSON content type)
- `content-encoding: gzip` (with automatic gzip decompression)

For other content types, use `--try-json` to attempt decoding, or respond to the interactive prompt.

## Output Files

Given an input file `example.log.gz`, the tool produces:

| File | Description |
|------|-------------|
| `example.log.json` | Decompressed and formatted log with headers and metadata |
| `example.json` | Extracted and decoded payload (the actual captured data) |

## Error Handling

The script will fail with descriptive error messages if:
- The input file doesn't exist or doesn't end with `.log.gz`
- Required dependencies are missing
- The payload isn't valid base64
- JSON decoding fails
- Gzip decompression fails

## Options

| Option | Description |
|--------|-------------|
| `--try-json` | Attempt JSON decoding even if content-type is not application/json |
| `-h, --help` | Display usage information |

## Interactive Mode

When the content type is not `application/json` and `--try-json` is not specified, the tool will prompt:
```
Try JSON decode anyway? [y/N]
```

Respond with `y` or `yes` to attempt decoding.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is provided as-is for use with Cloudflare DLP forensic analysis.

## Disclaimer

This tool is intended for legitimate forensic analysis and security investigation purposes only. Always ensure you have proper authorization before analyzing any data.
