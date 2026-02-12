package validator

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		ok    bool
	}{
		{"a@b.co", true},
		{"user@example.com", true},
		{" user@example.com ", true},
		{"", false},
		{"no-at", false},
		{"a@b", false},
		{"@example.com", false},
		{"user@.com", false},
	}
	for _, tt := range tests {
		err := ValidateEmail(tt.email)
		if tt.ok && err != nil {
			t.Fatalf("expected ok for %q, got %v", tt.email, err)
		}
		if !tt.ok && err == nil {
			t.Fatalf("expected error for %q", tt.email)
		}
	}
}

