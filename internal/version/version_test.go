package version

import "testing"

func TestStringDev(t *testing.T) {
	Version = "dev"
	Commit = "unknown"
	Date = "unknown"
	if got := String(); got != "dev" {
		t.Fatalf("got %q", got)
	}
}

func TestStringRelease(t *testing.T) {
	Version = "v1.5.0"
	Commit = "abc1234"
	Date = "2026-06-24"
	got := String()
	if got == "" || got == "v1.5.0" {
		t.Fatalf("expected full string, got %q", got)
	}
}
