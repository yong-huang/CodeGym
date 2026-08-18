package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ErrTimeout is returned when code execution exceeds the configured limit.
var ErrTimeout = errors.New("execution timeout")

// Engine executes reconstructed user code. Per design section 8 the full
// isolation story (Docker/Kubernetes) is future work; this MVP uses a local
// subprocess with a hard timeout, which satisfies the "Local Process" engine.
type Engine struct {
	PythonBin string
	GoBin     string
}

// New returns an Engine configured with the given interpreters.
func New(pythonBin string) *Engine {
	return &Engine{PythonBin: pythonBin, GoBin: "go"}
}

// Result carries stdout/stderr plus timing info for one execution.
type Result struct {
	Stdout string
	Stderr string
	TimeMS int64
}

// ErrUnsupportedLanguage is returned when a tech stack has no local engine.
var ErrUnsupportedLanguage = errors.New("unsupported tech stack for execution")

// Run dispatches code execution to the engine for the given tech stack.
func (e *Engine) Run(ctx context.Context, techStack, code string, timeout time.Duration) (Result, error) {
	switch strings.ToLower(techStack) {
	case "python":
		return e.RunPython(ctx, code, timeout)
	case "go":
		return e.RunGo(ctx, code, timeout)
	default:
		return Result{}, ErrUnsupportedLanguage
	}
}

// RunPython executes `code` with python3 under a timeout. The first non-empty
// stdout line is the effective output used for judging.
func (e *Engine) RunPython(ctx context.Context, code string, timeout time.Duration) (Result, error) {
	if e.PythonBin == "" {
		e.PythonBin = "python3"
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(cctx, e.PythonBin, "-c", code)
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	runErr := runWithTimeout(cctx, cmd)
	return Result{
		Stdout: strings.TrimSpace(out.String()),
		Stderr: strings.TrimSpace(errb.String()),
		TimeMS: time.Since(start).Milliseconds(),
	}, runErr
}

// RunGo executes a complete Go program under a timeout by staging it in a
// temporary module and running `go run .`.
func (e *Engine) RunGo(ctx context.Context, code string, timeout time.Duration) (Result, error) {
	if e.GoBin == "" {
		e.GoBin = "go"
	}
	dir, err := os.MkdirTemp("", "codegym-go-*")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module codegymtmp\n\ngo 1.22\n"), 0o644); err != nil {
		return Result{}, err
	}
	// prepend package+import-less main guard: if the user code already starts
	// with `package main`, write it as-is.
	mainFile := code
	if !strings.HasPrefix(strings.TrimSpace(code), "package ") {
		mainFile = "package main\n\n" + code
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainFile), 0o644); err != nil {
		return Result{}, err
	}

	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(cctx, e.GoBin, "run", ".")
	cmd.Dir = dir
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	runErr := runWithTimeout(cctx, cmd)
	return Result{
		Stdout: strings.TrimSpace(out.String()),
		Stderr: strings.TrimSpace(errb.String()),
		TimeMS: time.Since(start).Milliseconds(),
	}, runErr
}

func runWithTimeout(cctx context.Context, cmd *exec.Cmd) error {
	runErr := cmd.Run()
	// CommandContext kills the process on deadline; surface a clean timeout error.
	if cctx.Err() == context.DeadlineExceeded {
		return ErrTimeout
	}
	return runErr
}

// EffectiveOutput returns the first non-empty stdout line; this is what the
// judge compares against expected_output.
func EffectiveOutput(r Result) string {
	for _, line := range strings.Split(r.Stdout, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// FormatError renders a subprocess failure (timeout / exit / stderr) for the UI.
func FormatError(runErr error, stderr string) string {
	if runErr == nil {
		return stderr
	}
	if errors.Is(runErr, ErrTimeout) {
		return "⏱️ 执行超时，请检查是否存在死循环"
	}
	if stderr != "" {
		return fmt.Sprintf("❌ 运行错误: %s", stderr)
	}
	return fmt.Sprintf("❌ 运行失败: %v", runErr)
}
