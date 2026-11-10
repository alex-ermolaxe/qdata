package engine

import "github.com/alex-ermolaxe/qdata/internal/completer"

// CommandSpec describes a command and connects its interfaces to its handler.
type CommandSpec struct {
	Name        string
	Usage       string
	Description string
	Run         CommandHandler
	Completion  completer.CompletionFunc
	MCP         *MCPToolSpec
}

// CommandHandler executes a command with its unparsed argument string.
type CommandHandler func(*Engine, string) (CommandResult, error)

// CommandResult contains command output and an optional REPL control signal.
type CommandResult struct {
	Output string
	Exit   bool
}

// MCPToolSpec describes an MCP representation without depending on an MCP SDK.
type MCPToolSpec struct {
	Name               string
	Description        string
	Arguments          []MCPArgumentSpec
	DestructiveHint    *bool
	SessionIndependent bool
}

// MCPArgumentSpec describes one argument exposed by an MCP tool.
type MCPArgumentSpec struct {
	Name        string
	Description string
	Required    bool
}

func boolPointer(value bool) *bool {
	return &value
}
