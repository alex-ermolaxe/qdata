package engine

import (
	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

// Session хранит состояние текущей сессии
type Session struct {
	// FilePath — путь к исходному файлу
	FilePath string

	// Format — определённый формат файла
	Format format.Format

	// Original — исходные данные (никогда не меняются)
	Original []format.Record

	// Current — текущий промежуточный результат
	Current []format.Record

	// Schema — схема полей для автодополнения и валидации
	Schema *schema.Schema
}

// NewSession создаёт новую сессию
func NewSession(filePath string, f format.Format, records []format.Record) *Session {
	s := &Session{
		FilePath: filePath,
		Format:   f,
		Original: records,
		Current:  make([]format.Record, len(records)),
		Schema:   schema.Infer(records, 100),
	}

	// Копируем записи чтобы Original и Current не ссылались на одни данные
	copy(s.Current, records)

	return s
}

// Reset сбрасывает Current к Original
func (s *Session) Reset() {
	s.Current = make([]format.Record, len(s.Original))
	copy(s.Current, s.Original)
}

// TotalRecords возвращает количество записей в текущем результате
func (s *Session) TotalRecords() int {
	return len(s.Current)
}

// OriginalRecords возвращает количество записей в исходных данных
func (s *Session) OriginalRecords() int {
	return len(s.Original)
}
