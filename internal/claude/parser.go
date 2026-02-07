package claude

import (
	"strings"
	"time"

	"github.com/block/braindump/internal/model"
)

// parseMessage parses a Claude message from raw JSON
func parseMessage(raw map[string]any) *model.Message {
	uuid, _ := raw["uuid"].(string)
	parentUUID, _ := raw["parentUuid"].(string)

	var timestamp time.Time
	if tsStr, ok := raw["timestamp"].(string); ok {
		timestamp, _ = time.Parse(time.RFC3339, tsStr)
	}

	// Extract message content
	messageData, ok := raw["message"].(map[string]any)
	if !ok {
		return nil
	}

	role, _ := messageData["role"].(string)
	if role == "" {
		return nil
	}

	// Parse content blocks
	var contentBlocks []model.ContentBlock

	// Handle both string content and array content
	if contentStr, isString := messageData["content"].(string); isString {
		contentBlocks = []model.ContentBlock{
			{
				Type: "text",
				Text: contentStr,
			},
		}
	} else if contentArr, isArray := messageData["content"].([]any); isArray {
		for _, item := range contentArr {
			if block, isBlock := item.(map[string]any); isBlock {
				contentBlock := parseContentBlock(block)
				if contentBlock != nil {
					contentBlocks = append(contentBlocks, *contentBlock)
				}
			}
		}
	}

	// Build metadata
	metadata := model.MessageMetadata{}

	if isSidechain, hasSidechain := raw["isSidechain"].(bool); hasSidechain {
		metadata.IsSidechain = isSidechain
	}

	if agentID, hasAgentID := raw["agentId"].(string); hasAgentID {
		metadata.AgentID = agentID
	}

	if requestID, hasRequestID := raw["requestId"].(string); hasRequestID {
		metadata.RequestID = requestID
	}

	// Extract model from message
	if modelStr, hasModel := messageData["model"].(string); hasModel {
		metadata.Model = modelStr
	}

	// Extract token usage
	if usage, hasUsage := messageData["usage"].(map[string]any); hasUsage {
		metadata.Tokens = parseTokenUsage(usage)
	}

	return &model.Message{
		UUID:       uuid,
		ParentUUID: parentUUID,
		Timestamp:  timestamp,
		Role:       role,
		Content:    contentBlocks,
		Metadata:   metadata,
	}
}

// parseContentBlock parses a content block
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
		toolContent := extractToolContent(block["content"])

		return &model.ContentBlock{
			Type:        "tool_result",
			ToolUseID:   toolUseID,
			ToolContent: toolContent,
		}
	}

	return nil
}

// extractToolContent extracts content from tool result, handling both string and array formats
func extractToolContent(content any) string {
	// Content can be string or array
	if contentStr, isString := content.(string); isString {
		return contentStr
	}

	contentArr, isArray := content.([]any)
	if !isArray {
		return ""
	}

	// Join array elements
	var builder strings.Builder
	for i, item := range contentArr {
		if i > 0 {
			builder.WriteString("\n")
		}

		if itemMap, isMap := item.(map[string]any); isMap {
			if text, hasText := itemMap["text"].(string); hasText {
				builder.WriteString(text)
			}
		} else if str, isStr := item.(string); isStr {
			builder.WriteString(str)
		}
	}

	return builder.String()
}

// parseTokenUsage parses token usage from Claude message
func parseTokenUsage(usage map[string]any) *model.TokenUsage {
	tokens := &model.TokenUsage{}

	if inputTokens, ok := usage["input_tokens"].(float64); ok {
		tokens.InputTokens = int(inputTokens)
	}

	if outputTokens, ok := usage["output_tokens"].(float64); ok {
		tokens.OutputTokens = int(outputTokens)
	}

	// Calculate total
	tokens.TotalTokens = tokens.InputTokens + tokens.OutputTokens

	return tokens
}
