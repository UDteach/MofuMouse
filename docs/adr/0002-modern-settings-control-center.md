# ADR 0002: Modern Settings Control Center

Date: 2026-06-21
Status: Accepted for the next implementation slice

## Context

The current settings window is a plain Win32 form. It is useful for MVP QA, but it does not scale to a "Chuchu Mouse plus degu assistant" product:

- The settings list is already too long for a right-click menu or a single flat window.
- Beginner-facing copy needs short explanations, previews, and safer defaults.
- Future features need categories: companion behavior, target assist, keep-awake, appearance, updates, safety, and QA diagnostics.
- The overlay, cursor tracking, keep-awake, 1px movement, and tray behavior are already Go + Win32 and should remain small and reliable.

Official sources checked:

- Wails v2: https://wails.io/docs/introduction/
- Wails v3: https://v3.wails.io/
- Fyne: https://fyne.io/
- Gio: https://gioui.org/
- Windows App SDK: https://learn.microsoft.com/en-us/windows/apps/windows-app-sdk/
- WinUI 3: https://learn.microsoft.com/en-us/windows/apps/winui/winui3/
- Tauri 2: https://v2.tauri.app/
- Electron: https://electronjs.org/
- Fluent UI: https://developer.microsoft.com/en-us/fluentui
- shadcn/ui: https://ui.shadcn.com/

## Decision

Use a hybrid architecture:

- Keep the runtime assistant in Go + Win32 for the overlay, tray, cursor tracking, power management, 1px jiggle, and x86/x64 builds.
- Replace the current Win32 settings form with a Wails v2 control center using React + TypeScript.
- Prefer Fluent UI React for the first modern control center because MofuMouse is Windows-first and needs accessible, familiar tabs, switches, sliders, dialogs, and status surfaces.
- Keep Wails v3 on the watch list, but do not adopt it yet because its official site marks it as alpha.
- Re-evaluate WinUI 3 later if the project intentionally moves to a native Windows-only stack with .NET/C++ and Windows App SDK tooling.

## Comparison

| Candidate | Fit | Decision |
| --- | --- | --- |
| Go + Win32 only | Small, reliable, already working, but weak for modern tabbed UI. | Keep for runtime; stop expanding settings here. |
| Wails v2 + React + Fluent UI | Reuses Go core, supports modern web UI, good for a settings/control center. | Primary next step. |
| Wails v3 | Better tray/window direction, but official alpha status is too risky for this phase. | Watch list. |
| WinUI 3 / Windows App SDK | Best native Windows/Fluent feel, strong long-term option. Requires a stack migration. | Later v2 candidate. |
| Fyne | Pure Go and simple, but visual style is not the target "modern Windows" feel. | Not primary. |
| Gio | Powerful immediate-mode custom UI, but too expensive for forms/tabs/settings. | Not primary. |
| Tauri 2 | Strong modern shell, but would introduce Rust backend duplication around existing Go services. | Not primary. |
| Electron | Mature web UI, but heavier than needed for a cursor companion. | Avoid for now. |

## Control Center Layout

The first Wails control center should have these tabs:

1. Home: current state, quick toggles, update state, "what is the degu doing now".
2. Companion: cursor-side placement, movement, idle actions, reduced motion, multi-monitor behavior.
3. Assist: button detection, sniff/inspect behavior, target priority, safe denylist, per-app rules.
4. Keep Awake: OS keep-awake, 1px jiggle, schedule, presentation/demo presets.
5. Appearance: name, coat, size, opacity, preview, asset pack status.
6. Updates: current version, GitHub Releases check, x86/x64 asset selection, rollback notes.
7. Advanced: JSON settings path, diagnostics, QA capture, reset, export/import.

## Implementation Notes

- The Win32 settings window remains as a temporary fallback until Wails control center can read/write the same `core.Settings` JSON safely.
- The Wails window should not own the overlay lifecycle. It should call Go methods for settings snapshots, save, update check, and diagnostics.
- The UI should expose beginner labels by default and hide risky/advanced settings under the Advanced tab.
- No release is allowed until the Wails control center, overlay behavior, x86/x64 builds, and visual QA all pass.
