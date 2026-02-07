package filter

import (
	"strings"
	"time"

	"github.com/block/braindump/internal/model"
)

// Options contains filtering options
type Options struct {
	AgentType string
	SessionID string
	Since     time.Time
	Until     time.Time
}

// Apply applies filters to sessions
func Apply(sessions []model.Session, opts Options) []model.Session {
	var filtered []model.Session

	for _, session := range sessions {
		if shouldInclude(session, opts) {
			filtered = append(filtered, session)
		}
	}

	return filtered
}

// shouldInclude checks if a session should be included based on filters
func shouldInclude(session model.Session, opts Options) bool {
	// Filter by agent type
	if opts.AgentType != "" {
		if !strings.EqualFold(session.AgentType, opts.AgentType) {
			return false
		}
	}

	// Filter by session ID
	if opts.SessionID != "" {
		if session.SessionID != opts.SessionID {
			return false
		}
	}

	// Filter by date range
	if !opts.Since.IsZero() {
		if session.CreatedAt.Before(opts.Since) {
			return false
		}
	}

	if !opts.Until.IsZero() {
		if session.CreatedAt.After(opts.Until) {
			return false
		}
	}

	return true
}
