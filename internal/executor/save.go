package executor

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/format"
)

// Save saves records next to the original and returns the status text.
// If confirm is nil, an existing target returns an error instead of prompting.
func Save(records []format.Record, originalPath string, fileName string, f format.Format, confirm func(string) bool) (string, error) {
	dir := filepath.Dir(originalPath)
	ext := strings.TrimPrefix(filepath.Ext(originalPath), ".")

	// Form file name
	if fileName == "" {
		baseName := strings.TrimSuffix(filepath.Base(originalPath), filepath.Ext(originalPath))
		fileName = baseName + "_result"
	}

	targetPath := filepath.Join(dir, fileName+"."+ext)

	// If file already exists - ask for confirmation
	if _, err := os.Stat(targetPath); err == nil {
		if confirm == nil {
			return "", fmt.Errorf("file exists, use AS <name> or overwrite")
		}
		if !confirm(targetPath) {
			return "Save cancelled.\n", nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to inspect target file: %w", err)
	}

	// Create file
	file, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Serialize data
	if err := f.Encode(file, records); err != nil {
		return "", fmt.Errorf("failed to encode records: %w", err)
	}

	return fmt.Sprintf("✓ Saved %d records → %s\n", len(records), targetPath), nil
}

// ConfirmOverwrite asks for confirmation to overwrite a file in interactive use.
func ConfirmOverwrite(path string) bool {
	fmt.Printf("File already exists: %s\nOverwrite? [y/N]: ", path)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
