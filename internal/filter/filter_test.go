package filter

import (
	"testing"
	"time"

	"github.com/block/braindump/internal/model"
)

func TestApply(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	sessions := []model.Session{
		{
			AgentType: "claude",
			SessionID: "session-1",
			CreatedAt: now,
		},
		{
			AgentType: "goose",
			SessionID: "session-2",
			CreatedAt: yesterday,
		},
		{
			AgentType: "claude",
			SessionID: "session-3",
			CreatedAt: tomorrow,
		},
	}

	tests := []struct {
		name     string
		opts     Options
		expected int
	}{
		{
			name:     "no filters",
			opts:     Options{},
			expected: 3,
		},
		{
			name: "filter by agent type - claude",
			opts: Options{
				AgentType: "claude",
			},
			expected: 2,
		},
		{
			name: "filter by agent type - goose",
			opts: Options{
				AgentType: "goose",
			},
			expected: 1,
		},
		{
			name: "filter by session ID",
			opts: Options{
				SessionID: "session-2",
			},
			expected: 1,
		},
		{
			name: "filter by since date",
			opts: Options{
				Since: now.Add(-1 * time.Hour),
			},
			expected: 2, // now and tomorrow
		},
		{
			name: "filter by until date",
			opts: Options{
				Until: now.Add(1 * time.Hour),
			},
			expected: 2, // yesterday and now
		},
		{
			name: "combined filters",
			opts: Options{
				AgentType: "claude",
				Since:     yesterday,
				Until:     now.Add(1 * time.Hour),
			},
			expected: 1, // only session-1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Apply(sessions, tt.opts)

			if len(result) != tt.expected {
				t.Errorf("Expected %d sessions, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestShouldInclude(t *testing.T) {
	now := time.Now()
	session := model.Session{
		AgentType: "claude",
		SessionID: "test-session",
		CreatedAt: now,
	}

	tests := []struct {
		name     string
		opts     Options
		expected bool
	}{
		{
			name:     "no filters - include",
			opts:     Options{},
			expected: true,
		},
		{
			name: "matching agent type",
			opts: Options{
				AgentType: "claude",
			},
			expected: true,
		},
		{
			name: "non-matching agent type",
			opts: Options{
				AgentType: "goose",
			},
			expected: false,
		},
		{
			name: "matching session ID",
			opts: Options{
				SessionID: "test-session",
			},
			expected: true,
		},
		{
			name: "non-matching session ID",
			opts: Options{
				SessionID: "other-session",
			},
			expected: false,
		},
		{
			name: "since before created - include",
			opts: Options{
				Since: now.Add(-1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "since after created - exclude",
			opts: Options{
				Since: now.Add(1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "until after created - include",
			opts: Options{
				Until: now.Add(1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "until before created - exclude",
			opts: Options{
				Until: now.Add(-1 * time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldInclude(session, tt.opts)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
