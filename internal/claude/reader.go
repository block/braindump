package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/block/braindump/internal/model"
)

// Reader handles reading Claude session files
type Reader struct {
	homeDir string
}

// NewReader creates a new Claude reader
func NewReader() (*Reader, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	return &Reader{homeDir: homeDir}, nil
}

// ReadSessions reads all Claude sessions
func (r *Reader) ReadSessions() ([]model.Session, error) {
	claudeDir := filepath.Join(r.homeDir, ".claude", "projects")

	// Check if directory exists
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return []model.Session{}, nil // No Claude sessions
	}

	var sessions []model.Session

	// Walk through all project directories
	err := filepath.Walk(claudeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for .jsonl files (not in subagents directory)
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".jsonl") {
			// Skip subagent files (they're in subdirectories)
			if strings.Contains(path, "subagents") {
				return nil
			}

			session, err := r.readSessionFile(path)
			if err != nil {
				// Log error but continue processing other sessions
				fmt.Fprintf(os.Stderr, "Warning: failed to read session %s: %v\n", path, err)
				return nil
			}
			if session != nil {
				sessions = append(sessions, *session)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk Claude directory: %w", err)
	}

	return sessions, nil
}

// readSessionFile reads a single Claude session file
func (r *Reader) readSessionFile(path string) (*model.Session, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var messages []model.Message
	var sessionID string
	var createdAt, updatedAt time.Time
	var metadata model.SessionMetadata

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var raw map[string]any
		if err := json.Unmarshal(line, &raw); err != nil {
			// Skip malformed lines
			continue
		}

		// Extract session metadata from first message
		if sessionID == "" {
			if sid, ok := raw["sessionId"].(string); ok {
				sessionID = sid
			}
			if cwd, ok := raw["cwd"].(string); ok {
				metadata.WorkingDir = cwd
			}
			if branch, ok := raw["gitBranch"].(string); ok {
				metadata.GitBranch = branch
			}
		}

		// Parse timestamp
		if tsStr, ok := raw["timestamp"].(string); ok {
			ts, _ := time.Parse(time.RFC3339, tsStr)
			if createdAt.IsZero() || ts.Before(createdAt) {
				createdAt = ts
			}
			if updatedAt.IsZero() || ts.After(updatedAt) {
				updatedAt = ts
			}
		}

		// Parse message
		msgType, _ := raw["type"].(string)
		if msgType == "user" || msgType == "assistant" {
			msg := parseMessage(raw)
			if msg != nil {
				messages = append(messages, *msg)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// If no sessionID found, use filename
	if sessionID == "" {
		sessionID = strings.TrimSuffix(filepath.Base(path), ".jsonl")
	}

	// Read subagents
	subagents, err := r.readSubagents(filepath.Dir(path), sessionID)
	if err != nil {
		// Log but don't fail
		fmt.Fprintf(os.Stderr, "Warning: failed to read subagents for %s: %v\n", sessionID, err)
	}

	return &model.Session{
		AgentType: "claude",
		SessionID: sessionID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Metadata:  metadata,
		Messages:  messages,
		Subagents: subagents,
	}, nil
}

// readSubagents reads subagent sessions
func (r *Reader) readSubagents(sessionDir, sessionID string) ([]model.Subagent, error) {
	subagentsDir := filepath.Join(sessionDir, sessionID, "subagents")

	if _, err := os.Stat(subagentsDir); os.IsNotExist(err) {
		return nil, nil // No subagents
	}

	var subagents []model.Subagent

	files, err := os.ReadDir(subagentsDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".jsonl") {
			agentID := strings.TrimPrefix(strings.TrimSuffix(file.Name(), ".jsonl"), "agent-")

			path := filepath.Join(subagentsDir, file.Name())
			messages, slug, err := r.readSubagentFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to read subagent %s: %v\n", agentID, err)
				continue
			}

			subagents = append(subagents, model.Subagent{
				AgentID:  agentID,
				Slug:     slug,
				Messages: messages,
			})
		}
	}

	return subagents, nil
}

// readSubagentFile reads a subagent session file
func (r *Reader) readSubagentFile(path string) ([]model.Message, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	var messages []model.Message
	var slug string

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var raw map[string]any
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}

		// Extract slug from first message
		if slug == "" {
			if s, ok := raw["slug"].(string); ok {
				slug = s
			}
		}

		msgType, _ := raw["type"].(string)
		if msgType == "user" || msgType == "assistant" {
			msg := parseMessage(raw)
			if msg != nil {
				messages = append(messages, *msg)
			}
		}
	}

	return messages, slug, scanner.Err()
}
