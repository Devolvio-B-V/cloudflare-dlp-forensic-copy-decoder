package main

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

// helper runs `go run` for the CLI with given args and returns stdout+stderr and error
func runGoRunCLI(t *testing.T, args ...string) (string, error) {
    t.Helper()
    // Prepend the package path for go run
    cmdArgs := append([]string{"run", "./cmd/cf-dlp-decode"}, args...)
    cmd := exec.Command("go", cmdArgs...)
    // Ensure we run from the module root so the relative package path resolves.
    cwd, err := os.Getwd()
    if err == nil {
        // tests run in cmd/cf-dlp-decode, module root is two levels up
        cmd.Dir = filepath.Join(cwd, "..", "..")
    }
    // Enable print mode so the program doesn't start the interactive TUI
    cmd.Env = append(os.Environ(), "CF_DLP_DECODE_PRINT_MODE=1")
    out, err := cmd.CombinedOutput()
    return string(out), err
}

func TestDefaultToTUI_NoArgs(t *testing.T) {
    out, err := runGoRunCLI(t)
    if err != nil {
        t.Fatalf("go run failed: %v\noutput: %s", err, out)
    }
    if strings.TrimSpace(out) != "TUI" {
        t.Fatalf("expected TUI mode by default, got: %q", out)
    }
}

func TestGivenFilename_UsesCLI(t *testing.T) {
    out, err := runGoRunCLI(t, "--overwrite", "testdata/input.log.gz")
    if err != nil {
        t.Fatalf("go run failed: %v\noutput: %s", err, out)
    }
    // When a filename is provided the binary should not default to TUI; print nothing
    // and exit normally. Our print-mode prints only when TUI is selected, so expect empty output.
    if strings.TrimSpace(out) == "TUI" {
        t.Fatalf("expected non-TUI mode when filename provided, got: %q", out)
    }
}

func TestExplicitTUIFlag(t *testing.T) {
    out, err := runGoRunCLI(t, "--tui")
    if err != nil {
        t.Fatalf("go run failed: %v\noutput: %s", err, out)
    }
    if strings.TrimSpace(out) != "TUI" {
        t.Fatalf("expected TUI when --tui set, got: %q", out)
    }
}
