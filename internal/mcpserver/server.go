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
		mcp.WithDestructiveHintAnnotation(false),
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

	for _, spec := range engine.CommandSpecs() {
		if spec.MCP != nil {
			srv.addCommandTool(mcpSrv, spec)
		}
	}
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

func (srv *mcpServer) addCommandTool(mcpSrv *server.MCPServer, spec engine.CommandSpec) {
	mcpSpec := spec.MCP
	options := []mcp.ToolOption{
		mcp.WithDescription(mcpSpec.Description),
		mcp.WithString("session_id", mcp.Required(), mcp.Description("ID returned by load for the session to operate on.")),
	}
	for _, argument := range mcpSpec.Arguments {
		propertyOptions := []mcp.PropertyOption{mcp.Description(argument.Description)}
		if argument.Required {
			propertyOptions = append(propertyOptions, mcp.Required())
		}
		options = append(options, mcp.WithString(argument.Name, propertyOptions...))
	}
	if mcpSpec.DestructiveHint != nil {
		options = append(options, mcp.WithDestructiveHintAnnotation(*mcpSpec.DestructiveHint))
	}

	mcpSrv.AddTool(mcp.NewTool(mcpSpec.Name, options...), func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		line := spec.Name
		for _, argument := range mcpSpec.Arguments {
			value, err := request.RequireString(argument.Name)
			if err != nil {
				if argument.Required {
					return errorResult(err), nil
				}
				value = request.GetString(argument.Name, "")
			}
			if strings.TrimSpace(value) != "" {
				line += " " + value
			}
		}
		result, err := eng.ExecuteCommand(line)
		if err != nil {
			return errorResult(err), nil
		}
		return textResult(result.Output), nil
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
