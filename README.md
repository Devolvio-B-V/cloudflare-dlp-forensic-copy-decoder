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
cf-dlp-decode --tui captured-data.log.gz
```

**TUI Features:**
- Visual preview of decoded content
- Easy file navigation
- Export decoded payloads
- Toggle between formatted and raw views
- Keyboard shortcuts for all operations

**Keyboard Controls:**
- `Enter` - Decode file
- `s/e` - Save/Export decoded payload
- `r` - Toggle raw/formatted view
- `o` - Open new file
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

# Overwrite existing files
cf-dlp-decode --overwrite captured-data.log.gz
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

## 🔍 Supported Content Types

The tool automatically handles:

- **JSON**: `application/json*` → `.payload.json`
- **Plain Text**: `text/plain*` → `.payload.txt`
- **Form Data**: `multipart/form-data*` → `.payload.txt`
- **Gzipped**: `content-encoding: gzip` (automatic decompression)

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

See `legacy/README.md` for more details on differences and migration notes.

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
- **Winget**: Semi-automated submission to [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs)

### Setup Instructions

For detailed setup instructions, see [PACKAGE_MANAGERS_SETUP.md](PACKAGE_MANAGERS_SETUP.md).

**Quick setup:**
1. Create `homebrew-tap` repository (GoReleaser will populate it)
2. Generate GitHub Personal Access Token with `repo` scope
3. Add repository secret: `HOMEBREW_TAP_GITHUB_TOKEN`
4. For Winget: Use `wingetcreate` to submit updates (see setup guide)

### Required Repository Secrets

Configure these in repository Settings → Secrets and variables → Actions:
- `HOMEBREW_TAP_GITHUB_TOKEN`: Personal access token with write access to homebrew-tap repository

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
