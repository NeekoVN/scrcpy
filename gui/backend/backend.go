package backend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultAdbTimeout    = 10 * time.Second
	defaultScrcpyTimeout = 24 * time.Hour
)

type Device struct {
	ID    string
	State string
}

type ScrcpyOptions struct {
	BitRate       string
	MaxSize       int
	MaxFps        int
	TurnScreenOff bool
	Fullscreen    bool
	StayAwake     bool
	Record        string
	WindowTitle   string
	ExtraArgs     []string
}

type Backend struct {
	adbPath       string
	scrcpyPath    string
	adbTimeout    time.Duration
	scrcpyTimeout time.Duration

	mu           sync.Mutex
	scrcpyCmd    *exec.Cmd
	scrcpyCancel context.CancelFunc
}

func NewBackend(adbPath string, scrcpyPath string) *Backend {
	if adbPath == "" {
		adbPath = "adb"
	}
	if scrcpyPath == "" {
		scrcpyPath = "scrcpy"
	}
	return &Backend{
		adbPath:       adbPath,
		scrcpyPath:    scrcpyPath,
		adbTimeout:    defaultAdbTimeout,
		scrcpyTimeout: defaultScrcpyTimeout,
	}
}

func (b *Backend) ListDevices() ([]Device, error) {
	stdout, stderr, err := b.runAdb([]string{"devices"})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(stdout, "\n")
	devices := make([]Device, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices attached") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, newParseError("adb devices", stdout, stderr, fmt.Errorf("unexpected device line: %q", line))
		}
		devices = append(devices, Device{ID: fields[0], State: fields[1]})
	}
	return devices, nil
}

func (b *Backend) PairDevice(ip string, port int, code string) error {
	if ip == "" || port <= 0 || code == "" {
		return newInvalidInputError("pair requires ip, port, and code")
	}
	endpoint := fmt.Sprintf("%s:%d", ip, port)
	stdout, stderr, err := b.runAdb([]string{"pair", endpoint, code})
	if err != nil {
		return err
	}
	if containsAny(stdout, "Successfully paired to", "already paired") {
		return nil
	}
	if containsAny(stdout, "failed", "error") {
		return newCommandFailedError("adb pair", stdout, stderr, 0, fmt.Errorf("pair failed"))
	}
	return newParseError("adb pair", stdout, stderr, fmt.Errorf("unexpected output"))
}

func (b *Backend) ConnectDevice(ip string, port int) error {
	if ip == "" || port <= 0 {
		return newInvalidInputError("connect requires ip and port")
	}
	endpoint := fmt.Sprintf("%s:%d", ip, port)
	stdout, stderr, err := b.runAdb([]string{"connect", endpoint})
	if err != nil {
		return err
	}
	if containsAny(stdout, "connected to", "already connected") {
		return nil
	}
	if containsAny(stdout, "failed", "unable") {
		return newCommandFailedError("adb connect", stdout, stderr, 0, fmt.Errorf("connect failed"))
	}
	return newParseError("adb connect", stdout, stderr, fmt.Errorf("unexpected output"))
}

func (b *Backend) EnableTCPIP(port int) error {
	if port <= 0 {
		return newInvalidInputError("tcpip requires a port")
	}
	stdout, stderr, err := b.runAdb([]string{"tcpip", strconv.Itoa(port)})
	if err != nil {
		return err
	}
	if containsAny(stdout, "restarting in TCP mode", "already in tcp") {
		return nil
	}
	return newParseError("adb tcpip", stdout, stderr, fmt.Errorf("unexpected output"))
}

func (b *Backend) DisconnectDevice(ip string, port int) error {
	if ip == "" || port <= 0 {
		return newInvalidInputError("disconnect requires ip and port")
	}
	endpoint := fmt.Sprintf("%s:%d", ip, port)
	stdout, stderr, err := b.runAdb([]string{"disconnect", endpoint})
	if err != nil {
		return err
	}
	if containsAny(stdout, "disconnected", "no such device") {
		return nil
	}
	return newParseError("adb disconnect", stdout, stderr, fmt.Errorf("unexpected output"))
}

func (b *Backend) StartScrcpy(deviceID string, options ScrcpyOptions) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.scrcpyCmd != nil {
		return newCommandFailedError("scrcpy", "", "", 0, fmt.Errorf("scrcpy already running"))
	}

	args := buildScrcpyArgs(deviceID, options)
	ctx, cancel := context.WithTimeout(context.Background(), b.scrcpyTimeout)
	cmd := exec.CommandContext(ctx, b.scrcpyPath, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		cancel()
		return newCommandFailedError("scrcpy", "", "", 0, err)
	}

	b.scrcpyCmd = cmd
	b.scrcpyCancel = cancel

	go func() {
		_ = cmd.Wait()
		b.mu.Lock()
		b.scrcpyCmd = nil
		b.scrcpyCancel = nil
		b.mu.Unlock()
	}()

	return nil
}

func (b *Backend) StopScrcpy() error {
	b.mu.Lock()
	cmd := b.scrcpyCmd
	cancel := b.scrcpyCancel
	b.mu.Unlock()

	if cmd == nil || cancel == nil {
		return newNotRunningError("scrcpy is not running")
	}

	cancel()
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer waitCancel()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-waitCtx.Done():
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return newTimeoutError("scrcpy", waitCtx.Err())
	case <-done:
		return nil
	}
}

func buildScrcpyArgs(deviceID string, options ScrcpyOptions) []string {
	args := make([]string, 0, 8)
	if deviceID != "" {
		args = append(args, "-s", deviceID)
	}
	if options.BitRate != "" {
		args = append(args, "--bit-rate", options.BitRate)
	}
	if options.MaxSize > 0 {
		args = append(args, "--max-size", strconv.Itoa(options.MaxSize))
	}
	if options.MaxFps > 0 {
		args = append(args, "--max-fps", strconv.Itoa(options.MaxFps))
	}
	if options.TurnScreenOff {
		args = append(args, "--turn-screen-off")
	}
	if options.Fullscreen {
		args = append(args, "--fullscreen")
	}
	if options.StayAwake {
		args = append(args, "--stay-awake")
	}
	if options.Record != "" {
		args = append(args, "--record", options.Record)
	}
	if options.WindowTitle != "" {
		args = append(args, "--window-title", options.WindowTitle)
	}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	return args
}

func (b *Backend) runAdb(args []string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.adbTimeout)
	defer cancel()
	return runCommand(ctx, b.adbPath, args)
}

func runCommand(ctx context.Context, command string, args []string) (string, string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outStr := strings.TrimSpace(stdout.String())
	errStr := strings.TrimSpace(stderr.String())
	if err != nil {
		if isTimeout(err) || ctx.Err() != nil {
			return outStr, errStr, newTimeoutError(command, err)
		}
		exitCode := 0
		if exitErr := new(exec.ExitError); err != nil && errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		return outStr, errStr, newCommandFailedError(command, outStr, errStr, exitCode, err)
	}
	return outStr, errStr, nil
}

func containsAny(haystack string, needles ...string) bool {
	lower := strings.ToLower(haystack)
	for _, needle := range needles {
		if strings.Contains(lower, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}
