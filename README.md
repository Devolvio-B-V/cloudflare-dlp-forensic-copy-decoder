# Cloudflare DLP Forensic Copy Decoder

[![CI](https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/actions/workflows/ci.yml/badge.svg)](https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder)](https://goreportcard.com/report/github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder)

A powerful command-line tool and interactive TUI for decoding and extracting Cloudflare DLP (Data Loss Prevention) forensic copies from compressed log files.

## ✨ Features

- 🚀 **Native Go Implementation**: Single binary with no external dependencies
- 🎨 **Interactive TUI**: Terminal user interface built with [bubbletea](https://github.com/charmbracelet/bubbletea)
- 📦 **Cross-Platform**: Pre-built binaries for Linux, macOS, and Windows (amd64 & arm64)
- 🔄 **Full Feature Parity**: 100% compatible with the original shell script
- ✅ **Well Tested**: Comprehensive unit tests with high coverage
- 🎯 **Easy to Use**: Simple CLI for scripting and automation

### What It Does

When Cloudflare DLP captures sensitive data, it creates forensic copies stored as compressed, base64-encoded JSON payloads. This tool automates:

1. **Decompression**: Extracts the `.log.gz` file
2. **JSON Formatting**: Pretty-prints the log structure
3. **Payload Decoding**: Base64 decodes the payload
4. **Gzip Handling**: Automatically decompresses gzipped payloads
5. **Smart Detection**: Intelligently detects and processes content based on headers

## 📦 Installation

### Via Package Managers

**Homebrew (macOS and Linux)**
```bash
brew install devolvio-b-v/tap/cf-dlp-decode
```

**Winget (Windows)**
```bash
winget install Devolvio-B-V.cf-dlp-decode
```

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/releases):

```bash
# Linux (amd64)
curl -L -o cf-dlp-decode https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/releases/latest/download/cf-dlp-decode-linux-amd64
chmod +x cf-dlp-decode
sudo mv cf-dlp-decode /usr/local/bin/

# macOS (arm64)
curl -L -o cf-dlp-decode https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/releases/latest/download/cf-dlp-decode-darwin-arm64
chmod +x cf-dlp-decode
sudo mv cf-dlp-decode /usr/local/bin/

# Windows (amd64)
# Download cf-dlp-decode-windows-amd64.exe and add to PATH
```

### Build from Source

Requirements:
- Go 1.21 or higher

```bash
git clone https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder.git
cd cloudflare-dlp-forensic-copy-decoder
make build

# Or install directly
go install github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/cmd/cf-dlp-decode@latest
```

## 🚀 Usage

### Interactive TUI Mode

Launch the terminal user interface for an interactive experience:

```bash
# Launch TUI with file browser (no arguments)
cf-dlp-decode

# Launch TUI with a specific file
cf-dlp-decode --tui captured-data.log.gz
```

Note: Running `cf-dlp-decode` with no arguments now launches the interactive TUI with a file browser by default.
Provide a filename (positional argument or `--input`) to run the tool in non-interactive CLI mode.


**TUI Features:**
- **Interactive file browser** with vim-style navigation
- **Lazygit-inspired UI** with color-coded interface
- Visual preview of decoded content with scrolling
- Easy file navigation and selection
- Export decoded payloads
- Keyboard shortcuts for all operations

**Keyboard Controls:**

*File Browser Mode:*
- `↑↓` / `j/k` - Navigate up/down
- `Enter` - Open directory or select .log.gz file
- `h` - Go to home directory
- `g` / `G` - Go to top/bottom of list
- `q` - Quit

*Preview Mode:*
- `↑↓` / `j/k` - Scroll up/down
- `d` / `u` - Page down/up (half page)
- `g` / `G` - Go to top/bottom
- `n` / `p` - Next/previous payload (when file contains multiple payloads)
- `s` / `e` - Save/Export decoded payload
- `b` - Back to file browser
- `q` - Quit

### Non-Interactive CLI Mode

Perfect for scripts and automation:

```bash
# Basic usage
cf-dlp-decode captured-data.log.gz

# With custom output path
cf-dlp-decode --input captured-data.log.gz --output decoded.json

# Force text decoding for unsupported content types
cf-dlp-decode --try-text suspicious-upload.log.gz

# Read from stdin
cat captured-data.log.gz | cf-dlp-decode --input -

# Verbose output
cf-dlp-decode --verbose captured-data.log.gz

# Overwrite existing files (flags can be before or after filename)
cf-dlp-decode --overwrite captured-data.log.gz
cf-dlp-decode captured-data.log.gz --overwrite
```

## 📝 Command-Line Options

| Option | Description |
|--------|-------------|
| `--input PATH` | Input .log.gz file (or - for stdin) |
| `--output PATH` | Output file path (auto-generated if not specified) |
| `--tui` | Launch interactive TUI mode |
| `--try-text` | Attempt text decoding even if content-type is unsupported |
| `--overwrite` | Overwrite output files if they exist |
| `--verbose` | Enable verbose output with detailed error messages |
| `--help` | Display help information |
| `--version` | Show version information |

## 📂 Output Files

Given an input file `example.log.gz`, the tool produces:

| File | Description |
|------|-------------|
| `example.log.json` | Decompressed and formatted log with headers and metadata |
| `example.payload.json` | Decoded payload (for JSON content types) |
| `example.payload.txt` | Decoded payload (for text/form-data content types) |

### Multi-Payload Files

When a `.log.gz` file contains multiple concatenated JSON objects (e.g. when Cloudflare batches payloads at the same timestamp), the tool writes numbered output files:

| File | Description |
|------|-------------|
| `example.1.log.json` | Log metadata for payload 1 |
| `example.1.payload.json` | Decoded payload 1 (JSON) |
| `example.2.log.json` | Log metadata for payload 2 |
| `example.2.payload.json` | Decoded payload 2 (JSON) |
| … | … |

> **Note:** When processing a multi-payload file, the `--output` flag is ignored and a warning is printed. Use numbered output files instead.

## 🔍 Supported Content Types

The tool automatically handles:

- **JSON**: `application/json*` → `.payload.json`
- **XML**: `application/xml*`, `text/xml*` → `.payload.txt`
- **HTML**: `text/html*` → `.payload.txt`
- **Plain Text**: `text/plain*` → `.payload.txt`
- **CSV**: `text/csv*`, `application/csv*` → `.payload.txt`
- **JavaScript**: `application/javascript*`, `text/javascript*` → `.payload.txt`
- **TypeScript**: `application/typescript*`, `text/typescript*` → `.payload.txt`
- **Form Data**: `multipart/form-data*`, `application/x-www-form-urlencoded*` → `.payload.txt`
- **Generic Text**: Any `text/*` content type → `.payload.txt`

### Compression Support

The tool automatically detects and handles multiple compression formats:

- **Gzip**: `content-encoding: gzip` or automatic detection via magic number
- **Deflate**: `content-encoding: deflate` or automatic detection via magic number
- **Automatic Detection**: Even without content-encoding headers, the tool tries to detect compressed payloads

For other content types, use `--try-text` to force text decoding.

## 🛠️ Development

### Building

```bash
# Build for current platform
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter
make vet

# Cross-compile for all platforms
make cross-build

# Clean build artifacts
make clean
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...
```

### Project Structure

```
.
├── cmd/cf-dlp-decode/    # CLI entry point
├── internal/
│   ├── decoder/          # Core decoding logic
│   └── ui/               # TUI implementation
├── pkg/utils/            # Utility functions
├── legacy/               # Original shell script
├── .github/workflows/    # CI/CD workflows
├── Makefile              # Build automation
├── go.mod                # Go module definition
└── README.md             # This file
```

## 🔄 Migration from Shell Script

The original shell script has been moved to the `legacy/` directory and is still available for use. The Go implementation provides:

✅ **Same behavior** - Identical decoding algorithm and output format  
✅ **Same flags** - Compatible command-line interface  
✅ **Better performance** - Faster execution with native code  
✅ **No dependencies** - No need for jq, gzip, or base64 tools  
✅ **Enhanced features** - Interactive TUI mode and better error handling  

See [legacy/README.md](/legacy/README.md) for more details on differences and migration notes.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Write tests for new features
- Ensure all tests pass (`make test`)
- Format code with `go fmt`
- Run `go vet` before committing
- Update documentation as needed

## 📜 License

This project is provided as-is for use with Cloudflare DLP forensic analysis.

## 📦 Package Manager Publishing

This project is automatically published to multiple package managers via GitHub Actions on each release:

- **Homebrew**: Automatically updated via GoReleaser to the [homebrew-tap](https://github.com/Devolvio-B-V/homebrew-tap) repository
- **Winget**: Automatically submitted via wingetcreate (requires fork of [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs))

## ⚠️ Disclaimer

This tool is intended for legitimate forensic analysis and security investigation purposes only. Always ensure you have proper authorization before analyzing any data.

## 🐛 Issues and Support

If you encounter any issues or have questions:

1. Check the [existing issues](https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/issues)
2. Create a new issue with:
   - Clear description of the problem
   - Steps to reproduce
   - Expected vs actual behavior
   - System information (OS, Go version)

## 🔗 Links

- **Repository**: https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder
- **Issues**: https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/issues
- **Releases**: https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/releases

## 🙏 Acknowledgments

- Built with [bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI
- Inspired by the need for better DLP forensic tooling
