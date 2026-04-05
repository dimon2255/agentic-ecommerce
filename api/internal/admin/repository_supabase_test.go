package admin

import "testing"

func TestValidSort_AllowedColumn(t *testing.T) {
	allowed := map[string]bool{"created_at": true, "email": true, "status": true}

	got := validSort("email", allowed, "created_at")
	if got != "email" {
		t.Errorf("expected email, got %s", got)
	}
}

func TestValidSort_MaliciousInput(t *testing.T) {
	allowed := map[string]bool{"created_at": true, "email": true, "status": true}

	tests := []struct {
		name  string
		input string
	}{
		{"sql injection attempt", "id; DROP TABLE orders--"},
		{"unknown column", "secret_column"},
		{"empty string", ""},
		{"column with spaces", "created at"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validSort(tt.input, allowed, "created_at")
			if got != "created_at" {
				t.Errorf("expected fallback created_at, got %s", got)
			}
		})
	}
}

func TestValidSortDir(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"asc", "asc"},
		{"desc", "desc"},
		{"ASC", "desc"},
		{"", "desc"},
		{"DROP TABLE", "desc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := validSortDir(tt.input)
			if got != tt.want {
				t.Errorf("validSortDir(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidDateParam(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2026-01-15", "2026-01-15"},
		{"2026-12-31", "2026-12-31"},
		{"not-a-date", ""},
		{"2026/01/15", ""},
		{"", ""},
		{"2026-01-15; DROP TABLE", ""},
		{"2026-13-01", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := validDateParam(tt.input)
			if got != tt.want {
				t.Errorf("validDateParam(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
