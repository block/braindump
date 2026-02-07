package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/block/braindump/internal/model"
)

// SummaryWriter handles writing human-readable summaries
type SummaryWriter struct {
	writer io.Writer
}

// NewSummaryWriter creates a new summary writer
func NewSummaryWriter(w io.Writer) *SummaryWriter {
	return &SummaryWriter{writer: w}
}

// Write writes session summaries to output
func (w *SummaryWriter) Write(sessions []model.Session) error {
	if len(sessions) == 0 {
		_, err := fmt.Fprintln(w.writer, "No sessions found.")
		return err
	}

	for i, session := range sessions {
		if i > 0 {
			_, err := fmt.Fprintln(w.writer, "\n"+strings.Repeat("=", 80))
			if err != nil {
				return err
			}
		}

		if err := w.writeSessionSummary(session); err != nil {
			return err
		}
	}

	return nil
}

// writeSessionSummary writes a summary for a single session
func (w *SummaryWriter) writeSessionSummary(session model.Session) error {
	// Header
	_, err := fmt.Fprintf(w.writer, "\n📋 Session: %s\n", session.SessionID)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w.writer, "   Agent: %s\n", session.AgentType)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w.writer, "   Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}

	if session.Metadata.WorkingDir != "" {
		_, err = fmt.Fprintf(w.writer, "   Working Dir: %s\n", session.Metadata.WorkingDir)
		if err != nil {
			return err
		}
	}

	if session.Metadata.Model != "" {
		_, err = fmt.Fprintf(w.writer, "   Model: %s\n", session.Metadata.Model)
		if err != nil {
			return err
		}
	}

	// Find key messages
	firstUserMsg := findFirstUserMessage(session.Messages)
	lastUserMsg := findLastUserMessage(session.Messages)
	lastAgentMsgs := findLastAgentMessages(session.Messages, 2)

	// Initial user prompt
	if firstUserMsg != nil {
		_, err = fmt.Fprintf(w.writer, "\n🚀 Initial User Prompt:\n")
		if err != nil {
			return err
		}

		content := extractMessageContent(firstUserMsg)
		wrapped := wrapText(content, 76)
		_, err = fmt.Fprintf(w.writer, "   %s\n", strings.ReplaceAll(wrapped, "\n", "\n   "))
		if err != nil {
			return err
		}
	}

	// Last user prompt (if different from first)
	if lastUserMsg != nil && (firstUserMsg == nil || lastUserMsg.UUID != firstUserMsg.UUID) {
		_, err = fmt.Fprintf(w.writer, "\n💬 Last User Prompt:\n")
		if err != nil {
			return err
		}

		content := extractMessageContent(lastUserMsg)
		wrapped := wrapText(content, 76)
		_, err = fmt.Fprintf(w.writer, "   %s\n", strings.ReplaceAll(wrapped, "\n", "\n   "))
		if err != nil {
			return err
		}
	}

	// Last agent messages
	if len(lastAgentMsgs) > 0 {
		_, err = fmt.Fprintf(w.writer, "\n🤖 Last %d Agent Message(s):\n", len(lastAgentMsgs))
		if err != nil {
			return err
		}

		for i, msg := range lastAgentMsgs {
			content := extractMessageContent(&msg)
			if content == "" {
				content = "[Tool use or non-text content]"
			}

			wrapped := wrapText(content, 74)
			_, err = fmt.Fprintf(w.writer, "\n   [%d] %s\n", i+1, strings.ReplaceAll(wrapped, "\n", "\n       "))
			if err != nil {
				return err
			}
		}
	}

	// Statistics
	userCount := countMessagesByRole(session.Messages, "user")
	assistantCount := countMessagesByRole(session.Messages, "assistant")

	_, err = fmt.Fprintf(w.writer, "\n📊 Statistics:\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w.writer, "   Total Messages: %d (User: %d, Agent: %d)\n", len(session.Messages), userCount, assistantCount)
	if err != nil {
		return err
	}

	if len(session.Subagents) > 0 {
		_, err = fmt.Fprintf(w.writer, "   Subagents: %d\n", len(session.Subagents))
		if err != nil {
			return err
		}
	}

	return nil
}

// findFirstUserMessage finds the first user message in the session
func findFirstUserMessage(messages []model.Message) *model.Message {
	for i := range messages {
		if messages[i].Role == "user" {
			return &messages[i]
		}
	}
	return nil
}

// findLastUserMessage finds the last user message in the session
func findLastUserMessage(messages []model.Message) *model.Message {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return &messages[i]
		}
	}
	return nil
}

// findLastAgentMessages finds the last N assistant messages
func findLastAgentMessages(messages []model.Message, count int) []model.Message {
	var result []model.Message

	for i := len(messages) - 1; i >= 0 && len(result) < count; i-- {
		if messages[i].Role == "assistant" {
			result = append([]model.Message{messages[i]}, result...)
		}
	}

	return result
}

// extractMessageContent extracts text content from a message
func extractMessageContent(msg *model.Message) string {
	var parts []string

	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			if block.Text != "" {
				parts = append(parts, block.Text)
			}
		case "tool_use":
			// Skip tool use in summary
		case "tool_result":
			// Skip tool results in summary
		}
	}

	return strings.Join(parts, " ")
}

// countMessagesByRole counts messages by role
func countMessagesByRole(messages []model.Message, role string) int {
	count := 0
	for _, msg := range messages {
		if msg.Role == role {
			count++
		}
	}
	return count
}

// wrapText wraps text to a maximum width
func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			result.WriteString(currentLine)
			result.WriteString("\n")
			currentLine = word
		}
	}

	result.WriteString(currentLine)
	return result.String()
}
