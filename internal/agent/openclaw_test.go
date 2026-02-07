package agent

import (
	"bytes"
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
		t.Errorf("agentCopy should have same command: got %s vs %s", a.Command, a2.Command)
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

func TestLimitedWriter(t *testing.T) {
	var buf bytes.Buffer
	lw := &limitedWriter{w: &buf, n: 10}

	// Write within limit
	n, err := lw.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Errorf("expected 5 bytes reported, got %d, err: %v", n, err)
	}

	// Write exceeding limit - should truncate but report full consumption
	n, err = lw.Write([]byte("worldworld"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Should report full consumption (10) to avoid io.Copy ErrShortWrite
	if n != 10 {
		t.Errorf("expected 10 bytes reported (full consumption), got %d", n)
	}
	// Only 5 more bytes should actually be written
	if buf.String() != "helloworld" {
		t.Errorf("expected 'helloworld', got '%s'", buf.String())
	}

	// Further writes should be discarded but report success
	n, err = lw.Write([]byte("extra"))
	if err != nil || n != 5 {
		t.Errorf("expected 5 (discarded but reported), got %d, err: %v", n, err)
	}
	if buf.String() != "helloworld" {
		t.Errorf("buffer should not grow: got '%s'", buf.String())
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc..."},
		// UTF-8: should not split multibyte chars
		{"héllo", 3, "hél..."},
		{"日本語テスト", 3, "日本語..."},
	}

	for _, tc := range cases {
		got := truncate(tc.input, tc.maxLen)
		if got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.maxLen, got, tc.want)
		}
	}
}
