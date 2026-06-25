package health

import "testing"

func TestFormatDoctorOK(t *testing.T) {
	r := &DoctorResult{
		OK: true,
		Checks: []Check{
			{Name: "workspace", Status: StatusOK, Message: "ok"},
		},
	}
	out := FormatDoctor(r)
	if out == "" {
		t.Fatal("expected formatted output")
	}
}

func TestAppendUnique(t *testing.T) {
	list := appendUnique([]string{"a"}, "a")
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}
	list = appendUnique(list, "b")
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
}
