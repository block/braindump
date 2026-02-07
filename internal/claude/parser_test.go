package claude

import (
	"testing"
	"time"

	"github.com/block/braindump/internal/model"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name     string
		raw      map[string]any
		expected *model.Message
	}{
		{
			name: "simple user message",
			raw: map[string]any{
				"uuid":       "test-uuid",
				"parentUuid": "parent-uuid",
				"timestamp":  "2026-02-07T12:00:00Z",
				"message": map[string]any{
					"role":    "user",
					"content": "Hello, world!",
				},
			},
			expected: &model.Message{
				UUID:       "test-uuid",
				ParentUUID: "parent-uuid",
				Timestamp:  mustParseTime("2026-02-07T12:00:00Z"),
				Role:       "user",
				Content: []model.ContentBlock{
					{
						Type: "text",
						Text: "Hello, world!",
					},
				},
			},
		},
		{
			name: "assistant message with tool use",
			raw: map[string]any{
				"uuid":      "test-uuid",
				"timestamp": "2026-02-07T12:00:00Z",
				"message": map[string]any{
					"role": "assistant",
					"content": []any{
						map[string]any{
							"type": "tool_use",
							"name": "Read",
							"id":   "tool-123",
							"input": map[string]any{
								"file_path": "/test/path",
							},
						},
					},
					"model": "claude-sonnet-4-5",
				},
				"requestId": "req-123",
			},
			expected: &model.Message{
				UUID:      "test-uuid",
				Timestamp: mustParseTime("2026-02-07T12:00:00Z"),
				Role:      "assistant",
				Content: []model.ContentBlock{
					{
						Type:      "tool_use",
						ToolName:  "Read",
						ToolUseID: "tool-123",
						ToolInput: map[string]any{
							"file_path": "/test/path",
						},
					},
				},
				Metadata: model.MessageMetadata{
					Model:     "claude-sonnet-4-5",
					RequestID: "req-123",
				},
			},
		},
		{
			name: "tool result message",
			raw: map[string]any{
				"uuid":      "test-uuid",
				"timestamp": "2026-02-07T12:00:00Z",
				"message": map[string]any{
					"role": "user",
					"content": []any{
						map[string]any{
							"type":        "tool_result",
							"tool_use_id": "tool-123",
							"content":     "File contents here",
						},
					},
				},
			},
			expected: &model.Message{
				UUID:      "test-uuid",
				Timestamp: mustParseTime("2026-02-07T12:00:00Z"),
				Role:      "user",
				Content: []model.ContentBlock{
					{
						Type:        "tool_result",
						ToolUseID:   "tool-123",
						ToolContent: "File contents here",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMessage(tt.raw)

			if result == nil {
				t.Fatal("parseMessage returned nil")
			}

			if result.UUID != tt.expected.UUID {
				t.Errorf("UUID: got %q, want %q", result.UUID, tt.expected.UUID)
			}

			if result.ParentUUID != tt.expected.ParentUUID {
				t.Errorf("ParentUUID: got %q, want %q", result.ParentUUID, tt.expected.ParentUUID)
			}

			if result.Role != tt.expected.Role {
				t.Errorf("Role: got %q, want %q", result.Role, tt.expected.Role)
			}

			if len(result.Content) != len(tt.expected.Content) {
				t.Errorf("Content length: got %d, want %d", len(result.Content), len(tt.expected.Content))
			}
		})
	}
}

func TestParseContentBlock(t *testing.T) {
	tests := []struct {
		name     string
		block    map[string]any
		expected *model.ContentBlock
	}{
		{
			name: "text block",
			block: map[string]any{
				"type": "text",
				"text": "Hello",
			},
			expected: &model.ContentBlock{
				Type: "text",
				Text: "Hello",
			},
		},
		{
			name: "tool use block",
			block: map[string]any{
				"type": "tool_use",
				"name": "Read",
				"id":   "tool-123",
				"input": map[string]any{
					"file_path": "/test",
				},
			},
			expected: &model.ContentBlock{
				Type:      "tool_use",
				ToolName:  "Read",
				ToolUseID: "tool-123",
				ToolInput: map[string]any{
					"file_path": "/test",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseContentBlock(tt.block)

			if result == nil {
				t.Fatal("parseContentBlock returned nil")
			}

			if result.Type != tt.expected.Type {
				t.Errorf("Type: got %q, want %q", result.Type, tt.expected.Type)
			}

			if result.Text != tt.expected.Text {
				t.Errorf("Text: got %q, want %q", result.Text, tt.expected.Text)
			}

			if result.ToolName != tt.expected.ToolName {
				t.Errorf("ToolName: got %q, want %q", result.ToolName, tt.expected.ToolName)
			}
		})
	}
}

func TestParseTokenUsage(t *testing.T) {
	usage := map[string]any{
		"input_tokens":  float64(100),
		"output_tokens": float64(50),
	}

	result := parseTokenUsage(usage)

	if result.InputTokens != 100 {
		t.Errorf("InputTokens: got %d, want 100", result.InputTokens)
	}

	if result.OutputTokens != 50 {
		t.Errorf("OutputTokens: got %d, want 50", result.OutputTokens)
	}

	if result.TotalTokens != 150 {
		t.Errorf("TotalTokens: got %d, want 150", result.TotalTokens)
	}
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
