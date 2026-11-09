package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/completer"
	"github.com/alex-ermolaxe/qdata/internal/format"
	formatjson "github.com/alex-ermolaxe/qdata/internal/format/json"
)

func TestResetAndCountCommandSpecs(t *testing.T) {
	first := format.NewRecord()
	first.Set("id", 1)
	second := format.NewRecord()
	second.Set("id", 2)
	e := &Engine{session: NewSession("input.json", nil, []format.Record{first, second})}
	e.session.Current = e.session.Current[:1]

	count, err := e.ExecuteCommand("COUNT")
	if err != nil {
		t.Fatalf("COUNT error = %v", err)
	}
	if count.Output != "1 records\n" {
		t.Errorf("COUNT output = %q, want %q", count.Output, "1 records\n")
	}

	if _, err := e.ExecuteCommand("RESET"); err != nil {
		t.Fatalf("RESET error = %v", err)
	}
	if got := e.RecordCount(); got != 2 {
		t.Errorf("record count after RESET = %d, want 2", got)
	}
}

func TestShowCommandSpec(t *testing.T) {
	first := format.NewRecord()
	first.Set("name", "first")
	second := format.NewRecord()
	second.Set("name", "second")
	e := &Engine{session: NewSession("input.json", nil, []format.Record{first, second})}

	result, err := e.ExecuteCommand("SHOW LIMIT 1 OFFSET 1")
	if err != nil {
		t.Fatalf("SHOW error = %v", err)
	}
	if !strings.Contains(result.Output, "second") || strings.Contains(result.Output, "first") {
		t.Errorf("SHOW output = %q, want only the second record", result.Output)
	}
}

func TestSaveCommandSpec(t *testing.T) {
	formatjson.Register()
	sourcePath := filepath.Join(t.TempDir(), "source.json")
	targetPath := filepath.Join(filepath.Dir(sourcePath), "filtered.json")
	jsonFormat, err := format.Get("json")
	if err != nil {
		t.Fatal(err)
	}

	record := format.NewRecord()
	record.Set("name", "Alice")
	e := &Engine{session: NewSession(sourcePath, jsonFormat, []format.Record{record})}
	result, err := e.ExecuteCommand("SAVE AS filtered FORMAT json")
	if err != nil {
		t.Fatalf("SAVE error = %v", err)
	}
	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("SAVE did not create target file: %v", err)
	}
	if !strings.Contains(result.Output, targetPath) {
		t.Errorf("SAVE output = %q, want target path %q", result.Output, targetPath)
	}
}

func TestShowAndSaveCompletion(t *testing.T) {
	if got := completeShow(completer.CompletionContext{
		HasTrailingSpace: true,
	}); !containsString(got, "LIMIT") || !containsString(got, "OFFSET") {
		t.Errorf("SHOW completions = %v, want LIMIT and OFFSET", got)
	}
	if got := completeShow(completer.CompletionContext{
		Arguments:        []string{"LIMIT"},
		HasTrailingSpace: true,
	}); len(got) != 0 {
		t.Errorf("SHOW completions after LIMIT = %v, want none", got)
	}
	if got := completeSave(completer.CompletionContext{
		Arguments:        []string{"FORMAT"},
		HasTrailingSpace: true,
	}); !containsString(got, "json") {
		t.Errorf("SAVE completions = %v, want available formats", got)
	}
}
