package goose

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/block/braindump/internal/model"
	_ "modernc.org/sqlite" // Import sqlite driver for database/sql
)

// Reader handles reading Goose sessions from SQLite
type Reader struct {
	homeDir string
}

// NewReader creates a new Goose reader
func NewReader() (*Reader, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	return &Reader{homeDir: homeDir}, nil
}

// ReadSessions reads all Goose sessions from the SQLite database
func (r *Reader) ReadSessions() ([]model.Session, error) {
	dbPath := filepath.Join(r.homeDir, ".local", "share", "goose", "sessions", "sessions.db")

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return []model.Session{}, nil // No Goose sessions
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Query sessions
	sessionsQuery := `
		SELECT id, name, description, user_set_name, session_type, working_dir,
		       created_at, updated_at, extension_data, provider_name, model_config_json
		FROM sessions
		ORDER BY created_at DESC
	`

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, sessionsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.Session

	for rows.Next() {
		var (
			id              int
			name            sql.NullString
			description     sql.NullString
			userSetName     sql.NullString
			sessionType     sql.NullString
			workingDir      sql.NullString
			createdAt       sql.NullString
			updatedAt       sql.NullString
			extensionData   sql.NullString
			providerName    sql.NullString
			modelConfigJSON sql.NullString
		)

		err := rows.Scan(&id, &name, &description, &userSetName, &sessionType,
			&workingDir, &createdAt, &updatedAt, &extensionData, &providerName, &modelConfigJSON)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to scan session row: %v\n", err)
			continue
		}

		// Parse timestamps
		createdTime, _ := time.Parse(time.RFC3339, createdAt.String)
		updatedTime, _ := time.Parse(time.RFC3339, updatedAt.String)

		// Parse model config
		var modelConfig map[string]any
		if modelConfigJSON.Valid && modelConfigJSON.String != "" {
			_ = json.Unmarshal([]byte(modelConfigJSON.String), &modelConfig)
		}

		// Extract model name
		modelName := ""
		if modelConfig != nil {
			if model, ok := modelConfig["model"].(string); ok {
				modelName = model
			}
		}

		// Build metadata
		metadata := model.SessionMetadata{
			WorkingDir: workingDir.String,
			Provider:   providerName.String,
			Model:      modelName,
			Name:       name.String,
		}

		// Add extra metadata
		metadata.Extra = make(map[string]string)
		if userSetName.Valid {
			metadata.Extra["user_set_name"] = userSetName.String
		}
		if sessionType.Valid {
			metadata.Extra["session_type"] = sessionType.String
		}
		if description.Valid {
			metadata.Extra["description"] = description.String
		}

		// Read messages for this session
		messages, err := r.readMessages(db, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read messages for session %d: %v\n", id, err)
			continue
		}

		sessions = append(sessions, model.Session{
			AgentType: "goose",
			SessionID: fmt.Sprintf("%d", id),
			CreatedAt: createdTime,
			UpdatedAt: updatedTime,
			Metadata:  metadata,
			Messages:  messages,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// readMessages reads messages for a specific session
func (r *Reader) readMessages(db *sql.DB, sessionID int) ([]model.Message, error) {
	messagesQuery := `
		SELECT id, message_id, role, content_json, created_timestamp, tokens, metadata_json
		FROM messages
		WHERE session_id = ?
		ORDER BY created_timestamp ASC
	`

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, messagesQuery, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message

	for rows.Next() {
		var (
			id               int
			messageID        sql.NullString
			role             string
			contentJSON      sql.NullString
			createdTimestamp sql.NullString
			tokens           sql.NullInt64
			metadataJSON     sql.NullString
		)

		err := rows.Scan(&id, &messageID, &role, &contentJSON, &createdTimestamp, &tokens, &metadataJSON)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to scan message row: %v\n", err)
			continue
		}

		// Parse timestamp
		timestamp, _ := time.Parse(time.RFC3339, createdTimestamp.String)

		// Parse content
		var contentBlocks []model.ContentBlock
		if contentJSON.Valid && contentJSON.String != "" {
			contentBlocks = parseContent(contentJSON.String)
		}

		// Build metadata
		metadata := model.MessageMetadata{}
		if tokens.Valid {
			metadata.Tokens = &model.TokenUsage{
				TotalTokens: int(tokens.Int64),
			}
		}

		// Parse additional metadata from metadata_json
		if metadataJSON.Valid && metadataJSON.String != "" {
			var extraMetadata map[string]any
			if err := json.Unmarshal([]byte(metadataJSON.String), &extraMetadata); err == nil {
				metadata.Extra = make(map[string]string)
				for k, v := range extraMetadata {
					if str, ok := v.(string); ok {
						metadata.Extra[k] = str
					}
				}
			}
		}

		messages = append(messages, model.Message{
			UUID:      messageID.String,
			Timestamp: timestamp,
			Role:      role,
			Content:   contentBlocks,
			Metadata:  metadata,
		})
	}

	return messages, rows.Err()
}
