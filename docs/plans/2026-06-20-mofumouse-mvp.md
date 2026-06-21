# MofuMouse MVP Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Windows MVP where a degu-style assistant follows the cursor, exposes tray controls, supports safe keep-awake / 1px jiggle behavior, and builds for Windows x86/x64.

**Architecture:** Keep pure behavior in `internal/core` and Windows-specific runtime code in `internal/winapp`. The executable in `cmd/mofumouse` wires state, overlay, tray, keep-awake, and smoke mode together. The current drawn mascot is a placeholder; production assets must come from ImageGen.

**Tech Stack:** Go 1.25, Win32 API via `syscall`, `fyne.io/systray` for tray menu.

---

## Task 1: Core State Model

**Files:**
- Create: `internal/core/state.go`
- Create: `internal/core/state_test.go`

**Steps:**

1. Define app settings, runtime state, and feature toggles.
2. Add a helper that returns the active status text.
3. Add tests for default settings and toggle behavior.
4. Run `go test ./internal/core`.

## Task 2: Windows Overlay

**Files:**
- Create: `internal/winapp/overlay_windows.go`

**Steps:**

1. Create a topmost popup window with `WS_EX_LAYERED`, `WS_EX_TRANSPARENT`, `WS_EX_NOACTIVATE`, and `WS_EX_TOOLWINDOW`.
2. Paint a simple degu-like mascot with GDI.
3. Follow the cursor on a timer without stealing focus.
4. Stop cleanly via context cancellation.

## Task 3: Keep Awake And Jiggle

**Files:**
- Create: `internal/winapp/power_windows.go`
- Create: `internal/winapp/jiggle_windows.go`

**Steps:**

1. Wrap `SetThreadExecutionState`.
2. Wrap `GetLastInputInfo`, `GetCursorPos`, and `SetCursorPos`.
3. Move 1px and back only after idle time and unchanged cursor position.
4. Ensure shutdown clears power requests.

## Task 4: Tray And Main Entrypoint

**Files:**
- Create: `cmd/mofumouse/main_windows.go`

**Steps:**

1. Start systray.
2. Add checked menu items for assistant, OS Keep Awake, and 1px Jiggle.
3. Add Quit.
4. Add `--smoke` mode that starts overlay briefly and exits.

## Task 5: Verification

**Commands:**

```powershell
go test ./...
go build -ldflags="-H=windowsgui" -o mofumouse.exe ./cmd/mofumouse
GOARCH=386 go build -ldflags="-H=windowsgui" -o dist\mofumouse-x86.exe ./cmd/mofumouse
GOARCH=amd64 go build -ldflags="-H=windowsgui" -o dist\mofumouse-x64.exe ./cmd/mofumouse
.\mofumouse.exe --smoke
```

**Expected Result:**

- Tests pass.
- Build creates `mofumouse.exe`.
- Smoke mode exits successfully after a short overlay run.
