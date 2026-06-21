//go:build windows

package winapp

import (
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"unsafe"

	"mofumouse/internal/core"
)

type AssistTarget struct {
	HWND   uintptr
	Label  string
	App    string
	Window string
	Class  core.AssistTargetClass
	Rect   rect
	Point  point
}

type assistScanState struct {
	targets []AssistTarget
}

var enumAssistButtonsCallback = makeCallback(enumAssistButtons)

func ForegroundAssistTargets() []AssistTarget {
	return ForegroundAssistTargetsWithRules(nil)
}

func ForegroundAssistTargetsWithRules(rules []core.AssistRule) []AssistTarget {
	foreground, _, _ := procGetForegroundWindow.Call()
	if foreground == 0 {
		return nil
	}
	ctx := foregroundAssistContext(foreground)
	targets := classicForegroundAssistTargets(foreground)
	if len(targets) == 0 {
		targets = uiaForegroundAssistTargets(foreground)
	}
	return applyAssistRulesToTargets(targets, ctx, rules)
}

func classicForegroundAssistTargets(foreground uintptr) []AssistTarget {
	state := &assistScanState{}
	procEnumChildWindows.Call(foreground, enumAssistButtonsCallback, uintptr(unsafe.Pointer(state)))
	sortAssistTargets(state.targets)
	return state.targets
}

func SafeForegroundAssistTargets() []AssistTarget {
	return SafeForegroundAssistTargetsWithRules(nil)
}

func SafeForegroundAssistTargetsWithRules(rules []core.AssistRule) []AssistTarget {
	targets := ForegroundAssistTargetsWithRules(rules)
	safe := make([]AssistTarget, 0, len(targets))
	for _, target := range targets {
		if !target.Class.Dangerous && target.Class.Priority < 1000 {
			safe = append(safe, target)
		}
	}
	return safe
}

func foregroundAssistContext(hwnd uintptr) core.AssistRuleContext {
	return core.AssistRuleContext{
		App:    foregroundProcessName(hwnd),
		Window: strings.TrimSpace(windowText(hwnd)),
	}
}

func foregroundProcessName(hwnd uintptr) string {
	var pid uint32
	procGetWindowThreadProcID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return ""
	}
	handle, _, _ := procOpenProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if handle == 0 {
		return ""
	}
	defer procCloseHandle.Call(handle)

	buf := make([]uint16, 32768)
	size := uint32(len(buf))
	ret, _, _ := procQueryFullProcessImage.Call(
		handle,
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 || size == 0 {
		return ""
	}
	return filepath.Base(syscall.UTF16ToString(buf[:size]))
}

func applyAssistRulesToTargets(targets []AssistTarget, ctx core.AssistRuleContext, rules []core.AssistRule) []AssistTarget {
	if len(targets) == 0 {
		return nil
	}
	for i := range targets {
		targets[i].App = ctx.App
		targets[i].Window = ctx.Window
		targetCtx := ctx
		targetCtx.Button = targets[i].Label
		targets[i].Class = core.ApplyAssistRules(targetCtx, targets[i].Class, rules)
	}
	sortAssistTargets(targets)
	return targets
}

func enumAssistButtons(hwnd uintptr, lParam uintptr) uintptr {
	if hwnd == 0 || lParam == 0 {
		return 1
	}
	if ret, _, _ := procIsWindowVisible.Call(hwnd); ret == 0 {
		return 1
	}
	if ret, _, _ := procIsWindowEnabled.Call(hwnd); ret == 0 {
		return 1
	}
	if !strings.EqualFold(windowClassName(hwnd), "Button") {
		return 1
	}

	label := strings.TrimSpace(windowText(hwnd))
	if label == "" {
		return 1
	}

	var r rect
	if ret, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r))); ret == 0 {
		return 1
	}
	if r.Right <= r.Left || r.Bottom <= r.Top {
		return 1
	}

	state := (*assistScanState)(unsafe.Pointer(lParam))
	state.targets = append(state.targets, AssistTarget{
		HWND:  hwnd,
		Label: label,
		Class: core.ClassifyAssistTarget(label),
		Rect:  r,
		Point: point{
			X: r.Left + (r.Right-r.Left)/2,
			Y: r.Top + (r.Bottom-r.Top)/2,
		},
	})
	return 1
}

func mergeAssistTargets(primary []AssistTarget, fallback []AssistTarget) []AssistTarget {
	if len(primary) == 0 {
		targets := append([]AssistTarget(nil), fallback...)
		sortAssistTargets(targets)
		return targets
	}
	targets := append([]AssistTarget(nil), primary...)
	for _, target := range fallback {
		if !hasSimilarAssistTarget(targets, target) {
			targets = append(targets, target)
		}
	}
	sortAssistTargets(targets)
	return targets
}

func hasSimilarAssistTarget(targets []AssistTarget, candidate AssistTarget) bool {
	for _, target := range targets {
		if target.HWND != 0 && candidate.HWND != 0 && target.HWND == candidate.HWND {
			return true
		}
		if strings.EqualFold(strings.TrimSpace(target.Label), strings.TrimSpace(candidate.Label)) &&
			abs32(target.Rect.Left-candidate.Rect.Left) <= 3 &&
			abs32(target.Rect.Top-candidate.Rect.Top) <= 3 &&
			abs32(target.Rect.Right-candidate.Rect.Right) <= 3 &&
			abs32(target.Rect.Bottom-candidate.Rect.Bottom) <= 3 {
			return true
		}
		if abs32(target.Point.X-candidate.Point.X) <= 3 &&
			abs32(target.Point.Y-candidate.Point.Y) <= 3 {
			return true
		}
	}
	return false
}

func sortAssistTargets(targets []AssistTarget) {
	sort.SliceStable(targets, func(i, j int) bool {
		if targets[i].Class.Priority != targets[j].Class.Priority {
			return targets[i].Class.Priority < targets[j].Class.Priority
		}
		if targets[i].Rect.Top != targets[j].Rect.Top {
			return targets[i].Rect.Top < targets[j].Rect.Top
		}
		if targets[i].Rect.Left != targets[j].Rect.Left {
			return targets[i].Rect.Left < targets[j].Rect.Left
		}
		return strings.ToLower(targets[i].Label) < strings.ToLower(targets[j].Label)
	})
}

func windowClassName(hwnd uintptr) string {
	buf := make([]uint16, 128)
	ret, _, _ := procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:ret])
}
