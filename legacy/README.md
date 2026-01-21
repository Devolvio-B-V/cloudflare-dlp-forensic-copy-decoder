# Legacy Shell Script Implementation

This directory contains the original shell script implementation of the Cloudflare DLP Forensic Copy Decoder.

## Files

- `cf-dlp-decode.sh` - The original bash script implementation
- `README-original.md` - The original README documentation

## Why was this moved?

The project has been rewritten in Go with the following improvements:

1. **Native Binary**: No external dependencies (jq, gzip, base64) needed - everything is built into a single binary
2. **Interactive TUI**: A terminal user interface built with bubbletea for easier file browsing and decoding
3. **Cross-Platform**: Pre-built binaries for Linux, macOS, and Windows on both amd64 and arm64
4. **Better Error Handling**: More informative error messages and proper exit codes
5. **Unit Tests**: Comprehensive test coverage for the decoding logic
6. **Consistent Behavior**: Identical decoding algorithm and command-line interface

## Differences from Go Implementation

The Go implementation maintains full feature parity with the shell script:

- ✅ Same command-line flags: `--try-text`, `--help`
- ✅ Same input/output file conventions (`.log.gz` → `.log.json` + `.payload.json/.txt`)
- ✅ Same decoding algorithm: gzip decompression, base64 decoding, content-type detection
- ✅ Same error messages and exit codes
- ➕ **New**: Interactive TUI mode with `--tui` flag
- ➕ **New**: Additional flags for scripting: `--output`, `--overwrite`, `--verbose`

## When to use the legacy script

You may prefer the shell script if:

- You need to audit the exact decoding steps (simpler bash code)
- You're on a system where installing Go binaries is restricted
- You want to modify the script for custom workflows

## Running the legacy script

```bash
cd legacy
chmod +x cf-dlp-decode.sh
./cf-dlp-decode.sh <input.log.gz>
```

See `README-original.md` for full documentation.
