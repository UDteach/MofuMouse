# ADR-0001: Use Go + Win32 API For The First MVP

## Status

Accepted for MVP

## Context

MofuMouse needs Windows-specific behavior:

- a transparent, topmost, click-through cursor-side mascot overlay
- system tray controls
- mouse movement and idle detection
- keep-awake power requests
- future UI Automation for button detection

The best long-term fit for deep Windows UI Automation and polished desktop UI is C# with WPF or WinUI. Current local tooling does not include `dotnet`, while Go is already installed.

## Decision

Build the first MVP in Go using direct Win32 API calls and a small system tray library. Keep the architecture small and explicit so it can either mature as a Go app or be ported to C# later.

The release target is both Windows x86 and x64. Win32 interop code must avoid fixed-width pointer assumptions and use `uintptr` for handles, pointers, and `LONG_PTR`-style values.

## Consequences

### Positive

- No new SDK installation is required for the first working build.
- Go can produce a small native executable.
- Go can cross-build Windows x86 and x64 from the same source.
- Direct Win32 calls make overlay, input, power, and idle behavior explicit.
- The MVP can validate the product feel before committing to a heavier stack.

### Negative

- Go has weaker first-class support for Windows UI Automation than C#.
- Advanced UI and animation will take more manual Win32/GDI work.
- Polished settings UI may be slower to build than WPF/WinUI.

### Neutral

- C# / WPF remains the preferred candidate if target detection and settings UI become the main complexity.
- The MVP should keep domain logic separate from Win32 calls to preserve a migration path.

## Alternatives Considered

**C# / WPF**

Best Windows API and UI Automation fit, but local SDK is missing. Keep as long-term candidate.

**C# / WinUI 3**

Modern Windows UI, but heavier setup and less necessary for a cursor-side overlay MVP.

**Go + Wails**

Good for settings windows, but transparent click-through overlay and system-level input still require platform-specific Win32 work.

**Electron / Tauri**

Good visual layer, but heavier or more complex than needed for this Windows-first utility.

**Rust**

Excellent Windows API access through `windows-rs`, but slower iteration for this project unless memory safety around native calls becomes a primary concern.

## References

- Microsoft Learn: WPF documentation
- Microsoft Learn: WinUI 3 documentation
- Microsoft Learn: SendInput
- Microsoft Learn: SetThreadExecutionState
- Wails documentation
- Fyne systray documentation
