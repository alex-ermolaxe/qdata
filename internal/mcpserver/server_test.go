package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/engine"
	formatjson "github.com/alex-ermolaxe/qdata/internal/format/json"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestSessionsAreIndependent(t *testing.T) {
	formatjson.Register()
	dir := t.TempDir()
	usersPath := filepath.Join(dir, "users.json")
	ordersPath := filepath.Join(dir, "orders.json")
	if err := os.WriteFile(usersPath, []byte(`[{"status":"active"},{"status":"inactive"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ordersPath, []byte(`[{"status":"pending"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "users_result.json"), []byte("[]"), 0o600); err != nil {
		t.Fatal(err)
	}

	srv := &mcpServer{sessions: make(map[string]*engine.Engine)}
	mcpSrv := server.NewMCPServer("test", "test")
	srv.registerTools(mcpSrv)
	tools := mcpSrv.ListTools()
	for _, name := range []string{
		"load", "list_sessions", "close_session", "where", "select", "exclude",
		"sort", "flatten", "transform", "show", "save", "reset", "count", "schema",
	} {
		if _, exists := tools[name]; !exists {
			t.Errorf("MCP tool %q was not registered", name)
		}
	}
	for _, name := range []string{"exit", "info", "version"} {
		if _, exists := tools[name]; exists {
			t.Errorf("MCP tool %q must not be registered", name)
		}
	}
	loadAnnotations := tools["load"].Tool.Annotations
	if loadAnnotations.DestructiveHint == nil || *loadAnnotations.DestructiveHint {
		t.Errorf("load destructiveHint = %v, want false", loadAnnotations.DestructiveHint)
	}
	saveAnnotations := tools["save"].Tool.Annotations
	if saveAnnotations.DestructiveHint == nil || !*saveAnnotations.DestructiveHint {
		t.Errorf("save destructiveHint = %v, want true", saveAnnotations.DestructiveHint)
	}
	usersResult, err := srv.load(context.Background(), toolRequest(map[string]any{"path": usersPath}))
	if err != nil {
		t.Fatal(err)
	}
	if usersResult.IsError {
		t.Fatalf("load users returned error: %s", resultText(t, usersResult))
	}
	ordersResult, err := srv.load(context.Background(), toolRequest(map[string]any{"path": ordersPath}))
	if err != nil {
		t.Fatal(err)
	}
	if ordersResult.IsError {
		t.Fatalf("load orders returned error: %s", resultText(t, ordersResult))
	}

	where := mcpSrv.GetTool("where").Handler
	whereResult, err := where(context.Background(), toolRequest(map[string]any{
		"session_id": "users",
		"condition":  `status = "active"`,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if whereResult.IsError {
		t.Fatalf("where returned error: %s", resultText(t, whereResult))
	}
	if got := srv.sessions["users"].RecordCount(); got != 1 {
		t.Errorf("users record count = %d, want 1", got)
	}
	if got := srv.sessions["orders"].RecordCount(); got != 1 {
		t.Errorf("orders record count = %d, want 1", got)
	}
	count := mcpSrv.GetTool("count").Handler
	ordersCount, err := count(context.Background(), toolRequest(map[string]any{"session_id": "orders"}))
	if err != nil {
		t.Fatal(err)
	}
	if got := resultText(t, ordersCount); !strings.Contains(got, "1 records") {
		t.Errorf("orders count result = %q, want count 1", got)
	}
	save := mcpSrv.GetTool("save").Handler
	saveResult, err := save(context.Background(), toolRequest(map[string]any{
		"session_id": "users",
		"options":    "",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !saveResult.IsError || !strings.Contains(resultText(t, saveResult), "file exists, use AS <name> or overwrite") {
		t.Errorf("save existing target result = %#v, want non-interactive ERROR", saveResult)
	}

	listResult, err := srv.listSessions(context.Background(), toolRequest(nil))
	if err != nil {
		t.Fatal(err)
	}
	list := resultText(t, listResult)
	if !strings.Contains(list, `"session_id": "users"`) || !strings.Contains(list, `"session_id": "orders"`) {
		t.Errorf("list_sessions result missing sessions: %s", list)
	}

	unknownResult, err := srv.closeSession(context.Background(), toolRequest(map[string]any{"session_id": "missing"}))
	if err != nil {
		t.Fatal(err)
	}
	if !unknownResult.IsError || !strings.HasPrefix(resultText(t, unknownResult), "ERROR: ") {
		t.Errorf("unknown session result = %#v, want ERROR result", unknownResult)
	}
	if got := len(srv.sessions); got != 2 {
		t.Errorf("unknown session error changed session count to %d, want 2", got)
	}
	unknownWhere, err := where(context.Background(), toolRequest(map[string]any{
		"session_id": "missing",
		"condition":  `status = "active"`,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !unknownWhere.IsError || !strings.HasPrefix(resultText(t, unknownWhere), "ERROR: ") {
		t.Errorf("unknown where session result = %#v, want ERROR result", unknownWhere)
	}
	ordersCount, err = count(context.Background(), toolRequest(map[string]any{"session_id": "orders"}))
	if err != nil || ordersCount.IsError || !strings.Contains(resultText(t, ordersCount), "1 records") {
		t.Errorf("server did not remain usable after unknown session: result=%v error=%v", ordersCount, err)
	}

	duplicateResult, err := srv.load(context.Background(), toolRequest(map[string]any{
		"path":       usersPath,
		"session_id": "users",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if duplicateResult.IsError || !strings.Contains(resultText(t, duplicateResult), `"session_id": "users_2"`) {
		t.Errorf("duplicate session id response = %q", resultText(t, duplicateResult))
	}
}

func toolRequest(arguments map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: arguments},
	}
}

func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		return ""
	}
	content, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("result content has type %T, want text", result.Content[0])
	}
	return content.Text
}
