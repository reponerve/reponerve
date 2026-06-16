package sqlite

import (
	"testing"
	"time"
)

func TestParseDateTime_StringRFC3339(t *testing.T) {
	tm, err := ParseDateTime("2024-01-15T10:30:00Z")
	if err != nil {
		t.Fatalf("ParseDateTime failed: %v", err)
	}
	if tm.UTC().Format(time.RFC3339) != "2024-01-15T10:30:00Z" {
		t.Fatalf("unexpected time: %v", tm)
	}
}

func TestParseDateTime_SQLiteDatetime(t *testing.T) {
	tm, err := ParseDateTime("2024-01-15 10:30:00")
	if err != nil {
		t.Fatalf("ParseDateTime failed: %v", err)
	}
	if tm.Year() != 2024 || tm.Month() != time.January || tm.Day() != 15 {
		t.Fatalf("unexpected time: %v", tm)
	}
}

func TestParseDateTime_DuplicateTimezone(t *testing.T) {
	tm, err := ParseDateTime("2026-05-07 12:40:59 -0400 -0400")
	if err != nil {
		t.Fatalf("ParseDateTime failed: %v", err)
	}
	if tm.Year() != 2026 || tm.Month() != time.May || tm.Day() != 7 {
		t.Fatalf("unexpected time: %v", tm)
	}
}

func TestParseDateTime_GitLogLayout(t *testing.T) {
	tm, err := ParseDateTime("2026-05-07 12:40:59 -0400")
	if err != nil {
		t.Fatalf("ParseDateTime failed: %v", err)
	}
	if tm.IsZero() {
		t.Fatal("expected non-zero time")
	}
}

func TestFlexibleTime_Scan(t *testing.T) {
	var ft FlexibleTime
	if err := ft.Scan("2024-06-01T12:00:00Z"); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if ft.Time.IsZero() {
		t.Fatal("expected non-zero time")
	}
}
