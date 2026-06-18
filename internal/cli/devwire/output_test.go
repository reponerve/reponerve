package devwire

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
)

func TestWriteDEResultProse(t *testing.T) {
	cmd := newTestCmd(FormatProse, false, 0)
	out, err := captureWriteDEResult(cmd)
	if err != nil {
		t.Fatalf("WriteDEResult: %v", err)
	}
	if out != "hello\n" {
		t.Fatalf("expected prose output, got %q", out)
	}
}

func TestWriteDEResultJSON(t *testing.T) {
	cmd := newTestCmd(FormatJSON, true, 0)
	out, err := captureWriteDEResult(cmd)
	if err != nil {
		t.Fatalf("WriteDEResult: %v", err)
	}
	for _, part := range []string{`"formatted"`, `"structured"`, `"agent"`, `"completeness"`} {
		if !strings.Contains(out, part) {
			t.Fatalf("missing %s in %s", part, out)
		}
	}
}

func TestWriteDEResultCompact(t *testing.T) {
	cmd := newTestCmd(FormatCompact, false, 0)
	cmd.SetOut(&bytes.Buffer{})
	formatted := "ENTITY BRIEFINGS\n  foo [bar]\n"
	if err := WriteDEResult(cmd, formatted, &development.DevelopmentAnswer{Question: "q"}); err != nil {
		t.Fatal(err)
	}
	out := cmd.OutOrStdout().(*bytes.Buffer).String()
	if strings.Contains(out, "ENTITY BRIEFINGS") {
		t.Fatalf("expected compact header, got %q", out)
	}
}

func TestResolveFormatJSONFlag(t *testing.T) {
	cmd := &cobra.Command{}
	BindOutputFlags(cmd)
	_ = cmd.Flags().Set("json", "true")
	format, err := ResolveFormat(cmd)
	if err != nil || format != FormatJSON {
		t.Fatalf("json flag: format=%q err=%v", format, err)
	}
}

func newTestCmd(format string, useJSON bool, budget int) *cobra.Command {
	cmd := &cobra.Command{}
	BindOutputFlags(cmd)
	if useJSON {
		_ = cmd.Flags().Set("json", "true")
	} else {
		_ = cmd.Flags().Set("format", format)
	}
	if budget > 0 {
		_ = cmd.Flags().Set("token-budget", "10")
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
