# Go GUI framework evaluation

## Decision matrix

| Criterion | Wails (webview UI, Go backend) | Fyne (native widgets) | Gio (immediate mode) |
| --- | --- | --- | --- |
| UI paradigm | HTML/CSS/JS in webview | Native widget toolkit | Immediate-mode rendering | 
| Cross-platform (Win/macOS/Linux) | Yes (webview on each OS) | Yes | Yes | 
| Bundling & packaging | App bundle with embedded assets + webview runtime | Go binary with assets (plus platform-specific packaging) | Go binary with assets (plus platform-specific packaging) | 
| Developer experience | Leverage web UI tooling, fast iteration | Go-only UI, simpler stack | Go-only UI, steeper learning curve | 
| Performance & footprint | Webview overhead; acceptable for control UI | Native controls; moderate | High performance; more custom UI work | 
| State management | Frontend framework of choice + Go bindings | Fyne data bindings | Manual state management | 
| Fit for scrcpy-like control UI | Strong (rich layouts, dashboards) | Good for basic controls | Good for custom visuals, more effort | 
| Risk | Webview version differences across OS | Smaller widget set | More code for standard UI | 

**Decision:** **Wails**

## Rationale

Wails offers the best fit for a control/utility UI that benefits from rich layouts and fast iteration. It keeps the Go backend in-process (no RPC server needed) while letting the UI use familiar web tooling for responsive dashboards and device state views. This reduces UI complexity compared with immediate-mode (Gio) and provides more flexibility than the widget set in Fyne.

## Constraints

- **Cross-platform:** Windows/macOS/Linux must be supported.
- **Bundling:** Single installable application per OS that embeds UI assets and the Go backend.
- **Local-only:** UI talks to the Go backend via in-process bindings; no external service required.
- **Hardware access:** Go backend must manage device discovery/pair/connect orchestration directly.

## Target OS

- Windows
- macOS
- Linux

## Architecture (selected: Wails)

### Backend (Go)

Responsibilities:

- Device discovery, pairing, and connection orchestration.
- Process management for scrcpy/adb workflows.
- System integration: USB events, network discovery, logging.

Suggested structure:

- `internal/device` — device models, discovery, pairing/connect logic.
- `internal/adb` — adb wrappers, command execution, error mapping.
- `internal/orchestrator` — high-level workflows (pair → connect → launch).
- `internal/config` — app settings, persisted preferences.
- `internal/logging` — structured logs with UI-friendly events.

Expose backend methods to the UI via Wails bindings:

- `ListDevices()` → device list with status metadata.
- `PairDevice(code)` → returns pairing result and updated state.
- `ConnectDevice(id)` → connect and return session info.
- `StartSession(id, options)` → launches scrcpy session.
- `StopSession(id)` → teardown.

### UI layer (Webview)

Responsibilities:

- Screens: device list, pairing, session controls, settings.
- State bindings to backend via Wails (promises + events).
- UX: status indicators, logs/notifications, connection wizards.

Suggested structure:

- `frontend/` (or Wails default) with a framework like React/Vue/Svelte.
- State store (e.g., Zustand/Pinia) for device state and session lifecycle.
- Event bridge for backend log stream and device updates.

### State model

- **Source of truth:** backend maintains device/session state and emits updates.
- **UI state:** optimistic updates for actions, reconciled with backend events.
- **Errors:** backend returns structured errors mapped to user-friendly messages.

### Packaging

- Build OS-specific bundles with embedded UI assets.
- Ensure webview dependencies are satisfied per OS (Edge WebView2 on Windows, WebKit on macOS, WebKitGTK on Linux).
- Provide a minimal installer per OS with required runtime checks.
