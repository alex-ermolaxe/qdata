package format

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// Record — базовый тип для одной записи
type Record = map[string]any

// Format — интерфейс для работы с конкретным форматом файла.
// Чтобы добавить новый формат — достаточно реализовать этот интерфейс.
type Format interface {
	// Decode читает данные из reader и возвращает массив записей
	Decode(r io.Reader) ([]Record, error)

	// Encode сериализует массив записей в writer
	Encode(w io.Writer, records []Record) error

	// Extensions возвращает список расширений файлов для автоопределения
	Extensions() []string
}

// registry — реестр зарегистрированных форматов
var registry = map[string]Format{}

// Register регистрирует новый формат по имени
func Register(name string, f Format) {
	registry[name] = f
}

// Detect определяет формат по расширению файла
func Detect(filename string) (Format, error) {
	ext := strings.TrimPrefix(filepath.Ext(filename), ".")
	for _, f := range registry {
		for _, e := range f.Extensions() {
			if e == ext {
				return f, nil
			}
		}
	}
	return nil, fmt.Errorf("unsupported format: .%s", ext)
}

// Get возвращает формат по имени
func Get(name string) (Format, error) {
	f, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown format: %s", name)
	}
	return f, nil
}
