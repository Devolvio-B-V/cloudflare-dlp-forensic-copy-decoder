package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileBrowserLoadsEntries(t *testing.T) {
	tmp := t.TempDir()

	// Create files and dirs
	if err := os.WriteFile(filepath.Join(tmp, "a.log.gz"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmp, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	fb, err := NewFileBrowser(tmp)
	if err != nil {
		t.Fatalf("NewFileBrowser failed: %v", err)
	}

	if fb.GetCurrentDir() == "" {
		t.Fatalf("expected current dir to be set")
	}

	entries := fb.GetEntries()
	if len(entries) < 1 {
		t.Fatalf("expected entries, got none")
	}

	// Ensure our created entries exist
	foundA := false
	foundSub := false
	for _, e := range entries {
		if e.Name == "a.log.gz" {
			foundA = true
		}
		if e.Name == "subdir" {
			foundSub = true
		}
	}
	if !foundA || !foundSub {
		t.Fatalf("expected entries a.log.gz and subdir present, got %v", entries)
	}
}

func TestFileBrowser_GoToDirAndHome(t *testing.T) {
	tmp := t.TempDir()
	nested := filepath.Join(tmp, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}

	fb, err := NewFileBrowser(tmp)
	if err != nil {
		t.Fatalf("NewFileBrowser failed: %v", err)
	}

	if err := fb.GoToDir(nested); err != nil {
		t.Fatalf("GoToDir failed: %v", err)
	}
	if filepath.Base(fb.GetCurrentDir()) != "nested" {
		t.Fatalf("expected current dir nested, got %s", fb.GetCurrentDir())
	}

	// GoToHomeDir should not return error in CI environment
	if err := fb.GoToHomeDir(); err != nil {
		t.Fatalf("GoToHomeDir failed: %v", err)
	}
}

func TestFileBrowser_GotoTopBottomAndAdjustView(t *testing.T) {
	tmp := t.TempDir()

	// Create many files
	for i := 0; i < 20; i++ {
		name := filepath.Join(tmp, fmt.Sprintf("f%02d.txt", i))
		if err := os.WriteFile(name, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	fb, err := NewFileBrowser(tmp)
	if err != nil {
		t.Fatalf("NewFileBrowser failed: %v", err)
	}

	fb.GotoBottom()
	if fb.GetSelectedIndex() != len(fb.GetEntries())-1 {
		t.Fatalf("expected selected index at bottom, got %d", fb.GetSelectedIndex())
	}

	fb.GotoTop()
	if fb.GetSelectedIndex() != 0 {
		t.Fatalf("expected selected index at top, got %d", fb.GetSelectedIndex())
	}

	// Test AdjustViewOffset with small viewport
	fb.selectedIndex = len(fb.GetEntries()) - 1
	fb.AdjustViewOffset(5)
	if fb.GetViewOffset() < 0 {
		t.Fatalf("invalid view offset %d", fb.GetViewOffset())
	}
}
