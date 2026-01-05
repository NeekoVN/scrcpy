# GUI Flow Spec

## Overview
This document outlines primary GUI flows for device pairing, connecting, mirroring controls, and session persistence. It also defines screens/states, acceptance criteria, and wireframe-level widget/error lists.

## Screens / States
- **No device**: no USB device detected; no known devices connected.
- **Device connected (USB) not paired**: USB detected; wireless pairing not completed.
- **Pairing in progress**: pairing request sent; awaiting confirmation.
- **Pairing failed**: pairing rejected/timeout/error.
- **Paired device available**: device paired and ready to connect.
- **Connecting (USB)**: initiating USB connection.
- **Connecting (TCP/IP)**: initiating TCP/IP connection.
- **Connecting (known device)**: initiating connection from recent list.
- **Connection failed**: connect error (auth, unreachable, timeout).
- **Mirroring active**: stream running; controls available.
- **Mirroring paused/stopped**: session ended; device may remain connected.

## Flow: Pairing (USB + Wireless, Android 11+ `adb pair`)
### Summary
Allow pairing over Wi‑Fi using an ADB pairing code after initial USB detection or via direct pairing UI.

### Steps
1. **No device** ➜ user plugs USB device.
2. **Device connected (USB) not paired** ➜ user opens Pairing panel.
3. User selects **Pair via ADB (Android 11+)**.
4. User enters **IP:Port** and **pairing code** (or scans QR if supported).
5. System runs `adb pair` and displays result.
6. On success, device appears in **Paired device available** and **Recent devices**.

### Acceptance Criteria
- Pairing UI validates IP:Port and pairing code format before submit.
- Pairing shows progress and prevents duplicate submission while in progress.
- On success, device is stored in recent devices with timestamp.
- On failure, error is displayed with actionable hint (e.g., “Check pairing code”).
- Pairing flow accessible both from **No device** and **Device connected (USB) not paired** states.

### Wireframe: Key Widgets
- Pairing panel title, explanation text.
- Input fields: **IP:Port**, **Pairing Code**.
- Actions: **Pair**, **Cancel**.
- Status area: progress spinner, success banner.

### Error States
- Invalid IP/port format.
- Invalid/expired pairing code.
- Device not on same network.
- `adb pair` timeout or auth failure.

## Flow: Connecting (USB, TCP/IP, Known Devices)
### Summary
Connect to devices via USB, TCP/IP, or a recent device list.

### Steps
1. **USB**: Detect device ➜ user clicks **Connect USB** ➜ **Connecting (USB)** ➜ **Mirroring active**.
2. **TCP/IP**: User enters IP:Port ➜ **Connecting (TCP/IP)** ➜ **Mirroring active**.
3. **Known device**: User selects from **Recent devices** ➜ **Connecting (known device)** ➜ **Mirroring active**.

### Acceptance Criteria
- USB connection is one‑click when device is detected.
- TCP/IP connection validates IP:Port before connect.
- Recent devices list shows name, IP, last used time.
- Connection failure displays reason and retry option.
- Successful connection updates recent devices and last‑used options.

### Wireframe: Key Widgets
- Primary action buttons: **Connect USB**, **Connect TCP/IP**.
- TCP/IP input field with **Connect** button.
- **Recent devices** list with **Connect** action per row.
- Status banner area for connection state.

### Error States
- USB device unauthorized.
- Device offline/unreachable.
- Port blocked or refused.
- ADB not available.

## Flow: Mirroring Controls (Start/Stop, Resolution, FPS, Bitrate, Audio, Input)
### Summary
Provide controls for starting/stopping mirroring and adjusting streaming options.

### Steps
1. User configures options (resolution, FPS, bitrate, audio, input) in **Options** panel.
2. User clicks **Start Mirroring**.
3. **Mirroring active** shows stream and controls.
4. User clicks **Stop Mirroring** ➜ **Mirroring paused/stopped**.

### Acceptance Criteria
- Options are editable before start; disabled or applied live per capability.
- Start button disabled until a device is connected.
- Stop button ends stream and returns to connected state.
- Audio/input toggles clearly reflect enabled/disabled state.
- Resolution/FPS/bitrate values display current selection and limits.

### Wireframe: Key Widgets
- **Start Mirroring** / **Stop Mirroring** primary button.
- Dropdowns: **Resolution**, **FPS**.
- Input: **Bitrate** (numeric + units).
- Toggles: **Audio**, **Enable Input**.
- Status bar: streaming indicator, connection info.

### Error States
- Unsupported resolution/FPS for device.
- Bitrate value out of range.
- Audio capture not supported.
- Input disabled due to permissions.

## Flow: Session Persistence (Recent Devices, Last‑Used Options)
### Summary
Persist recent devices and last‑used options across app restarts.

### Steps
1. On successful connection, store device metadata and options used.
2. On app launch, load **Recent devices** and **Last used options**.
3. Preselect last‑used options in the **Options** panel.
4. Allow user to clear recent devices list.

### Acceptance Criteria
- Recent devices persist across restarts and are capped (e.g., last 5–10).
- Last‑used options are loaded and visible on startup.
- User can clear recent devices via a settings action.
- Failed or partial connections do not overwrite last‑used options.

### Wireframe: Key Widgets
- **Recent devices** section with clear/remove action.
- **Options** panel with prefilled values.
- Settings menu item: **Clear Recent Devices**.

### Error States
- Corrupt persistence data (fallback to defaults).
- Removed device entry gracefully handled (not selectable).
