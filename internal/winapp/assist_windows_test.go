//go:build windows

package winapp

import (
	"testing"

	"mofumouse/internal/core"
)

func TestMergeAssistTargetsDeduplicatesSameButton(t *testing.T) {
	primary := []AssistTarget{
		testAssistTarget("OK", 10, rect{Left: 100, Top: 200, Right: 160, Bottom: 230}),
	}
	fallback := []AssistTarget{
		testAssistTarget("OK", 10, rect{Left: 101, Top: 201, Right: 161, Bottom: 231}),
	}

	got := mergeAssistTargets(primary, fallback)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Label != "OK" {
		t.Fatalf("Label = %q, want OK", got[0].Label)
	}
}

func TestMergeAssistTargetsSortsByPriorityThenPosition(t *testing.T) {
	targets := mergeAssistTargets(nil, []AssistTarget{
		testAssistTarget("Cancel", 80, rect{Left: 20, Top: 20, Right: 80, Bottom: 50}),
		testAssistTarget("Save", 30, rect{Left: 300, Top: 80, Right: 380, Bottom: 110}),
		testAssistTarget("Apply", 20, rect{Left: 200, Top: 80, Right: 280, Bottom: 110}),
		testAssistTarget("OK", 10, rect{Left: 100, Top: 200, Right: 160, Bottom: 230}),
		testAssistTarget("OK", 10, rect{Left: 60, Top: 100, Right: 120, Bottom: 130}),
	})

	gotLabels := make([]string, 0, len(targets))
	for _, target := range targets {
		gotLabels = append(gotLabels, target.Label)
	}
	want := []string{"OK", "OK", "Apply", "Save", "Cancel"}
	if len(gotLabels) != len(want) {
		t.Fatalf("labels = %v, want %v", gotLabels, want)
	}
	for i := range want {
		if gotLabels[i] != want[i] {
			t.Fatalf("labels = %v, want %v", gotLabels, want)
		}
	}
	if targets[0].Rect.Top != 100 || targets[1].Rect.Top != 200 {
		t.Fatalf("same priority sort by position failed: %+v", targets[:2])
	}
}

func TestApplyAssistRulesToTargets(t *testing.T) {
	targets := []AssistTarget{
		testAssistTarget("Save", 30, rect{Left: 100, Top: 100, Right: 180, Bottom: 130}),
		testAssistTarget("Later", 1000, rect{Left: 200, Top: 100, Right: 280, Bottom: 130}),
	}
	got := applyAssistRulesToTargets(targets, core.AssistRuleContext{
		App:    "sample.exe",
		Window: "Example dialog",
	}, []core.AssistRule{
		{Enabled: true, App: "sample", Button: "later", Action: core.AssistRulePrefer},
		{Enabled: true, App: "sample", Button: "save", Action: core.AssistRuleBlock},
	})

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Label != "Later" {
		t.Fatalf("first label = %q, want Later", got[0].Label)
	}
	if got[0].App != "sample.exe" || got[0].Window != "Example dialog" {
		t.Fatalf("context not copied to target: %+v", got[0])
	}
	if got[1].Label != "Save" || !got[1].Class.Dangerous {
		t.Fatalf("blocked target not marked dangerous: %+v", got[1])
	}
}

func testAssistTarget(label string, priority int, r rect) AssistTarget {
	return AssistTarget{
		HWND:  0,
		Label: label,
		Class: core.AssistTargetClass{Label: label, Normalized: label, Priority: priority},
		Rect:  r,
		Point: point{
			X: r.Left + (r.Right-r.Left)/2,
			Y: r.Top + (r.Bottom-r.Top)/2,
		},
	}
}
