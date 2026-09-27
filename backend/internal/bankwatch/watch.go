// Package bankwatch hot-loads question-bank JSON files from a directory:
// new or changed files are parsed and upserted into the store while the
// server keeps running, so adding a bank file does not require a restart.
package bankwatch

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codegym/backend/internal/seed"
	"codegym/backend/internal/store"
)

type fileState struct {
	modTime time.Time
	size    int64
}

// Watcher periodically scans a bank directory for new or changed files.
type Watcher struct {
	dir      string
	interval time.Duration
	st       *store.Store
	logf     func(format string, args ...any)
	seen     map[string]fileState
}

// New creates a watcher for dir. interval <= 0 disables periodic scanning
// (ScanOnce still works).
func New(dir string, interval time.Duration, st *store.Store) *Watcher {
	return &Watcher{
		dir: dir, interval: interval, st: st,
		logf: log.Printf, seen: map[string]fileState{},
	}
}

// SetLogger overrides the log sink (used by tests).
func (w *Watcher) SetLogger(fn func(string, ...any)) { w.logf = fn }

// Run scans immediately, then every interval until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	w.ScanOnce()
	if w.interval <= 0 {
		return
	}
	t := time.NewTicker(w.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.ScanOnce()
		}
	}
}

// ScanOnce loads every new/changed *.json bank file and returns how many files
// were (re)loaded. Parse errors are logged and retried on the next scan.
func (w *Watcher) ScanOnce() int {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		w.logf("bankwatch: 读取目录 %s 失败: %v", w.dir, err)
		return 0
	}
	loaded := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		state := fileState{modTime: info.ModTime(), size: info.Size()}
		if prev, ok := w.seen[e.Name()]; ok && prev.size == state.size && prev.modTime.Equal(state.modTime) {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		exs, err := seed.BuildBank(w.dir, name)
		if err != nil {
			w.logf("bankwatch: 跳过 %s（下次扫描重试）: %v", e.Name(), err)
			continue
		}
		ok := true
		for _, ex := range exs {
			if err := w.st.UpsertExercise(ex); err != nil {
				w.logf("bankwatch: 写入 %s 失败: %v", ex.ID, err)
				ok = false
			}
		}
		if !ok {
			continue
		}
		if err := w.st.SaveBankSource(name, "dir", len(exs), ""); err != nil {
			w.logf("bankwatch: 记录题库来源 %s 失败: %v", name, err)
		}
		w.seen[e.Name()] = state
		loaded++
		w.logf("bankwatch: 已热加载题库 %s（%d 题）", filepath.Base(e.Name()), len(exs))
	}
	return loaded
}
