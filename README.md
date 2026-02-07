# braindump

A Go CLI tool to dump AI agent session histories (Claude Code and Goose) to stdout in a unified JSON format.

## Overview

`braindump` reads conversation histories from Claude Code and Goose AI agents stored locally on your system and outputs them in a consistent, structured JSON format. This makes it easy to analyze, archive, or process agent interactions programmatically.

## Features

- **Multi-Agent Support**: Reads sessions from both Claude Code and Goose AI
- **Unified Format**: Standardized JSON schema across different agent types
- **Filtering**: Filter sessions by agent type, session ID, or date range
- **Complete History**: Includes messages, tool calls, tool results, and metadata
- **Subagent Support**: Captures subagent conversations from Claude Code
- **No External Dependencies**: Pure Go implementation with no CGo requirements

## Installation

### Prerequisites

- Go 1.21 or higher

### From Source

```bash
git clone https://github.com/block/braindump.git
cd braindump
go build -o braindump ./cmd/braindump
```

### Using Hermit

```bash
hermit init
hermit install go
source bin/activate-hermit
go build -o braindump ./cmd/braindump
```

## Usage

### Basic Usage

Dump all agent sessions to stdout:

```bash
./braindump
```

Pretty-print the JSON output:

```bash
./braindump --pretty
```

Get a human-readable summary of sessions:

```bash
./braindump --summary
```

The summary output includes:
- Session metadata (ID, agent type, creation date, model)
- Initial user prompt
- Last user prompt
- Last 2 agent messages
- Message statistics

### Filtering Options

Filter by agent type:

```bash
./braindump --agent claude
./braindump --agent goose
```

Filter by specific session ID:

```bash
./braindump --session-id ae52213c-04a4-49ab-b17c-01641c246f7d
```

Filter by date range:

```bash
# Sessions since a specific date
./braindump --since 2026-01-01T00:00:00Z

# Sessions until a specific date
./braindump --until 2026-02-01T00:00:00Z

# Sessions within a date range
./braindump --since 2026-01-01T00:00:00Z --until 2026-02-01T00:00:00Z
```

### Output Options

Save output to a file:

```bash
./braindump --output sessions.json
./braindump -o sessions.json
```

Combine filters:

```bash
./braindump --agent claude --since 2026-01-01T00:00:00Z --pretty -o claude-sessions.json
```

## Command-Line Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--agent` | Filter by agent type (claude, goose) | `--agent claude` |
| `--session-id` | Filter by specific session ID | `--session-id abc123` |
| `--since` | Filter sessions since timestamp (RFC3339) | `--since 2026-01-01T00:00:00Z` |
| `--until` | Filter sessions until timestamp (RFC3339) | `--until 2026-02-01T00:00:00Z` |
| `-o, --output` | Output file (default: stdout) | `-o sessions.json` |
| `--pretty` | Pretty-print JSON output | `--pretty` |
| `--summary` | Output human-readable summary instead of JSON | `--summary` |
| `--help` | Show help message | `--help` |

## Output Schema

The tool outputs a single JSON document with the following structure:

### Root Object

```json
{
  "version": "1.0.0",
  "generated_at": "2026-02-07T00:00:00Z",
  "sessions": [...]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Schema version (currently "1.0.0") |
| `generated_at` | timestamp | When the dump was generated (RFC3339) |
| `sessions` | array | Array of session objects |

### Session Object

```json
{
  "agent_type": "claude",
  "session_id": "ae52213c-04a4-49ab-b17c-01641c246f7d",
  "created_at": "2026-02-06T21:43:52.446Z",
  "updated_at": "2026-02-07T19:53:53.265Z",
  "metadata": {...},
  "messages": [...],
  "subagents": [...]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `agent_type` | string | Agent type: "claude" or "goose" |
| `session_id` | string | Unique session identifier |
| `created_at` | timestamp | Session creation time (RFC3339) |
| `updated_at` | timestamp | Last update time (RFC3339) |
| `metadata` | object | Session metadata (see below) |
| `messages` | array | Array of message objects |
| `subagents` | array | Array of subagent objects (Claude only) |

### Session Metadata

```json
{
  "working_dir": "/home/user/project",
  "git_branch": "main",
  "model": "claude-sonnet-4-5",
  "provider": "anthropic",
  "name": "Session Name",
  "extra": {...}
}
```

| Field | Type | Description |
|-------|------|-------------|
| `working_dir` | string | Working directory for the session |
| `git_branch` | string | Git branch (Claude only) |
| `model` | string | Model name (e.g., "claude-sonnet-4-5") |
| `provider` | string | Provider name (e.g., "anthropic", "databricks") |
| `name` | string | Session name (Goose only) |
| `extra` | object | Additional metadata key-value pairs |

### Message Object

```json
{
  "uuid": "c84ebe8b-c27d-48de-896d-69f917d7478e",
  "parent_uuid": "566fd1b2-8409-456f-92ca-a6f80ebc88d2",
  "timestamp": "2026-02-06T21:43:52.499Z",
  "role": "user",
  "content": [...],
  "metadata": {...}
}
```

| Field | Type | Description |
|-------|------|-------------|
| `uuid` | string | Unique message identifier |
| `parent_uuid` | string | Parent message UUID (for threading) |
| `timestamp` | timestamp | Message timestamp (RFC3339) |
| `role` | string | Message role: "user" or "assistant" |
| `content` | array | Array of content blocks |
| `metadata` | object | Message metadata |

### Content Block

Content blocks represent different types of message content:

**Text Block:**
```json
{
  "type": "text",
  "text": "Hello, world!"
}
```

**Tool Use Block:**
```json
{
  "type": "tool_use",
  "tool_name": "Read",
  "tool_use_id": "toolu_123",
  "tool_input": {
    "file_path": "/path/to/file"
  }
}
```

**Tool Result Block:**
```json
{
  "type": "tool_result",
  "tool_use_id": "toolu_123",
  "tool_content": "File contents here..."
}
```

### Message Metadata

```json
{
  "is_sidechain": false,
  "agent_id": "a2367c4",
  "tokens": {
    "input_tokens": 100,
    "output_tokens": 50,
    "total_tokens": 150
  },
  "model": "claude-sonnet-4-5",
  "request_id": "req_123"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `is_sidechain` | boolean | Whether this is a subagent message |
| `agent_id` | string | Subagent identifier |
| `tokens` | object | Token usage statistics |
| `model` | string | Model used for this message |
| `request_id` | string | API request ID |

### Subagent Object

```json
{
  "agent_id": "a2367c4",
  "slug": "quirky-popping-kernighan",
  "messages": [...]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `agent_id` | string | Subagent identifier |
| `slug` | string | Human-readable subagent name |
| `messages` | array | Array of message objects |

## Data Sources

### Claude Code

- **Location**: `~/.claude/projects/*/`
- **Format**: JSONL (newline-delimited JSON)
- **Files**:
  - Main sessions: `{sessionId}.jsonl`
  - Subagent sessions: `{sessionId}/subagents/agent-{agentId}.jsonl`

### Goose AI

- **Location**: `~/.local/share/goose/sessions/sessions.db`
- **Format**: SQLite database
- **Tables**: `sessions`, `messages`

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o braindump ./cmd/braindump
```

### Project Structure

```
.
├── cmd/
│   └── braindump/
│       └── main.go              # CLI entry point
├── internal/
│   ├── model/
│   │   └── types.go             # Unified data structures
│   ├── claude/
│   │   ├── reader.go            # Claude session reader
│   │   ├── parser.go            # Claude format parser
│   │   └── parser_test.go       # Parser tests
│   ├── goose/
│   │   ├── reader.go            # Goose SQLite reader
│   │   ├── parser.go            # Goose format parser
│   │   └── parser_test.go       # Parser tests
│   ├── filter/
│   │   ├── filter.go            # Session filtering
│   │   └── filter_test.go       # Filter tests
│   └── output/
│       └── writer.go            # JSON output writer
├── go.mod
├── go.sum
└── README.md
```

## Examples

### Example 1: Quick Summary of Recent Sessions

```bash
./braindump --summary
```

### Example 2: Archive All Sessions

```bash
./braindump --pretty -o archive-$(date +%Y%m%d).json
```

### Example 3: Extract Claude Sessions from Last Week

```bash
./braindump --agent claude --since $(date -d '7 days ago' -Iseconds) --pretty
```

### Example 4: Analyze Token Usage

```bash
./braindump | jq '[.sessions[].messages[].metadata.tokens.total_tokens] | add'
```

### Example 5: List All Tool Calls

```bash
./braindump | jq -r '.sessions[].messages[].content[] | select(.type=="tool_use") | .tool_name' | sort | uniq -c
```

### Example 6: Extract Specific Session

```bash
./braindump --session-id ae52213c-04a4-49ab-b17c-01641c246f7d --pretty
```

## License

Apache License, Version 2.0

## Contributing

See [GOVERNANCE.md](./GOVERNANCE.md) for contribution guidelines.
