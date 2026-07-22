package database

import "testing"

func TestQualifyTable(t *testing.T) {
	if got := QualifyTable("", "users"); got != "users" {
		t.Fatalf("expected users, got %q", got)
	}
	if got := QualifyTable("mf", "users"); got != "mf.users" {
		t.Fatalf("expected mf.users, got %q", got)
	}
	if got := QualifyTable("bad-schema!", "users"); got != "users" {
		t.Fatalf("expected unqualified fallback, got %q", got)
	}
}
