package core

import "strings"

type AssistTargetClass struct {
	Label      string
	Normalized string
	Priority   int
	Dangerous  bool
}

type AssistRuleAction string

const (
	AssistRulePrefer AssistRuleAction = "prefer"
	AssistRuleBlock  AssistRuleAction = "block"

	maxAssistRules      = 64
	maxAssistRuleLength = 80
)

type AssistRule struct {
	Enabled bool
	App     string
	Window  string
	Button  string
	Action  AssistRuleAction
	Note    string
}

type AssistRuleContext struct {
	App    string
	Window string
	Button string
}

var dangerousAssistTerms = []string{
	"delete",
	"remove",
	"format",
	"pay",
	"purchase",
	"send",
	"publish",
	"sign",
	"don't save",
	"dont save",
	"don’t save",
	"do not save",
	"discard",
	"discard changes",
	"close without saving",
	"uninstall",
	"overwrite",
	"irreversible",
	"削除",
	"消去",
	"送信",
	"購入",
	"支払",
	"公開",
	"上書き",
	"アンインストール",
	"保存しない",
	"保存せず",
	"破棄",
}

var assistPriorities = []struct {
	terms    []string
	priority int
}{
	{terms: []string{"ok", "はい", "了解"}, priority: 10},
	{terms: []string{"apply", "適用"}, priority: 20},
	{terms: []string{"save", "保存"}, priority: 30},
	{terms: []string{"next", "次へ"}, priority: 40},
	{terms: []string{"yes"}, priority: 50},
	{terms: []string{"cancel", "キャンセル", "閉じる", "close"}, priority: 80},
	{terms: []string{"no", "いいえ"}, priority: 90},
}

func ClassifyAssistTarget(label string) AssistTargetClass {
	normalized := normalizeAssistLabel(label)
	class := AssistTargetClass{
		Label:      label,
		Normalized: normalized,
		Priority:   1000,
	}
	if normalized == "" {
		return class
	}
	for _, term := range dangerousAssistTerms {
		if strings.Contains(normalized, term) {
			class.Dangerous = true
			class.Priority = 10000
			return class
		}
	}
	for _, rule := range assistPriorities {
		for _, term := range rule.terms {
			if normalized == term || strings.Contains(normalized, term) {
				class.Priority = rule.priority
				return class
			}
		}
	}
	return class
}

func ApplyAssistRules(ctx AssistRuleContext, class AssistTargetClass, rules []AssistRule) AssistTargetClass {
	ctx = AssistRuleContext{
		App:    normalizeAssistLabel(ctx.App),
		Window: normalizeAssistLabel(ctx.Window),
		Button: normalizeAssistLabel(ctx.Button),
	}
	for _, rule := range SanitizeAssistRules(rules) {
		if !rule.Enabled {
			continue
		}
		if !assistRuleMatches(ctx, rule) {
			continue
		}
		switch rule.Action {
		case AssistRuleBlock:
			class.Dangerous = true
			class.Priority = 10000
			return class
		case AssistRulePrefer:
			if !class.Dangerous && class.Priority > 1 {
				class.Priority = 1
			}
		}
	}
	return class
}

func SanitizeAssistRules(rules []AssistRule) []AssistRule {
	if len(rules) == 0 {
		return nil
	}
	out := make([]AssistRule, 0, minInt(len(rules), maxAssistRules))
	for _, rule := range rules {
		rule.App = sanitizeAssistRuleText(rule.App)
		rule.Window = sanitizeAssistRuleText(rule.Window)
		rule.Button = sanitizeAssistRuleText(rule.Button)
		rule.Note = sanitizeAssistRuleText(rule.Note)
		if rule.App == "" && rule.Window == "" && rule.Button == "" {
			continue
		}
		if rule.Action != AssistRuleBlock {
			rule.Action = AssistRulePrefer
		}
		out = append(out, rule)
		if len(out) >= maxAssistRules {
			break
		}
	}
	return out
}

func assistRuleMatches(ctx AssistRuleContext, rule AssistRule) bool {
	return assistRuleFieldMatches(ctx.App, rule.App) &&
		assistRuleFieldMatches(ctx.Window, rule.Window) &&
		assistRuleFieldMatches(ctx.Button, rule.Button)
}

func assistRuleFieldMatches(value, pattern string) bool {
	pattern = normalizeAssistLabel(pattern)
	if pattern == "" {
		return true
	}
	return strings.Contains(value, pattern)
}

func sanitizeAssistRuleText(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > maxAssistRuleLength {
		runes = runes[:maxAssistRuleLength]
	}
	return string(runes)
}

func normalizeAssistLabel(label string) string {
	label = strings.ToLower(strings.TrimSpace(label))
	replacer := strings.NewReplacer(
		"&", "",
		"_", "",
		"'", "",
		"’", "",
		"…", "",
		"...", "",
		"　", " ",
	)
	label = replacer.Replace(label)
	return strings.Join(strings.Fields(label), " ")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
