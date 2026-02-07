package model

import "time"

// Output represents the complete output structure
type Output struct {
	Version     string    `json:"version"`
	GeneratedAt time.Time `json:"generated_at"`
	Sessions    []Session `json:"sessions"`
}

// Session represents a unified agent session
type Session struct {
	AgentType string          `json:"agent_type"` // "claude" or "goose"
	SessionID string          `json:"session_id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Metadata  SessionMetadata `json:"metadata"`
	Messages  []Message       `json:"messages"`
	Subagents []Subagent      `json:"subagents,omitempty"`
}

// SessionMetadata contains session-level metadata
type SessionMetadata struct {
	WorkingDir string            `json:"working_dir,omitempty"`
	GitBranch  string            `json:"git_branch,omitempty"`
	Model      string            `json:"model,omitempty"`
	Provider   string            `json:"provider,omitempty"`
	Name       string            `json:"name,omitempty"`
	Extra      map[string]string `json:"extra,omitempty"`
}

// Message represents a single message in a conversation
type Message struct {
	UUID       string          `json:"uuid"`
	ParentUUID string          `json:"parent_uuid,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
	Role       string          `json:"role"` // "user" or "assistant"
	Content    []ContentBlock  `json:"content"`
	Metadata   MessageMetadata `json:"metadata"`
}

// ContentBlock represents a piece of content (text, tool use, or tool result)
type ContentBlock struct {
	Type        string         `json:"type"` // "text", "tool_use", "tool_result"
	Text        string         `json:"text,omitempty"`
	ToolName    string         `json:"tool_name,omitempty"`
	ToolInput   map[string]any `json:"tool_input,omitempty"`
	ToolUseID   string         `json:"tool_use_id,omitempty"`
	ToolContent string         `json:"tool_content,omitempty"`
}

// MessageMetadata contains message-level metadata
type MessageMetadata struct {
	IsSidechain bool              `json:"is_sidechain,omitempty"`
	AgentID     string            `json:"agent_id,omitempty"`
	Tokens      *TokenUsage       `json:"tokens"`
	Model       string            `json:"model,omitempty"`
	RequestID   string            `json:"request_id,omitempty"`
	Extra       map[string]string `json:"extra,omitempty"`
}

// TokenUsage represents token usage statistics
type TokenUsage struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
	TotalTokens  int `json:"total_tokens,omitempty"`
}

// Subagent represents a subagent conversation
type Subagent struct {
	AgentID  string    `json:"agent_id"`
	Slug     string    `json:"slug,omitempty"`
	Messages []Message `json:"messages"`
}
