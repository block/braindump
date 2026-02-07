package goose

import (
	"encoding/json"

	"github.com/block/braindump/internal/model"
)

// parseContent parses Goose content JSON into content blocks
func parseContent(contentJSON string) []model.ContentBlock {
	var content any
	if err := json.Unmarshal([]byte(contentJSON), &content); err != nil {
		return nil
	}

	var blocks []model.ContentBlock

	// Handle different content formats
	switch v := content.(type) {
	case string:
		// Simple string content
		blocks = append(blocks, model.ContentBlock{
			Type: "text",
			Text: v,
		})

	case []any:
		// Array of content blocks
		for _, item := range v {
			if block, isBlock := item.(map[string]any); isBlock {
				contentBlock := parseContentBlock(block)
				if contentBlock != nil {
					blocks = append(blocks, *contentBlock)
				}
			} else if str, isStr := item.(string); isStr {
				blocks = append(blocks, model.ContentBlock{
					Type: "text",
					Text: str,
				})
			}
		}

	case map[string]any:
		// Single content block
		contentBlock := parseContentBlock(v)
		if contentBlock != nil {
			blocks = append(blocks, *contentBlock)
		}
	}

	return blocks
}

// parseContentBlock parses a single content block
func parseContentBlock(block map[string]any) *model.ContentBlock {
	blockType, _ := block["type"].(string)

	switch blockType {
	case "text":
		text, _ := block["text"].(string)
		return &model.ContentBlock{
			Type: "text",
			Text: text,
		}

	case "tool_use":
		toolName, _ := block["name"].(string)
		toolUseID, _ := block["id"].(string)
		toolInput, _ := block["input"].(map[string]any)

		return &model.ContentBlock{
			Type:      "tool_use",
			ToolName:  toolName,
			ToolUseID: toolUseID,
			ToolInput: toolInput,
		}

	case "tool_result":
		toolUseID, _ := block["tool_use_id"].(string)

		// Content can be string or nested
		var toolContent string
		if contentStr, ok := block["content"].(string); ok {
			toolContent = contentStr
		}

		return &model.ContentBlock{
			Type:        "tool_result",
			ToolUseID:   toolUseID,
			ToolContent: toolContent,
		}

	default:
		// For unknown types, try to extract text field
		if text, ok := block["text"].(string); ok {
			return &model.ContentBlock{
				Type: blockType,
				Text: text,
			}
		}
	}

	return nil
}
