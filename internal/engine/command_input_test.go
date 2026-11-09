package engine

import "testing"

func TestParseCommandInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantCommand string
		wantArgs    string
		wantOK      bool
	}{
		{
			name:   "empty input",
			input:  " \t\n",
			wantOK: false,
		},
		{
			name:        "command without arguments",
			input:       " count ",
			wantCommand: "COUNT",
			wantOK:      true,
		},
		{
			name:        "leading whitespace and mixed case command",
			input:       " \twhere\t status = \"active\" \t",
			wantCommand: "WHERE",
			wantArgs:    `status = "active"`,
			wantOK:      true,
		},
		{
			name:        "preserve spaces inside arguments",
			input:       `WHERE  name = "Alice Johnson"  `,
			wantCommand: "WHERE",
			wantArgs:    `name = "Alice Johnson"`,
			wantOK:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseCommandInput(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("parseCommandInput() ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}

			if got.name != tt.wantCommand {
				t.Errorf("command = %q, want %q", got.name, tt.wantCommand)
			}
			if got.args != tt.wantArgs {
				t.Errorf("arguments = %q, want %q", got.args, tt.wantArgs)
			}
		})
	}
}
