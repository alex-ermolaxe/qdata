package executor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/format"
)

// Save сохраняет записи в файл рядом с оригинальным
func Save(records []format.Record, originalPath string, fileName string, f format.Format) error {
	dir := filepath.Dir(originalPath)
	ext := strings.TrimPrefix(filepath.Ext(originalPath), ".")

	// Формируем имя файла
	if fileName == "" {
		baseName := strings.TrimSuffix(filepath.Base(originalPath), filepath.Ext(originalPath))
		fileName = baseName + "_result"
	}

	targetPath := filepath.Join(dir, fileName+"."+ext)

	// Если файл уже существует — запрашиваем подтверждение
	if _, err := os.Stat(targetPath); err == nil {
		if !confirmOverwrite(targetPath) {
			fmt.Println("Save cancelled.")
			return nil
		}
	}

	// Создаём файл
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Сериализуем данные
	if err := f.Encode(file, records); err != nil {
		return fmt.Errorf("failed to encode records: %w", err)
	}

	fmt.Printf("✓ Saved %d records → %s\n", len(records), targetPath)

	return nil
}

// confirmOverwrite запрашивает подтверждение перезаписи файла
func confirmOverwrite(path string) bool {
	fmt.Printf("File already exists: %s\nOverwrite? [y/N]: ", path)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
