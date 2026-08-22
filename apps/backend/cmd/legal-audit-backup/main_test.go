package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPruneExpiredFilesRemovesOnlyFullyExpiredBackups(t *testing.T) {
	t.Parallel()
	destination := t.TempDir()
	expiredName, activeName := "expired.json.aes", "active.json.aes"
	for _, name := range []string{expiredName, activeName} {
		if err := os.WriteFile(filepath.Join(destination, name), []byte("test"), 0600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	now := time.Now().UTC()
	expiredAt := now.Add(-time.Minute)
	activeAt := now.Add(time.Minute)
	state := checkpoint{
		Records: map[string]recordState{
			"expired": {RetentionUntil: &expiredAt},
			"active":  {RetentionUntil: &activeAt},
		},
		Files: map[string]fileState{
			expiredName: {RecordIDs: []string{"expired"}},
			activeName:  {RecordIDs: []string{"active"}},
		},
	}
	if err := pruneExpiredFiles(destination, &state, now); err != nil {
		t.Fatalf("pruneExpiredFiles() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(destination, expiredName)); !os.IsNotExist(err) {
		t.Fatalf("expired backup still exists, err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(destination, activeName)); err != nil {
		t.Fatalf("active backup missing, err = %v", err)
	}
	if _, exists := state.Files[expiredName]; exists {
		t.Fatal("expired file state was not removed")
	}
}
