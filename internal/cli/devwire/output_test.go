package devwire

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
)

func TestWriteDEResultText(t *testing.T) {
	cmd := newTestCmd(false)
	out, err := captureWriteDEResult(cmd)
	if err != nil {
		t.Fatalf("WriteDEResult: %v", err)
	}
	if out != "hello" {
		t.Fatalf("expected text output, got %q", out)
	}
}

func TestWriteDEResultJSON(t *testing.T) {
	cmd := newTestCmd(true)
	out, err := captureWriteDEResult(cmd)
	if err != nil {
		t.Fatalf("WriteDEResult: %v", err)
	}
	if out == "" || out == "hello" {
		t.Fatalf("expected JSON envelope, got %q", out)
	}
	for _, part := range []string{`"formatted"`, `"structured"`, `"agent"`, `"completeness"`} {
		if !bytes.Contains([]byte(out), []byte(part)) {
			t.Fatalf("missing %s in %s", part, out)
		}
	}
}

func newTestCmd(useJSON bool) *cobra.Command {
	cmd := &cobra.Command{}
	BindJSONFlag(cmd)
	if useJSON {
		_ = cmd.Flags().Set("json", "true")
	}
	return cmd
}

func captureWriteDEResult(cmd *cobra.Command) (string, error) {
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	answer := &development.DevelopmentAnswer{Question: "sqlite"}
	if err := WriteDEResult(cmd, "hello", answer); err != nil {
		return "", err
	}
	return buf.String(), nil
}
