package engine

import "testing"

func TestCommandSpecsHaveUniqueNamesAndHandlers(t *testing.T) {
	seen := make(map[string]struct{})
	for _, spec := range CommandSpecs() {
		if spec.Name == "" {
			t.Error("command spec has an empty name")
		}
		if _, exists := seen[spec.Name]; exists {
			t.Errorf("duplicate command spec %q", spec.Name)
		}
		seen[spec.Name] = struct{}{}
		if spec.Run == nil {
			t.Errorf("command %q has no handler", spec.Name)
		}
	}
}
