package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRegisterDir(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	timestamps, err := registerDir(tempDir)
	if err != nil {
		t.Fatalf("registerDir failed: %v", err)
	}

	if len(timestamps) != 3 { // tempDir, test.txt, subdir
		t.Errorf("expected 3 entries, got %d", len(timestamps))
	}

	if _, exists := timestamps[testFile]; !exists {
		t.Error("test file not found in timestamps")
	}
}

func TestResetTimestamps(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	origTimestamps, err := registerDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatal(err)
	}

	defaultTime := time.Unix(0, 0)
	err = resetTimestamps(tempDir, defaultTime, origTimestamps)
	if err != nil {
		t.Fatalf("resetTimestamps failed: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if info.ModTime().Equal(origTimestamps[testFile].Mtime) {
		return
	}
}

func TestResetTimestampsNewFile(t *testing.T) {
	tempDir := t.TempDir()

	origTimestamps, err := registerDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	newFile := filepath.Join(tempDir, "new.txt")
	if err := os.WriteFile(newFile, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	defaultTime := time.Unix(1234567890, 0)
	err = resetTimestamps(tempDir, defaultTime, origTimestamps)
	if err != nil {
		t.Fatalf("resetTimestamps failed: %v", err)
	}

	info, err := os.Stat(newFile)
	if err != nil {
		t.Fatal(err)
	}

	if !info.ModTime().Equal(defaultTime) {
		t.Errorf("new file timestamp not set to default, got %v, want %v",
			info.ModTime(), defaultTime)
	}
}

func TestResetTimestampsWithSourceDateEpoch(t *testing.T) {
	originalEnv := os.Getenv("SOURCE_DATE_EPOCH")
	defer os.Setenv("SOURCE_DATE_EPOCH", originalEnv)

	os.Setenv("SOURCE_DATE_EPOCH", "1609459200")

	tempDir := t.TempDir()
	newFile := filepath.Join(tempDir, "test.txt")

	if err := os.WriteFile(newFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("SOURCE_DATE_EPOCH") != "1609459200" {
		t.Error("SOURCE_DATE_EPOCH not set correctly")
	}
}
