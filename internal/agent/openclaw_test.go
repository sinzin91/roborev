package agent

import (
	"testing"
)

func TestOpenClawAgentName(t *testing.T) {
	a := NewOpenClawAgent("")
	if a.Name() != "openclaw" {
		t.Errorf("expected name 'openclaw', got '%s'", a.Name())
	}
}

func TestOpenClawAgentCommandName(t *testing.T) {
	a := NewOpenClawAgent("")
	if a.CommandName() != "openclaw" {
		t.Errorf("expected command 'openclaw', got '%s'", a.CommandName())
	}

	a2 := NewOpenClawAgent("custom-openclaw")
	if a2.CommandName() != "custom-openclaw" {
		t.Errorf("expected command 'custom-openclaw', got '%s'", a2.CommandName())
	}
}

func TestOpenClawAgentWithAgentic(t *testing.T) {
	a := NewOpenClawAgent("")
	a2 := a.WithAgentic()

	// Should return a copy, not the same pointer
	if a == a2 {
		t.Error("WithAgentic should return a copy, not the same instance")
	}

	// Copy should have same values
	if a.Command != a2.Command {
		t.Errorf("copy should have same command: got %s vs %s", a.Command, a2.Command)
	}
}

func TestOpenClawAgentRegistered(t *testing.T) {
	a, err := Get("openclaw")
	if err != nil {
		t.Fatalf("openclaw should be registered: %v", err)
	}
	if a.Name() != "openclaw" {
		t.Errorf("expected name 'openclaw', got '%s'", a.Name())
	}
}
