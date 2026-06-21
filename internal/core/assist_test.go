package core

import "testing"

func TestClassifyAssistTargetPriorities(t *testing.T) {
	tests := []struct {
		label    string
		priority int
	}{
		{label: "OK", priority: 10},
		{label: "はい", priority: 10},
		{label: "Apply", priority: 20},
		{label: "適用", priority: 20},
		{label: "Save", priority: 30},
		{label: "保存", priority: 30},
		{label: "Next >", priority: 40},
		{label: "キャンセル", priority: 80},
		{label: "いいえ", priority: 90},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			got := ClassifyAssistTarget(tt.label)
			if got.Dangerous {
				t.Fatalf("Dangerous = true, want false")
			}
			if got.Priority != tt.priority {
				t.Fatalf("Priority = %d, want %d", got.Priority, tt.priority)
			}
		})
	}
}

func TestClassifyAssistTargetDangerousLabels(t *testing.T) {
	labels := []string{
		"Delete file",
		"Remove account",
		"Format disk",
		"Pay now",
		"Send message",
		"Publish",
		"Don't Save",
		"Don’t Save",
		"Discard changes",
		"削除",
		"送信する",
		"購入",
		"上書き保存",
		"保存しない",
		"破棄",
	}
	for _, label := range labels {
		t.Run(label, func(t *testing.T) {
			got := ClassifyAssistTarget(label)
			if !got.Dangerous {
				t.Fatalf("Dangerous = false, want true")
			}
			if got.Priority != 10000 {
				t.Fatalf("Priority = %d, want 10000", got.Priority)
			}
		})
	}
}

func TestApplyAssistRulesPreferMatchingContext(t *testing.T) {
	class := ClassifyAssistTarget("Later")
	got := ApplyAssistRules(AssistRuleContext{
		App:    "sample.exe",
		Window: "Example dialog",
		Button: "Later",
	}, class, []AssistRule{
		{Enabled: true, App: "sample", Window: "dialog", Button: "later", Action: AssistRulePrefer},
	})

	if got.Dangerous {
		t.Fatal("prefer rule should not mark the target dangerous")
	}
	if got.Priority != 1 {
		t.Fatalf("priority = %d, want 1", got.Priority)
	}
}

func TestApplyAssistRulesBlockMatchingContext(t *testing.T) {
	class := ClassifyAssistTarget("OK")
	got := ApplyAssistRules(AssistRuleContext{
		App:    "sample.exe",
		Window: "Example dialog",
		Button: "OK",
	}, class, []AssistRule{
		{Enabled: true, App: "sample", Button: "ok", Action: AssistRuleBlock},
	})

	if !got.Dangerous {
		t.Fatal("block rule should mark the target dangerous")
	}
	if got.Priority != 10000 {
		t.Fatalf("priority = %d, want 10000", got.Priority)
	}
}

func TestApplyAssistRulesIgnoresDisabledRule(t *testing.T) {
	class := ClassifyAssistTarget("OK")
	got := ApplyAssistRules(AssistRuleContext{Button: "OK"}, class, []AssistRule{
		{Enabled: false, Button: "ok", Action: AssistRuleBlock},
	})

	if got.Dangerous {
		t.Fatal("disabled block rule should be ignored")
	}
	if got.Priority != 10 {
		t.Fatalf("priority = %d, want 10", got.Priority)
	}
}

func TestApplyAssistRulesDoesNotPreferDangerousTarget(t *testing.T) {
	class := ClassifyAssistTarget("Delete file")
	got := ApplyAssistRules(AssistRuleContext{Button: "Delete file"}, class, []AssistRule{
		{Enabled: true, Button: "delete", Action: AssistRulePrefer},
	})

	if !got.Dangerous {
		t.Fatal("dangerous target should remain dangerous")
	}
	if got.Priority != 10000 {
		t.Fatalf("priority = %d, want 10000", got.Priority)
	}
}

func TestSanitizeAssistRules(t *testing.T) {
	long := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"
	got := SanitizeAssistRules([]AssistRule{
		{Enabled: true, Action: AssistRuleBlock},
		{Enabled: true, Button: "  Save  ", Action: "unknown", Note: long},
	})

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Button != "Save" {
		t.Fatalf("button = %q, want Save", got[0].Button)
	}
	if got[0].Action != AssistRulePrefer {
		t.Fatalf("action = %q, want prefer", got[0].Action)
	}
	if len([]rune(got[0].Note)) != maxAssistRuleLength {
		t.Fatalf("note length = %d, want %d", len([]rune(got[0].Note)), maxAssistRuleLength)
	}
}
