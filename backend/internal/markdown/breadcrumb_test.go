package markdown

import "testing"

func TestHeadingStack_UpdateAndClone(t *testing.T) {
	h := NewHeadingStack()
	h.Update(1, "Root")
	h.Update(2, "Child")
	h.Update(4, "Deep")
	if got := h.Breadcrumb(); got != "Root > Child > Deep" {
		t.Fatalf("unexpected breadcrumb: %q", got)
	}

	clone := h.Clone()
	h.Update(1, "Reset")
	if got := clone.Breadcrumb(); got != "Root > Child > Deep" {
		t.Fatalf("clone should be isolated, got %q", got)
	}
	if got := h.Breadcrumb(); got != "Reset" {
		t.Fatalf("expected h1 reset, got %q", got)
	}
}
