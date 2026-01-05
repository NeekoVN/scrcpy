package backend

import (
	"context"
	"fmt"
)

type ErrorKind string

const (
	ErrorKindTimeout       ErrorKind = "timeout"
	ErrorKindCommandFailed ErrorKind = "command_failed"
	ErrorKindParse         ErrorKind = "parse"
	ErrorKindInvalidInput  ErrorKind = "invalid_input"
	ErrorKindNotRunning    ErrorKind = "not_running"
)

type UIError struct {
	Kind     ErrorKind
	Message  string
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

func (e *UIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Command != "" {
		return fmt.Sprintf("%s: %s", e.Kind, e.Command)
	}
	return string(e.Kind)
}

func (e *UIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func newInvalidInputError(message string) *UIError {
	return &UIError{
		Kind:    ErrorKindInvalidInput,
		Message: message,
	}
}

func newNotRunningError(message string) *UIError {
	return &UIError{
		Kind:    ErrorKindNotRunning,
		Message: message,
	}
}

func newTimeoutError(command string, err error) *UIError {
	return &UIError{
		Kind:    ErrorKindTimeout,
		Message: fmt.Sprintf("timeout while running %s", command),
		Command: command,
		Err:     err,
	}
}

func newCommandFailedError(command string, stdout string, stderr string, exitCode int, err error) *UIError {
	return &UIError{
		Kind:     ErrorKindCommandFailed,
		Message:  fmt.Sprintf("command failed: %s", command),
		Command:  command,
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
		Err:      err,
	}
}

func newParseError(command string, stdout string, stderr string, err error) *UIError {
	return &UIError{
		Kind:    ErrorKindParse,
		Message: fmt.Sprintf("unexpected output from %s", command),
		Command: command,
		Stdout:  stdout,
		Stderr:  stderr,
		Err:     err,
	}
}

func isTimeout(err error) bool {
	return err == context.DeadlineExceeded || err == context.Canceled
}
