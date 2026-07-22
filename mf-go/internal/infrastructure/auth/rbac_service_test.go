package auth

import "testing"

func TestMatchesPermission(t *testing.T) {
	tests := []struct {
		granted  string
		required string
		want     bool
	}{
		{"*", "endpoint:write", true},
		{"org:*", "org:write", true},
		{"org:*", "app:write", false},
		{"*:read", "endpoint:read", true},
		{"*:read", "endpoint:write", false},
		{"endpoint:read", "endpoint:read", true},
		{"endpoint:read", "endpoint:write", false},
	}

	for _, tt := range tests {
		if got := matchesPermission(tt.granted, tt.required); got != tt.want {
			t.Fatalf("matchesPermission(%q, %q) = %v, want %v", tt.granted, tt.required, got, tt.want)
		}
	}
}
