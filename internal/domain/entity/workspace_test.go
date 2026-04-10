package entity

import (
	"strings"
	"testing"
)

func TestNewWorkspace_Success(t *testing.T) {
	ws, err := NewWorkspace("My Workspace", "A description", "owner-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Name != "My Workspace" {
		t.Errorf("expected name My Workspace, got %s", ws.Name)
	}
	if ws.OwnerID != "owner-1" {
		t.Errorf("expected ownerID owner-1, got %s", ws.OwnerID)
	}
	if ws.MemberCount != 1 {
		t.Errorf("expected memberCount 1, got %d", ws.MemberCount)
	}
	if ws.ID.String() == "" {
		t.Error("expected non-empty UUID")
	}
}

func TestNewWorkspace_Validation(t *testing.T) {
	tests := []struct {
		name        string
		wsName      string
		description string
		wantErr     bool
	}{
		{"empty name", "", "desc", true},
		{"name too long", strings.Repeat("a", 101), "desc", true},
		{"description too long", "name", strings.Repeat("a", 501), true},
		{"valid with empty description", "name", "", false},
		{"valid with max name", strings.Repeat("a", 100), "desc", false},
		{"valid with max description", "name", strings.Repeat("a", 500), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWorkspace(tt.wsName, tt.description, "owner-1")
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got err %v", tt.wantErr, err)
			}
		})
	}
}

func TestWorkspace_Update(t *testing.T) {
	ws, _ := NewWorkspace("Old", "Old desc", "owner-1")
	err := ws.Update("New", "New desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Name != "New" || ws.Description != "New desc" {
		t.Error("workspace not updated")
	}
}

func TestWorkspace_Update_Validation(t *testing.T) {
	ws, _ := NewWorkspace("Old", "desc", "owner-1")

	// Empty name
	if err := ws.Update("", "desc"); err == nil {
		t.Error("expected error for empty name")
	}

	// Name too long
	if err := ws.Update(strings.Repeat("a", 101), "desc"); err == nil {
		t.Error("expected error for long name")
	}

	// Should not have mutated on error
	if ws.Name != "Old" {
		t.Error("workspace should not be mutated on validation error")
	}
}

func TestWorkspace_MemberCount(t *testing.T) {
	ws, _ := NewWorkspace("WS", "", "owner-1")
	if ws.MemberCount != 1 {
		t.Fatalf("expected 1, got %d", ws.MemberCount)
	}

	ws.IncrementMemberCount()
	if ws.MemberCount != 2 {
		t.Errorf("expected 2, got %d", ws.MemberCount)
	}

	ws.IncrementMemberCount()
	if ws.MemberCount != 3 {
		t.Errorf("expected 3, got %d", ws.MemberCount)
	}

	// Decrement to 1, should not go below
	ws.DecrementMemberCount()
	ws.DecrementMemberCount()
	if ws.MemberCount != 1 {
		t.Errorf("expected 1 (floor), got %d", ws.MemberCount)
	}

	// Decrement below 1 should stay at 1
	ws.DecrementMemberCount()
	if ws.MemberCount != 1 {
		t.Errorf("expected 1 (floor), got %d", ws.MemberCount)
	}
}
