package main

import (
	"fmt"
	"os"
	"time"

	"github.com/block/braindump/internal/claude"
	"github.com/block/braindump/internal/filter"
	"github.com/block/braindump/internal/goose"
	"github.com/block/braindump/internal/model"
	"github.com/block/braindump/internal/output"
	"github.com/spf13/cobra"
)

var (
	agentType string
	sessionID string
	since     string
	until     string
	outFile   string
	pretty    bool
	summary   bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "braindump",
		Short: "Dump agent session histories to JSON",
		Long: `braindump reads Claude Code and Goose AI agent session histories
and outputs them in a unified JSON format.`,
		RunE: run,
	}

	rootCmd.Flags().StringVar(&agentType, "agent", "", "Filter by agent type (claude, goose)")
	rootCmd.Flags().StringVar(&sessionID, "session-id", "", "Filter by specific session ID")
	rootCmd.Flags().StringVar(&since, "since", "", "Filter sessions since timestamp (RFC3339)")
	rootCmd.Flags().StringVar(&until, "until", "", "Filter sessions until timestamp (RFC3339)")
	rootCmd.Flags().StringVarP(&outFile, "output", "o", "", "Output file (default: stdout)")
	rootCmd.Flags().BoolVar(&pretty, "pretty", false, "Pretty-print JSON output")
	rootCmd.Flags().BoolVar(&summary, "summary", false, "Output human-readable summary instead of JSON")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Parse time filters
	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = time.Parse(time.RFC3339, since)
		if err != nil {
			return fmt.Errorf("invalid --since timestamp: %w", err)
		}
	}

	if until != "" {
		untilTime, err = time.Parse(time.RFC3339, until)
		if err != nil {
			return fmt.Errorf("invalid --until timestamp: %w", err)
		}
	}

	// Read sessions from all sources
	var allSessions []model.Session

	// Read Claude sessions (if not filtered to goose only)
	if agentType == "" || agentType == "claude" {
		claudeReader, err := claude.NewReader()
		if err != nil {
			return fmt.Errorf("failed to create Claude reader: %w", err)
		}

		claudeSessions, err := claudeReader.ReadSessions()
		if err != nil {
			return fmt.Errorf("failed to read Claude sessions: %w", err)
		}

		allSessions = append(allSessions, claudeSessions...)
	}

	// Read Goose sessions (if not filtered to claude only)
	if agentType == "" || agentType == "goose" {
		gooseReader, err := goose.NewReader()
		if err != nil {
			return fmt.Errorf("failed to create Goose reader: %w", err)
		}

		gooseSessions, err := gooseReader.ReadSessions()
		if err != nil {
			return fmt.Errorf("failed to read Goose sessions: %w", err)
		}

		allSessions = append(allSessions, gooseSessions...)
	}

	// Apply filters
	filterOpts := filter.Options{
		AgentType: agentType,
		SessionID: sessionID,
		Since:     sinceTime,
		Until:     untilTime,
	}

	filteredSessions := filter.Apply(allSessions, filterOpts)

	// Write output
	var writer *os.File
	if outFile != "" {
		writer, err = os.Create(outFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer writer.Close()
	} else {
		writer = os.Stdout
	}

	// Choose output format
	if summary {
		summaryWriter := output.NewSummaryWriter(writer)
		if err := summaryWriter.Write(filteredSessions); err != nil {
			return fmt.Errorf("failed to write summary: %w", err)
		}
	} else {
		outputWriter := output.NewWriter(writer, pretty)
		if err := outputWriter.Write(filteredSessions); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}
