package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

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
	copy := *a
	return &copy
}

func (a *OpenClawAgent) Review(ctx context.Context, repoPath, commitSHA, prompt string) (string, error) {
	// OpenClaw CLI: openclaw run "<prompt>"
	// Runs a one-shot agent turn with the given prompt
	args := []string{"run", prompt}

	cmd := exec.CommandContext(ctx, a.Command, args...)
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(
			"openclaw failed: %w\nstdout: %s\nstderr: %s",
			err,
			stdout.String(),
			stderr.String(),
		)
	}

	result := stdout.String()
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

func init() {
	Register(NewOpenClawAgent(""))
}
