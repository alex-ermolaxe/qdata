package engine

import (
	"strings"
	"testing"
)

func TestInfoCommandListsRegisteredCommands(t *testing.T) {
	e := &Engine{}
	result, err := e.ExecuteCommand("INFO")
	if err != nil {
		t.Fatalf("INFO error = %v", err)
	}

	for _, name := range []string{
		"WHERE", "SELECT", "EXCLUDE", "SORT", "SHOW", "SAVE", "FLATTEN",
		"TRANSFORM", "RESET", "COUNT", "SCHEMA", "INFO", "EXIT",
	} {
		if !strings.Contains(result.Output, name) {
			t.Errorf("INFO output does not include %s", name)
		}
	}
	if strings.Contains(result.Output, "VERSION") {
		t.Error("INFO lists VERSION, which is a flag rather than a command")
	}
}

func TestExitCommandReturnsExitSignal(t *testing.T) {
	result, err := (&Engine{}).ExecuteCommand("EXIT")
	if err != nil {
		t.Fatalf("EXIT error = %v", err)
	}
	if !result.Exit || result.Output != "Bye!\n" {
		t.Errorf("EXIT result = %#v, want exit signal and Bye output", result)
	}
	if _, err := (&Engine{}).ExecuteCommand("EXIT now"); err == nil {
		t.Error("EXIT with arguments did not return an error")
	}
}
