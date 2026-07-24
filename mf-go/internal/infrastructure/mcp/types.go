package mcp

// MCPPayload is the standard MCP request envelope.
type MCPPayload struct {
	RequestType string         `json:"request_type"`
	Context     map[string]any `json:"context"`
	Query       string         `json:"query"`
}

// RichResult is the enriched MCP response envelope.
type RichResult struct {
	Data     map[string]any `json:"data"`
	Metadata map[string]any `json:"metadata"`
	Source   string         `json:"source"`
}
