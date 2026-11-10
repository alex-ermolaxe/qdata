package engine

import (
	"strings"
	"testing"
)

func TestHelpCommandListsRegisteredCommands(t *testing.T) {
	e := &Engine{}
	result, err := e.ExecuteCommand("HELP")
	if err != nil {
		t.Fatalf("HELP error = %v", err)
	}

	for _, name := range []string{
		"WHERE", "SELECT", "EXCLUDE", "SORT", "SHOW", "SAVE", "FLATTEN",
		"TRANSFORM", "RESET", "COUNT", "SCHEMA", "HELP", "EXIT",
	} {
		if !strings.Contains(result.Output, name) {
			t.Errorf("HELP output does not include %s", name)
		}
	}
	if strings.Contains(result.Output, "VERSION") {
		t.Error("HELP lists VERSION, which is a flag rather than a command")
	}
	if _, err := e.ExecuteCommand("INFO"); err == nil {
		t.Error("INFO is still registered as a command")
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
