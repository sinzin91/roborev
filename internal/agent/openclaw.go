package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

// maxOutputSize limits captured stdout/stderr to prevent memory exhaustion
const maxOutputSize = 1 << 20 // 1MB

// OpenClawAgent runs code reviews using the OpenClaw gateway
type OpenClawAgent struct {
	Command string // The openclaw command to run (default: "openclaw")
}

// NewOpenClawAgent creates a new OpenClaw agent
func NewOpenClawAgent(command string) *OpenClawAgent {
	if command == "" {
		command = "openclaw"
	}
	return &OpenClawAgent{Command: command}
}

func (a *OpenClawAgent) Name() string {
	return "openclaw"
}

func (a *OpenClawAgent) CommandName() string {
	return a.Command
}

// WithAgentic returns a copy configured for agentic (tool-use) mode.
// OpenClaw agents are always agentic, so this just returns a copy.
func (a *OpenClawAgent) WithAgentic() *OpenClawAgent {
	agentCopy := *a
	return &agentCopy
}

// Review runs a code review using the OpenClaw CLI.
// commitSHA is part of the Agent interface but unused here; OpenClaw reviews
// the current working directory state, not a specific commit.
func (a *OpenClawAgent) Review(ctx context.Context, repoPath, commitSHA, prompt string) (string, error) {
	// OpenClaw CLI: openclaw run "<prompt>"
	// Runs a one-shot agent turn with the given prompt
	_ = commitSHA // unused: OpenClaw reviews current state, not specific commits
	args := []string{"run", prompt}

	cmd := exec.CommandContext(ctx, a.Command, args...)
	cmd.Dir = repoPath

	// Use limited writers to prevent memory exhaustion from large outputs
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &limitedWriter{w: &stdoutBuf, n: maxOutputSize}
	cmd.Stderr = &limitedWriter{w: &stderrBuf, n: maxOutputSize}

	if err := cmd.Run(); err != nil {
		// Truncate output in errors to avoid leaking large amounts of data
		return "", fmt.Errorf(
			"openclaw failed: %w\nstdout: %s\nstderr: %s",
			err,
			truncate(stdoutBuf.String(), 500),
			truncate(stderrBuf.String(), 500),
		)
	}

	result := stdoutBuf.String()
	if len(result) == 0 {
		return "No review output generated", nil
	}

	// Try to parse as JSON first (openclaw may return structured response)
	var resp openclawResponse
	if err := json.Unmarshal([]byte(result), &resp); err == nil {
		if resp.Error != "" {
			return "", fmt.Errorf("openclaw error: %s", resp.Error)
		}
		if resp.Reply != "" {
			return resp.Reply, nil
		}
	}

	// If not JSON or no reply field, return raw output
	return result, nil
}

// openclawResponse represents the JSON response from openclaw run
type openclawResponse struct {
	Reply string `json:"reply,omitempty"`
	Error string `json:"error,omitempty"`
}

// limitedWriter wraps a writer and stops writing after n bytes.
// It always reports full consumption to avoid io.Copy short-write errors.
type limitedWriter struct {
	w io.Writer
	n int
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	if l.n <= 0 {
		return len(p), nil // discard but report full consumption
	}
	writeLen := len(p)
	if writeLen > l.n {
		writeLen = l.n
	}
	n, err := l.w.Write(p[:writeLen])
	l.n -= n
	if err != nil {
		return n, err
	}
	// Always report full consumption to avoid io.Copy ErrShortWrite
	return len(p), nil
}

// truncate returns s truncated to approximately maxLen runes with an ellipsis if needed.
// Uses rune-aware truncation to avoid splitting UTF-8 characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Convert to runes to avoid splitting multibyte characters
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func init() {
	Register(NewOpenClawAgent(""))
}
