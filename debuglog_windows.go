// SPDX-License-Identifier: MIT
//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const diagnosticVersion = "1.0.12-stable"

var (
	diagnosticLogMu   sync.Mutex
	diagnosticLogFile *os.File
	diagnosticLogPath string
)

func initDiagnosticLog() {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		if userDir, err := os.UserConfigDir(); err == nil {
			base = userDir
		}
	}
	if base == "" {
		base = os.TempDir()
	}

	dir := filepath.Join(base, appName)
	_ = os.MkdirAll(dir, 0o755)
	diagnosticLogPath = filepath.Join(dir, "DualSenseHUB-debug.log")

	if info, err := os.Stat(diagnosticLogPath); err == nil && info.Size() > 4*1024*1024 {
		_ = os.Rename(diagnosticLogPath, diagnosticLogPath+".old")
	}

	f, err := os.OpenFile(diagnosticLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err == nil {
		diagnosticLogFile = f
	}

	diagnosticLogf("============================================================")
	diagnosticLogf("SESSION START version=%s go=%s os=%s arch=%s", diagnosticVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	diagnosticLogf("log_path=%s", diagnosticLogPath)
}

func diagnosticLogf(format string, args ...any) {
	diagnosticLogMu.Lock()
	defer diagnosticLogMu.Unlock()

	line := fmt.Sprintf("%s | %s\r\n", time.Now().Format("2006-01-02 15:04:05.000"), fmt.Sprintf(format, args...))
	if diagnosticLogFile != nil {
		_, _ = diagnosticLogFile.WriteString(line)
		_ = diagnosticLogFile.Sync()
	}
}

func closeDiagnosticLog() {
	diagnosticLogMu.Lock()
	defer diagnosticLogMu.Unlock()
	if diagnosticLogFile != nil {
		_, _ = diagnosticLogFile.WriteString(time.Now().Format("2006-01-02 15:04:05.000") + " | SESSION END\r\n")
		_ = diagnosticLogFile.Sync()
		_ = diagnosticLogFile.Close()
		diagnosticLogFile = nil
	}
}
