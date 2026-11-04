package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alex-ermolaxe/qdata/internal/engine"
	"github.com/alex-ermolaxe/qdata/internal/version"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mcpServer struct {
	mu       sync.Mutex
	sessions map[string]*engine.Engine
	counter  int
}

type sessionInfo struct {
	SessionID   string `json:"session_id"`
	FilePath    string `json:"file_path"`
	RecordCount int    `json:"record_count"`
}

// Serve starts qdata's MCP server on stdio.
func Serve() error {
	srv := &mcpServer{sessions: make(map[string]*engine.Engine)}
	mcpSrv := server.NewMCPServer("qdata", version.Version)
	srv.registerTools(mcpSrv)
	return server.ServeStdio(mcpSrv)
}

func (srv *mcpServer) registerTools(mcpSrv *server.MCPServer) {
	mcpSrv.AddTool(mcp.NewTool("load",
		mcp.WithDescription("Load a JSON, CSV, or supported data file into a named qdata session. Use the returned session_id in every subsequent tool call. If session_id is omitted, it is generated from the file name; collisions receive a numeric suffix."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the data file to load.")),
		mcp.WithString("format", mcp.Description("Optional format name such as json or csv. If omitted, qdata detects it from the file extension.")),
		mcp.WithString("session_id", mcp.Description("Optional name for this session. If omitted, use the file name without its extension.")),
	), srv.load)

	mcpSrv.AddTool(mcp.NewTool("list_sessions",
		mcp.WithDescription("List all loaded qdata sessions, including each session's id, source file path, and current record count. Call load if there are no sessions."),
	), srv.listSessions)

	mcpSrv.AddTool(mcp.NewTool("close_session",
		mcp.WithDescription("Close and remove a loaded qdata session. This discards its current query state."),
		mcp.WithString("session_id", mcp.Required(), mcp.Description("ID of the session to close.")),
	), srv.closeSession)

	srv.addCommandTool(mcpSrv, "where",
		"Filter records in a session. Syntax: <field> <operator> <value>, optionally combined with AND or OR. Operators: =, !=, >, <, >=, <=, ~, !~, ^, $, IN, EXISTS. Nested fields use dot paths (for example address.city). Example: status = \"active\" AND age >= 18.",
		"condition", "Filter condition in qdata syntax; nested field paths use dots.",
		"WHERE")
	srv.addCommandTool(mcpSrv, "select",
		"Keep only the specified comma-separated fields in each record. Supports nested dot paths and fields inside arrays of objects, such as items.name. Use * to retain all fields.",
		"fields", "Comma-separated field paths, for example id, name, items.name.",
		"SELECT")
	srv.addCommandTool(mcpSrv, "exclude",
		"Remove the specified comma-separated fields from each record. Nested fields can be addressed with dot paths.",
		"fields", "Comma-separated field paths to remove.",
		"EXCLUDE")
	srv.addCommandTool(mcpSrv, "sort",
		"Sort the current records by a field path. The optional direction is ASC or DESC; it defaults to ASC.",
		"spec", "Field path and optional direction, for example created_at DESC.",
		"SORT")
	srv.addCommandTool(mcpSrv, "flatten",
		"Flatten a nested object field into its parent record. If any nested field collides with an existing top-level field, all fields from the object are prefixed with the object's name.",
		"field", "Name of the object field to flatten.",
		"FLATTEN")
	srv.addCommandTool(mcpSrv, "transform",
		"Transform each record into a one-key object. The KEY path supplies the dynamic property name; VALUE supplies the property value. Paths through arrays of objects extract the value from each element. Example: TO-MAP KEY sap_order_id VALUE assets.name.",
		"spec", "Transform specification, for example TO-MAP KEY sap_order_id VALUE assets.name.",
		"TRANSFORM")
	srv.addCommandTool(mcpSrv, "show",
		"Display the current records as formatted JSON. Options may include LIMIT <n> and OFFSET <n>; provide an empty options string to show all records.",
		"options", "Optional display options such as LIMIT 20 OFFSET 10; use an empty string for all records.",
		"SHOW")
	srv.addCommandTool(mcpSrv, "save",
		"Save the current records beside the source file. Options may include AS <filename> and FORMAT <format>. If the target exists, this tool returns an error instead of prompting; choose a different filename.",
		"options", "Save options such as AS filtered_users FORMAT json. Use an empty string for the default output name and format.",
		"SAVE")

	srv.addSessionCommandTool(mcpSrv, "reset",
		"Restore the session's current records to the original loaded dataset.",
		"RESET")
	srv.addSessionCommandTool(mcpSrv, "count",
		"Return the number of records in the session's current result.",
		"COUNT")
	srv.addSessionCommandTool(mcpSrv, "schema",
		"Show the fields and inferred types available in the session's current result. Use this before constructing field paths and filter conditions.",
		"SCHEMA")
}

func (srv *mcpServer) load(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return errorResult(err), nil
	}

	formatName := request.GetString("format", "")
	requestedID := request.GetString("session_id", "")

	srv.mu.Lock()
	defer srv.mu.Unlock()

	eng, err := engine.NewFromFile(path, formatName)
	if err != nil {
		return errorResult(err), nil
	}
	sessionID := requestedID
	if sessionID == "" {
		sessionID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	sessionID = srv.uniqueSessionID(sessionID)
	srv.sessions[sessionID] = eng

	response := sessionInfo{
		SessionID:   sessionID,
		FilePath:    eng.FilePath(),
		RecordCount: eng.RecordCount(),
	}
	return jsonResult(response)
}

func (srv *mcpServer) listSessions(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if len(srv.sessions) == 0 {
		return textResult("No sessions are loaded. Call load to open a data file."), nil
	}
	sessions := make([]sessionInfo, 0, len(srv.sessions))
	for id, eng := range srv.sessions {
		sessions = append(sessions, sessionInfo{
			SessionID:   id,
			FilePath:    eng.FilePath(),
			RecordCount: eng.RecordCount(),
		})
	}
	return jsonResult(sessions)
}

func (srv *mcpServer) closeSession(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID, err := request.RequireString("session_id")
	if err != nil {
		return errorResult(err), nil
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if _, exists := srv.sessions[sessionID]; !exists {
		return errorResult(fmt.Errorf("unknown session_id %q", sessionID)), nil
	}
	delete(srv.sessions, sessionID)
	return textResult(fmt.Sprintf("Closed session %q.", sessionID)), nil
}

func (srv *mcpServer) addCommandTool(mcpSrv *server.MCPServer, name, description, argument, argumentDescription, keyword string) {
	mcpSrv.AddTool(mcp.NewTool(name,
		mcp.WithDescription(description),
		mcp.WithString("session_id", mcp.Required(), mcp.Description("ID returned by load for the session to operate on.")),
		mcp.WithString(argument, mcp.Required(), mcp.Description(argumentDescription)),
	), func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, err := request.RequireString("session_id")
		if err != nil {
			return errorResult(err), nil
		}
		argumentValue, err := request.RequireString(argument)
		if err != nil {
			return errorResult(err), nil
		}

		srv.mu.Lock()
		defer srv.mu.Unlock()
		eng, exists := srv.sessions[sessionID]
		if !exists {
			return errorResult(fmt.Errorf("unknown session_id %q", sessionID)), nil
		}

		line := keyword
		if strings.TrimSpace(argumentValue) != "" {
			line += " " + argumentValue
		}
		output, err := eng.ExecuteCommand(line)
		if err != nil {
			return errorResult(err), nil
		}
		return textResult(output), nil
	})
}

func (srv *mcpServer) addSessionCommandTool(mcpSrv *server.MCPServer, name, description, keyword string) {
	mcpSrv.AddTool(mcp.NewTool(name,
		mcp.WithDescription(description),
		mcp.WithString("session_id", mcp.Required(), mcp.Description("ID returned by load for the session to operate on.")),
	), func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, err := request.RequireString("session_id")
		if err != nil {
			return errorResult(err), nil
		}
		srv.mu.Lock()
		defer srv.mu.Unlock()
		eng, exists := srv.sessions[sessionID]
		if !exists {
			return errorResult(fmt.Errorf("unknown session_id %q", sessionID)), nil
		}
		output, err := eng.ExecuteCommand(keyword)
		if err != nil {
			return errorResult(err), nil
		}
		return textResult(output), nil
	})
}

func (srv *mcpServer) uniqueSessionID(base string) string {
	if base == "" {
		base = "session"
	}
	if _, exists := srv.sessions[base]; !exists {
		return base
	}
	for suffix := 2; ; suffix++ {
		candidate := fmt.Sprintf("%s_%d", base, suffix)
		if _, exists := srv.sessions[candidate]; !exists {
			return candidate
		}
	}
}

func jsonResult(value any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return errorResult(err), nil
	}
	return textResult(string(data)), nil
}

func textResult(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

func errorResult(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError("ERROR: " + err.Error())
}
