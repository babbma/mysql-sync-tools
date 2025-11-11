package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/db-sync-tools/internal/logger"
)

func TestDailyRotation_UsesIsolatedTempDir(t *testing.T) {
	tmp := t.TempDir()
	logDir := filepath.Join(tmp, "logs")
	logFile := filepath.Join(logDir, "sync.log")

	if err := logger.Init("info", logFile, false); err != nil {
		t.Fatalf("init logger failed: %v", err)
	}
	defer logger.Close()

	logger.Info("first line")
	if _, err := os.Stat(logFile); err != nil {
		t.Fatalf("expected main log file exists, got err: %v", err)
	}

	yesterday := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	logger.TestSetCurrentDay(yesterday)

	logger.Info("new day line")

	base := "sync"
	ext := ".log"
	backup := filepath.Join(logDir, base+"."+yesterday+ext)
	_, statErr := os.Stat(backup)
	if statErr != nil {
		pattern := filepath.Join(logDir, base+"."+yesterday+"-*.log")
		matches, _ := filepath.Glob(pattern)
		if len(matches) == 0 {
			t.Fatalf("expected backup file not found: %s or %s", backup, pattern)
		}
	}

	if _, err := os.Stat(logFile); err != nil {
		t.Fatalf("expected main log file exists after rotation, got err: %v", err)
	}
}
