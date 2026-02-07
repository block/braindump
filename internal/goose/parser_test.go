package goose

import (
	"testing"

	"github.com/block/braindump/internal/model"
)

func TestParseContent(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected int // expected number of content blocks
	}{
		{
			name:     "simple string",
			json:     `"Hello, world!"`,
			expected: 1,
		},
		{
			name: "array of blocks",
			json: `[
				{"type": "text", "text": "Hello"},
				{"type": "text", "text": "World"}
			]`,
			expected: 2,
		},
		{
			name:     "single block object",
			json:     `{"type": "text", "text": "Hello"}`,
			expected: 1,
		},
		{
			name:     "invalid json",
			json:     `{invalid}`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseContent(tt.json)

			if len(result) != tt.expected {
				t.Errorf("Expected %d blocks, got %d", tt.expected, len(result))
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
				"name": "bash",
				"id":   "tool-456",
				"input": map[string]any{
					"command": "ls -la",
				},
			},
			expected: &model.ContentBlock{
				Type:      "tool_use",
				ToolName:  "bash",
				ToolUseID: "tool-456",
				ToolInput: map[string]any{
					"command": "ls -la",
				},
			},
		},
		{
			name: "tool result block",
			block: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool-456",
				"content":     "file1.txt\nfile2.txt",
			},
			expected: &model.ContentBlock{
				Type:        "tool_result",
				ToolUseID:   "tool-456",
				ToolContent: "file1.txt\nfile2.txt",
			},
		},
		{
			name: "unknown type with text",
			block: map[string]any{
				"type": "custom",
				"text": "Custom content",
			},
			expected: &model.ContentBlock{
				Type: "custom",
				Text: "Custom content",
			},
		},
		{
			name: "unknown type without text",
			block: map[string]any{
				"type": "unknown",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseContentBlock(tt.block)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %+v", result)
				}
				return
			}

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
