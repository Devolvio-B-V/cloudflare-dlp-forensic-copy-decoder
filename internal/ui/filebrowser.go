package ui

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileEntry represents a file or directory in the browser
type FileEntry struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
}

// FileBrowser manages file navigation
type FileBrowser struct {
	currentDir    string
	entries       []FileEntry
	selectedIndex int
	viewOffset    int
}

// NewFileBrowser creates a new file browser starting at the given directory
func NewFileBrowser(startDir string) (*FileBrowser, error) {
	if startDir == "" {
		var err error
		startDir, err = os.Getwd()
		if err != nil {
			startDir = "."
		}
	}

	fb := &FileBrowser{
		currentDir: startDir,
	}

	if err := fb.loadEntries(); err != nil {
		return nil, err
	}

	return fb, nil
}

// loadEntries reads the current directory's contents
func (fb *FileBrowser) loadEntries() error {
	entries := []FileEntry{}

	// Add parent directory entry if not at root
	if fb.currentDir != "/" && fb.currentDir != "." {
		entries = append(entries, FileEntry{
			Name:  "..",
			Path:  filepath.Dir(fb.currentDir),
			IsDir: true,
		})
	}

	// Read directory contents
	dirEntries, err := os.ReadDir(fb.currentDir)
	if err != nil {
		return err
	}

	// Convert to FileEntry
	for _, entry := range dirEntries {
		// Skip hidden files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(fb.currentDir, entry.Name())

		entries = append(entries, FileEntry{
			Name:  entry.Name(),
			Path:  fullPath,
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		})
	}

	// Sort: directories first, then files, alphabetically
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	fb.entries = entries
	fb.selectedIndex = 0
	fb.viewOffset = 0

	return nil
}

// MoveUp moves the selection up
func (fb *FileBrowser) MoveUp() {
	if fb.selectedIndex > 0 {
		fb.selectedIndex--
		// Adjust view offset if selection moves out of view
		if fb.selectedIndex < fb.viewOffset {
			fb.viewOffset = fb.selectedIndex
		}
	}
}

// MoveDown moves the selection down
func (fb *FileBrowser) MoveDown() {
	if fb.selectedIndex < len(fb.entries)-1 {
		fb.selectedIndex++
	}
}

// Enter navigates into a directory or selects a file
func (fb *FileBrowser) Enter() (string, bool, error) {
	if len(fb.entries) == 0 {
		return "", false, nil
	}

	selected := fb.entries[fb.selectedIndex]

	if selected.IsDir {
		fb.currentDir = selected.Path
		if err := fb.loadEntries(); err != nil {
			return "", false, err
		}
		return "", false, nil
	}

	// File selected - return it if it's a supported forensic copy file
	if strings.HasSuffix(selected.Name, ".log.gz") || strings.HasSuffix(selected.Name, ".json") {
		return selected.Path, true, nil
	}

	return "", false, nil
}

// GetSelectedEntry returns the currently selected entry
func (fb *FileBrowser) GetSelectedEntry() *FileEntry {
	if len(fb.entries) == 0 {
		return nil
	}
	return &fb.entries[fb.selectedIndex]
}

// GetEntries returns all entries
func (fb *FileBrowser) GetEntries() []FileEntry {
	return fb.entries
}

// GetCurrentDir returns the current directory path
func (fb *FileBrowser) GetCurrentDir() string {
	return fb.currentDir
}

// GetSelectedIndex returns the current selection index
func (fb *FileBrowser) GetSelectedIndex() int {
	return fb.selectedIndex
}

// GetViewOffset returns the view offset for scrolling
func (fb *FileBrowser) GetViewOffset() int {
	return fb.viewOffset
}

// SetViewOffset sets the view offset for scrolling
func (fb *FileBrowser) SetViewOffset(offset int) {
	fb.viewOffset = offset
}

// GoToHomeDir navigates to the user's home directory
func (fb *FileBrowser) GoToHomeDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	fb.currentDir = homeDir
	return fb.loadEntries()
}

// GoToDir navigates to a specific directory
func (fb *FileBrowser) GoToDir(dir string) error {
	// Resolve to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	// Check if directory exists
	info, err := os.Stat(absDir)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fs.ErrNotExist
	}

	fb.currentDir = absDir
	return fb.loadEntries()
}

// GotoTop moves selection to the first entry
func (fb *FileBrowser) GotoTop() {
	fb.selectedIndex = 0
	fb.viewOffset = 0
}

// GotoBottom moves selection to the last entry
func (fb *FileBrowser) GotoBottom() {
	if len(fb.entries) > 0 {
		fb.selectedIndex = len(fb.entries) - 1
	}
}

// AdjustViewOffset adjusts the view offset to keep the selected item visible
func (fb *FileBrowser) AdjustViewOffset(maxVisible int) {
	selectedIdx := fb.selectedIndex
	startIdx := fb.viewOffset
	endIdx := startIdx + maxVisible

	if endIdx > len(fb.entries) {
		endIdx = len(fb.entries)
	}

	// Adjust view offset if selected is out of view
	if selectedIdx < startIdx {
		fb.viewOffset = selectedIdx
	} else if selectedIdx >= endIdx {
		fb.viewOffset = selectedIdx - maxVisible + 1
		if fb.viewOffset < 0 {
			fb.viewOffset = 0
		}
	}
}
