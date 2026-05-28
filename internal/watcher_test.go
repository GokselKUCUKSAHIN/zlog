package internal

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewFileWatcher(t *testing.T) {
	called := false
	w := NewFileWatcher("/tmp/test.json", func() { called = true })
	if w == nil {
		t.Fatal("Expected non-nil FileWatcher")
	}
	if w.path != "/tmp/test.json" {
		t.Errorf("Expected path '/tmp/test.json', got %q", w.path)
	}
	if w.onChange == nil {
		t.Error("Expected onChange to be set")
	}
	if w.done == nil {
		t.Error("Expected done channel to be initialized")
	}
	if called {
		t.Error("onChange should not be called on creation")
	}
}
func TestFileWatcher_Start_NilOnChange(t *testing.T) {
	w := NewFileWatcher("/tmp/test.json", nil)
	err := w.Start()
	if err != nil {
		t.Errorf("Expected nil error for nil onChange, got %v", err)
	}
}
func TestFileWatcher_Start_NonExistentFile(t *testing.T) {
	w := NewFileWatcher("/tmp/non_existent_file_xyz_abc.json", func() {})
	err := w.Start()
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}
func TestFileWatcher_Start_ValidFile(t *testing.T) {
	tmpFile := createTempFile(t, `{"test": true}`)
	w := NewFileWatcher(tmpFile, func() {})
	defer w.Stop()
	err := w.Start()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
func TestFileWatcher_StartOnce_CalledMultipleTimes(t *testing.T) {
	tmpFile := createTempFile(t, `{"test": true}`)
	w := NewFileWatcher(tmpFile, func() {})
	defer w.Stop()
	err1 := w.Start()
	if err1 != nil {
		t.Fatalf("First Start failed: %v", err1)
	}
	err2 := w.Start()
	if err2 != nil {
		t.Fatalf("Second Start returned error: %v", err2)
	}
}
func TestFileWatcher_Stop_MultipleTimes(t *testing.T) {
	tmpFile := createTempFile(t, `{"test": true}`)
	w := NewFileWatcher(tmpFile, func() {})
	err := w.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	w.Stop()
	w.Stop()
}
func TestFileWatcher_DetectsChange(t *testing.T) {
	tmpFile := createTempFile(t, `{"version": 1}`)
	var callCount atomic.Int32
	w := NewFileWatcher(tmpFile, func() {
		callCount.Add(1)
	})
	defer w.Stop()
	err := w.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	err = os.WriteFile(tmpFile, []byte(`{"version": 2}`), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	time.Sleep(11 * time.Second)
	if callCount.Load() == 0 {
		t.Error("Expected onChange to be called at least once after file modification")
	}
}
func TestFileWatcher_StopPreventsCallbacks(t *testing.T) {
	tmpFile := createTempFile(t, `{"test": true}`)
	var callCount atomic.Int32
	w := NewFileWatcher(tmpFile, func() {
		callCount.Add(1)
	})
	err := w.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	w.Stop()
	time.Sleep(100 * time.Millisecond)
	_ = os.WriteFile(tmpFile, []byte(`{"changed": true}`), 0644)
	time.Sleep(200 * time.Millisecond)
	if callCount.Load() != 0 {
		t.Errorf("Expected no callbacks after Stop, got %d", callCount.Load())
	}
}
func TestFileWatcher_FileDeletedDuringWatch(t *testing.T) {
	tmpFile := createTempFile(t, `{"test": true}`)
	var callCount atomic.Int32
	w := NewFileWatcher(tmpFile, func() {
		callCount.Add(1)
	})
	err := w.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer w.Stop()
	os.Remove(tmpFile)
	time.Sleep(11 * time.Second)
	if callCount.Load() != 0 {
		t.Errorf("Expected no callbacks when file is deleted, got %d", callCount.Load())
	}
}
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test-config.json")
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return path
}
