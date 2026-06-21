# MofuMouse Chuchu Mouse Feature Roadmap

Date: 2026-06-21
Status: Active requirements backlog

MofuMouse should exceed the original Chuchu Mouse by preserving the useful cursor-assist idea while adding modern safety, local privacy, keep-awake utility, and degu behavior.

## Core Requirements

| Area | Requirement | Initial behavior | QA evidence |
| --- | --- | --- | --- |
| Button detection | Detect foreground dialog buttons. | Classic Win32 `Button` child scanner is used first when available; x64 UI Automation is the fallback for modern controls and is timeout guarded. x86 uses classic Win32 only. | Win32 settings dialog, independent fixture, Wails control center screenshots, Notepad unsaved-state QA, and Calculator modern-app QA. |
| Target ranking | Rank safe targets. | Default/OK/Apply/Save before Cancel/No. Dangerous labels denied. | Unit tests for labels and ranking. |
| Sniff assist | Degu notices a chosen safe target from the cursor side with a natural sniff/investigate pose. | Implemented for safe classic Win32 and UI Automation button candidates. Overlay only; no physical cursor movement by default. Human-like pointing assets are rejected. | `.codex/qa/overlay-sniff-assist-settings.png` and `.codex/qa/overlay-sniff-assist-uia-control-center.png` visual PASS. Wails/WebView UIA transient misses are covered by same-foreground grace. |
| Optional cursor move | Move cursor to a safe target only when explicitly enabled. | Implemented, off by default, no click. User movement cancels pending return. | `scripts/verify_target_move_qa.ps1` move and cancel checks; `scripts/verify_assist_fixture_qa.ps1` independent Win32 fixture checks. |
| Return to origin | Optionally return to the previous cursor position. | Implemented, off by default. Returns only if the user has not moved the mouse. Dedicated agouti return pose is now used as a short transient action. | `scripts/verify_target_move_qa.ps1 -ExpectReturn`, `scripts/capture_return_pose_qa.ps1`, `.codex/qa/return-agouti-runtime-tiers.png`. |
| Per-app rules | Save app/window/button preferences locally. | Implemented as `TargetRules`: enabled rules can prefer or block a target by app name, window title, and button label. | Core and Win32 assist rule unit tests; Control Center build; fixture rule QA preferring `Later`. |
| Quick scroll | Scroll the active window a small amount from the tray. | Implemented as a fixed safe up/down wheel action with configurable line count. No automatic repeat. | `scripts/verify_quickscroll_qa.ps1` allowlist/delta tests. |
| Quick key | Safe shortcut palette. | Implemented as a fixed allowlist: Find, Copy, Select All, and New Tab. No destructive shortcuts by default. | `scripts/verify_quickkey_qa.ps1` allowlist tests. |
| Browser navigation | Back/forward shortcuts. | Already available from tray. | Manual smoke. |
| Edge warp | Move cursor across screen edges. | Already implemented, off by default. | Display/DPI inventory QA added; physical multi-monitor edge tests still required. |
| Launch at login | Start MofuMouse when the user signs in. | Implemented with the current user's Run key, off by default, exposed in tray and settings. | `scripts/verify_autostart_qa.ps1` isolated-registry QA and settings round-trip tests. |
| Hover activate | Optional activate-window-under-cursor behavior. | Off by default; advanced setting only. | No unexpected focus change in default mode. |
| Keep awake | OS keep-awake and 1px movement are separate modes. | Already implemented, off by default. 1px jiggle stays inside the virtual screen, including negative-origin monitors. | Idle/no-user-input tests, visible tray state, `verify_display_qa.ps1`, unit tests for edge nudge behavior. |
| Sounds | No sound in the current product scope. | Removed from runtime, tray, settings UI, and QA at user request. | `go test ./...` and source grep should show no sound runtime code. |
| Degu actions | Walk, run, return, guard, sniff, groom, nibble, patpat, dig, roll, sleepy, alert. | Many runtime poses exist; walk/run need remake; return exists for agouti only. Human-like point poses are kept only as rejected asset history. | Runtime contact sheet and overlay screenshots. |
| Modern settings | Tabbed control center. | Wails v2 + React + Fluent UI x64 control center implemented; Win32 settings remains fallback. | `.codex/qa/control-center-home.png`, `.codex/qa/control-center-assist.png`. |
| Updates | Check GitHub Releases, select x86/x64 assets, and keep rollback-ready replacement tooling. | Check entry exists, the selected `.exe` / `.zip` asset can be downloaded to the local update cache. The tray can apply a downloaded update after native confirmation by launching a copied helper, waiting for MofuMouse to exit, replacing the target, and restarting it. | `scripts/verify_update_qa.ps1` plus helper copy and temp install/rollback tests. |
| Privacy | Local-only settings; no keystroke content or password capture. | Required invariant. | Code search and tests. |

## Safety Rules

- Never auto-click by default.
- Do not target labels containing high-risk terms: delete, remove, format, pay, purchase, send, publish, sign, uninstall, overwrite, irreversible, or Japanese equivalents such as 削除, 消去, 送信, 購入, 支払, 公開, 上書き.
- Do not automate UAC or secure desktop.
- Do not read password fields or key contents.
- Stop assist movement as soon as the user moves the mouse.
- Keep awake and 1px movement must always be visible in tray/status and easy to turn off.

## Implementation Phases

### Phase A: Asset Foundation

- Generate runtime tiers: 32/48/64/96px.
- Renderer selects nearest larger tier and downscales when possible.
- Remake low-profile walk/run/dig/roll/sleepy poses with ImageGen one by one.
- Add contact-sheet QA for every accepted runtime tier.

### Phase B: Modern Control Center

- Add Wails v2 + React + TypeScript shell. Done.
- Add tabs: Home, Companion, Assist, Keep Awake, Appearance, Updates, Advanced. Done.
- Reuse the same `core.Settings` JSON model. Done.
- Keep current Win32 settings window as fallback until Wails QA passes. Done; `--legacy-settings` forces fallback.
- Add deeper keyboard navigation and save/reload end-to-end QA.

### Phase C: Safe Target Assist

- Implement Win32 child-button scanner for classic dialogs as the first target source. Done.
- Add target label normalization and denylist. Done.
- Add sniff-only assist first. Done for classic Win32 dialogs and Wails/Fluent UI via UI Automation.
- Implement UI Automation foreground-window scanner. Done for x64; x86 uses classic fallback.
- Add optional physical cursor movement only after cancellation tests pass. Done; still needs broader dialog/app QA before release.
- Add per-app/window/button preference rules. Done; rules are local-only and cannot override the built-in dangerous-label denylist.

### Phase D: Productivity Utilities

- Quick scroll. Done for fixed safe tray up/down actions; cursor-side visual handle remains later polish.
- Quick key allowlist. Done for fixed safe shortcuts; user-defined arbitrary keys remain out of scope until stronger safety UI exists.
- Browser navigation improvements.
- Edge warp polish and multi-monitor QA.

### Phase E: Release Readiness

- Full x86/x64 builds.
- Console-free subsystem check.
- x86/x64 DPI awareness and virtual-screen union QA.
- Overlay screenshot QA at 32/48/64px.
- UIA Wails control-center QA and sample dialog QA.
- Update-check mock tests.
- No GitHub Release or Pages publication until every release gate passes.
