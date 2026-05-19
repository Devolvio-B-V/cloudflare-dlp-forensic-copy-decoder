package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/internal/decoder"
	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/internal/ui"
	"github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder/pkg/utils"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "2.3.0"

// boolFlags contains all boolean flags that don't take a value
var boolFlags = map[string]bool{
	"--tui":       true,
	"--try-text":  true,
	"--overwrite": true,
	"--verbose":   true,
	"--help":      true,
	"--version":   true,
	"-tui":        true,
	"-try-text":   true,
	"-overwrite":  true,
	"-verbose":    true,
	"-help":       true,
	"-version":    true,
}

// reorderArgs moves all flags to the beginning of the argument list
// This allows flags to be specified after positional arguments
func reorderArgs(args []string) []string {
	var flags []string
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			// Check if this flag takes a value (has = or next arg is not a flag)
			if strings.Contains(arg, "=") {
				// Flag with value like --input=file.log.gz
				continue
			} else if !boolFlags[arg] && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// This flag takes a value, include it
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, arg)
		}
	}

	// Return flags first, then positional arguments
	return append(flags, positional...)
}

func main() {
	// Define flags
	var (
		inputPath  = flag.String("input", "", "Input .log.gz file (or - for stdin)")
		outputPath = flag.String("output", "", "Output file path (default: <input>.payload.json/txt)")
		useTUI     = flag.Bool("tui", false, "Launch interactive TUI mode")
		tryText    = flag.Bool("try-text", false, "Attempt text decoding even if content-type is not supported")
		overwrite  = flag.Bool("overwrite", false, "Overwrite output files if they exist")
		verbose    = flag.Bool("verbose", false, "Enable verbose output")
		showHelp   = flag.Bool("help", false, "Show help message")
		showVer    = flag.Bool("version", false, "Show version")
	)

	// Reorder arguments to handle flags after positional arguments
	// This allows both "cf-dlp-decode --overwrite file.log.gz" and "cf-dlp-decode file.log.gz --overwrite"
	reorderedArgs := reorderArgs(os.Args[1:])
	if err := flag.CommandLine.Parse(reorderedArgs); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to parse flags: %v\n", err)
		os.Exit(2)
	}

	// Handle help
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Handle version
	if *showVer {
		fmt.Printf("Cloudflare DLP Forensic Copy Decoder v%s\n", version)
		os.Exit(0)
	}

	// Get input from positional argument if not specified via flag
	if *inputPath == "" && flag.NArg() > 0 {
		*inputPath = flag.Arg(0)
	}

	// If the user ran the binary without any arguments, default to TUI mode.
	// This keeps the non-interactive CLI behavior when a filename or --input
	// is provided, but makes the TUI the default for zero-argument runs.
	if *inputPath == "" && flag.NArg() == 0 && !*useTUI {
		*useTUI = true
	}

	// Validate input (only for non-TUI runs)
	if *inputPath == "" && !*useTUI {
		fmt.Fprintln(os.Stderr, "ERROR: No input file specified")
		printHelp()
		os.Exit(2)
	}

	// Check input file extension (unless reading from stdin) -- only for non-TUI runs
	if !*useTUI {
		if *inputPath != "-" && !strings.HasSuffix(*inputPath, ".log.gz") && !strings.HasSuffix(*inputPath, ".json") {
			fmt.Fprintf(os.Stderr, "ERROR: Input must end with .log.gz or .json (got: %s)\n", *inputPath)
			os.Exit(1)
		}
	}

	// Launch TUI mode if requested
	if *useTUI {
		// If in print mode (used by tests), print the resolved mode instead of
		// launching the interactive TUI. This avoids blocking tests.
		if os.Getenv("CF_DLP_DECODE_PRINT_MODE") == "1" {
			fmt.Println("TUI")
			return
		}
		if err := runTUI(*inputPath); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run non-interactive CLI mode
	if err := runCLI(*inputPath, *outputPath, *tryText, *overwrite, *verbose); err != nil {
		if *verbose {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		} else {
			// Machine-friendly error message
			fmt.Fprintln(os.Stderr, err.Error())
		}
		os.Exit(1)
	}
}

func runTUI(inputPath string) error {
	model := ui.NewModel(inputPath)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func runCLI(inputPath, outputPath string, tryText, overwrite, verbose bool) error {
	// Read input
	reader, err := utils.ReadInput(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Decode (supports single and multi-payload files)
	opts := decoder.DecodeOptions{
		TryText: tryText,
		Verbose: verbose,
	}

	results, err := decoder.DecodeLogFileMulti(reader, opts)
	if err != nil {
		return err
	}

	if len(results) == 1 {
		// Single payload: use the original (un-numbered) output filenames.
		result := results[0]

		logJSONPath := decoder.GetLogJSONFilename(inputPath)
		payloadPath := outputPath
		if payloadPath == "" {
			payloadPath = decoder.GetOutputFilename(inputPath, result.IsJSON)
		}

		if err := utils.WriteFile(logJSONPath, []byte(result.LogJSON), overwrite); err != nil {
			return fmt.Errorf("failed to write log JSON: %w", err)
		}
		if err := utils.WriteFile(payloadPath, result.Payload, overwrite); err != nil {
			return fmt.Errorf("failed to write payload: %w", err)
		}

		fmt.Println("Wrote:")
		fmt.Printf("  %s\n", logJSONPath)
		fmt.Printf("  %s\n", payloadPath)
		return nil
	}

	// Multiple payloads: use numbered output filenames (1-based).
	// The --output flag is ignored when there are multiple payloads.
	if outputPath != "" {
		fmt.Fprintf(os.Stderr, "WARNING: --output is ignored when the input contains multiple payloads\n")
	}

	fmt.Printf("Wrote (%d payloads):\n", len(results))
	for i, result := range results {
		n := i + 1
		logJSONPath := decoder.GetLogJSONFilenameN(inputPath, n)
		payloadPath := decoder.GetOutputFilenameN(inputPath, result.IsJSON, n)

		if err := utils.WriteFile(logJSONPath, []byte(result.LogJSON), overwrite); err != nil {
			return fmt.Errorf("payload %d: failed to write log JSON: %w", n, err)
		}
		if err := utils.WriteFile(payloadPath, result.Payload, overwrite); err != nil {
			return fmt.Errorf("payload %d: failed to write payload: %w", n, err)
		}

		fmt.Printf("  [%d] %s\n", n, logJSONPath)
		fmt.Printf("  [%d] %s\n", n, payloadPath)
	}

	return nil
}

func printHelp() {
	fmt.Print(`Cloudflare DLP Forensic Copy Decoder

Usage:
  cf-dlp-decode [OPTIONS] <input.log.gz>
  cf-dlp-decode [OPTIONS] <input.json>
  cf-dlp-decode --tui <input.log.gz>

What it does:
  1) Reads <input.log.gz> (gzip-compressed) or <input.json> (plain JSON) -> <input.log.json>
  2) Pretty-prints <input.log.json> (jq equivalent)
  3) Decodes .Payload based on:
       - .Headers.content-type starts with "application/json"    => base64 -> JSON
       - .Headers.content-encoding equals "gzip"                 => base64 -> gzip -> JSON (or other format)
       - .Headers.content-type starts with "text/plain"          => base64 -> plain text
       - .Headers.content-type starts with "multipart/form-data" => base64 -> plain text
  4) Optionally: with --try-text, attempts text decode even when content-type is not supported

Outputs:
  <input.log.json>     (pretty-printed log)
  <input.payload.json> (decoded payload for JSON content)
  <input.payload.txt>  (decoded payload for text content)

Options:
  --input PATH       Input .log.gz or .json file (or - for stdin)
  --output PATH      Output file path (default: auto-generated based on input)
  --tui              Launch interactive TUI mode
  --try-text         Attempt text decoding even if content-type is not supported
  --overwrite        Overwrite output files if they exist
  --verbose          Enable verbose output
  --help             Show this help message
  --version          Show version information

Examples:
  # Basic usage (non-interactive) with gzip input
  cf-dlp-decode input.log.gz

  # Basic usage with plain JSON input
  cf-dlp-decode input.json

  # Interactive TUI mode
  cf-dlp-decode --tui input.log.gz

  # Force text decoding
  cf-dlp-decode --try-text input.log.gz

  # Custom output path
  cf-dlp-decode --input input.log.gz --output custom-output.json

  # Read from stdin
  cat input.log.gz | cf-dlp-decode --input -

For more information, see: https://github.com/Devolvio-B-V/cloudflare-dlp-forensic-copy-decoder
`)
}
