package engine

import (
	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

// Session stores the state of the current session
type Session struct {
	// FilePath - path to the source file
	FilePath string

	// Format - determined file format
	Format format.Format

	// Original - source data (never changes)
	Original []format.Record

	// Current - current intermediate result
	Current []format.Record

	// Schema - field schema for autocomplete and validation
	Schema *schema.Schema
}

// NewSession creates a new session
func NewSession(filePath string, f format.Format, records []format.Record) *Session {
	s := &Session{
		FilePath: filePath,
		Format:   f,
		Original: records,
		Current:  make([]format.Record, len(records)),
		Schema:   schema.Infer(records, 100),
	}

	// Copy records so Original and Current don't reference the same data
	copy(s.Current, records)

	return s
}

// Reset resets Current to Original
func (s *Session) Reset() {
	s.Current = make([]format.Record, len(s.Original))
	copy(s.Current, s.Original)
	s.UpdateSchema()
}

// UpdateSchema rebuilds the schema based on current records
func (s *Session) UpdateSchema() {
	s.Schema = schema.Infer(s.Current, 100)
}

// TotalRecords returns the number of records in the current result
func (s *Session) TotalRecords() int {
	return len(s.Current)
}

// OriginalRecords returns the number of records in the source data
func (s *Session) OriginalRecords() int {
	return len(s.Original)
}
