package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	log "github.com/nocodeleaks/quepasa/qplog"
)

// LogsTool implements log analysis tool for MCP (master key only)
type LogsTool struct{}

// LogsRequest represents the request for log analysis
type LogsRequest struct {
	Lines  int    `json:"lines,omitempty"`  // Number of lines to retrieve (default: 100, max: 1000)
	Level  string `json:"level,omitempty"`  // Filter by log level: trace, debug, info, warn, error (default: all)
	Search string `json:"search,omitempty"` // Search for keyword in log messages
	Tail   bool   `json:"tail,omitempty"`   // Real-time tail mode (not implemented yet)
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string `json:"timestamp,omitempty"`
	Level     string `json:"level,omitempty"`
	Module    string `json:"module,omitempty"`
	Message   string `json:"message"`
}

// LogsResponse represents the log analysis response
type LogsResponse struct {
	Entries     []LogEntry `json:"entries"`
	TotalLines  int        `json:"total_lines"`
	FilteredBy  string     `json:"filtered_by,omitempty"`
	SearchQuery string     `json:"search_query,omitempty"`
	Message     string     `json:"message,omitempty"`
}

// ExecuteWithContext runs the log analysis with authentication context
func (l *LogsTool) ExecuteWithContext(ctx *MCPToolContext, params json.RawMessage) (interface{}, error) {
	// SECURITY: Only allow master key access
	if !ctx.IsMaster {
		return map[string]interface{}{
			"error":   "access_denied",
			"message": "This tool requires master key authentication",
		}, nil
	}

	// Parse parameters
	var req LogsRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid parameters: %v", err)
		}
	}

	// Set defaults
	if req.Lines == 0 {
		req.Lines = 100
	}
	if req.Lines > 1000 {
		req.Lines = 1000 // Cap at 1000 lines for performance
	}

	// Normalize level filter
	if req.Level != "" {
		req.Level = strings.ToLower(req.Level)
		validLevels := map[string]bool{
			"trace": true, "debug": true, "info": true,
			"warn": true, "warning": true, "error": true,
		}
		if !validLevels[req.Level] {
			return nil, fmt.Errorf("invalid log level: %s (valid: trace, debug, info, warn, error)", req.Level)
		}
	}

	log.Infof("MCP: Logs tool requested by master key (lines=%d, level=%s, search=%s)",
		req.Lines, req.Level, req.Search)

	// Get logs from journalctl (systemd) or docker logs
	entries, err := l.fetchLogs(req.Lines, req.Level, req.Search)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %v", err)
	}

	response := &LogsResponse{
		Entries:    entries,
		TotalLines: len(entries),
	}

	if req.Level != "" {
		response.FilteredBy = fmt.Sprintf("level=%s", req.Level)
	}
	if req.Search != "" {
		response.SearchQuery = req.Search
	}

	if len(entries) == 0 {
		response.Message = "No log entries found matching the criteria"
	}

	return response, nil
}

// fetchLogs retrieves logs using different methods depending on environment
func (l *LogsTool) fetchLogs(lines int, levelFilter, search string) ([]LogEntry, error) {
	var entries []LogEntry

	// Try journalctl first (systemd environment)
	if logs, err := l.fetchFromJournalctl(lines, levelFilter, search); err == nil && len(logs) > 0 {
		return logs, nil
	}

	// Try docker logs if running in container
	if logs, err := l.fetchFromDocker(lines, levelFilter, search); err == nil && len(logs) > 0 {
		return logs, nil
	}

	// Fallback: try to read from common log paths
	return entries, fmt.Errorf("no log source available (tried journalctl, docker logs)")
}

// fetchFromJournalctl retrieves logs from systemd journal
func (l *LogsTool) fetchFromJournalctl(lines int, levelFilter, search string) ([]LogEntry, error) {
	args := []string{
		"-u", "quepasa", // Systemd service name
		"-n", fmt.Sprintf("%d", lines),
		"--no-pager",
		"-o", "json",
	}

	cmd := exec.Command("journalctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return l.parseJournalctlOutput(string(output), levelFilter, search)
}

// fetchFromDocker retrieves logs from Docker container
func (l *LogsTool) fetchFromDocker(lines int, levelFilter, search string) ([]LogEntry, error) {
	args := []string{
		"logs",
		"--tail", fmt.Sprintf("%d", lines),
		"quepasa", // Container name - could be made configurable
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return l.parseDockerOutput(string(output), levelFilter, search)
}

// parseJournalctlOutput parses JSON output from journalctl
func (l *LogsTool) parseJournalctlOutput(output, levelFilter, search string) ([]LogEntry, error) {
	var entries []LogEntry

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var jsonLog map[string]interface{}
		if err := json.Unmarshal([]byte(line), &jsonLog); err != nil {
			continue // Skip malformed lines
		}

		message, _ := jsonLog["MESSAGE"].(string)
		level := l.extractLogLevel(message)

		// Apply level filter
		if levelFilter != "" && !strings.EqualFold(level, levelFilter) {
			continue
		}

		// Apply search filter
		if search != "" && !strings.Contains(strings.ToLower(message), strings.ToLower(search)) {
			continue
		}

		timestamp := ""
		if ts, ok := jsonLog["__REALTIME_TIMESTAMP"].(string); ok {
			if usec, err := time.ParseDuration(ts + "us"); err == nil {
				timestamp = time.Unix(0, int64(usec)).Format(time.RFC3339)
			}
		}

		entries = append(entries, LogEntry{
			Timestamp: timestamp,
			Level:     level,
			Message:   message,
		})
	}

	return entries, nil
}

// parseDockerOutput parses text output from docker logs
func (l *LogsTool) parseDockerOutput(output, levelFilter, search string) ([]LogEntry, error) {
	var entries []LogEntry

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		level := l.extractLogLevel(line)

		// Apply level filter
		if levelFilter != "" && !strings.EqualFold(level, levelFilter) {
			continue
		}

		// Apply search filter
		if search != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(search)) {
			continue
		}

		// Extract timestamp if present (common format: 2024-01-15T10:30:45Z)
		timestamp := ""
		module := ""
		message := line

		// Try to parse timestamp from beginning of line
		if len(line) > 20 {
			if t, err := time.Parse(time.RFC3339, line[:20]); err == nil {
				timestamp = t.Format(time.RFC3339)
				message = strings.TrimSpace(line[20:])
			}
		}

		// Extract module if present (format: [MODULE] or module=X)
		if idx := strings.Index(message, "["); idx >= 0 {
			if endIdx := strings.Index(message[idx:], "]"); endIdx > 0 {
				module = message[idx+1 : idx+endIdx]
				message = strings.TrimSpace(message[idx+endIdx+1:])
			}
		} else if strings.Contains(message, "module=") {
			parts := strings.Fields(message)
			for _, part := range parts {
				if strings.HasPrefix(part, "module=") {
					module = strings.TrimPrefix(part, "module=")
					break
				}
			}
		}

		entries = append(entries, LogEntry{
			Timestamp: timestamp,
			Level:     level,
			Module:    module,
			Message:   message,
		})
	}

	return entries, nil
}

// extractLogLevel attempts to extract log level from a log message
func (l *LogsTool) extractLogLevel(message string) string {
	message = strings.ToLower(message)

	levels := []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
	for _, level := range levels {
		// Look for level=X, [LEVEL], or level: X patterns
		if strings.Contains(message, "level="+level) ||
			strings.Contains(message, "["+level+"]") ||
			strings.Contains(message, level+":") ||
			strings.Contains(message, " "+level+" ") {
			if level == "warning" {
				return "warn"
			}
			return level
		}
	}

	return "info" // Default level
}

// Name returns the tool name
func (l *LogsTool) Name() string {
	return "analyze_logs"
}

// Description returns the tool description
func (l *LogsTool) Description() string {
	return "Analyze QuePasa service logs. Master key access only. Retrieve recent logs, filter by level, search for keywords."
}

// InputSchema returns the JSON schema for the tool input
func (l *LogsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"lines": map[string]interface{}{
				"type":        "integer",
				"description": "Number of log lines to retrieve (default: 100, max: 1000)",
				"minimum":     1,
				"maximum":     1000,
				"default":     100,
			},
			"level": map[string]interface{}{
				"type":        "string",
				"description": "Filter by log level",
				"enum":        []string{"trace", "debug", "info", "warn", "warning", "error"},
			},
			"search": map[string]interface{}{
				"type":        "string",
				"description": "Search for keyword in log messages (case-insensitive)",
			},
			"tail": map[string]interface{}{
				"type":        "boolean",
				"description": "Real-time tail mode (not yet implemented)",
				"default":     false,
			},
		},
		"required": []string{},
	}
}
