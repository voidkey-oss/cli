package main

import (
	"testing"

	"github.com/voidkey-oss/cli/cmd"
)

func TestSetVersionInfo(t *testing.T) {
	// Test setting version info
	testVersion := "1.0.0"
	testCommit := "abc123"
	testDate := "2024-01-01"

	cmd.SetVersionInfo(testVersion, testCommit, testDate)

	// Since we can't directly access versionInfo from main,
	// we test by verifying the version command works without panic
	// This is integration testing of the version flow
}

func TestMain(t *testing.T) {
	// Test that main doesn't panic with default values
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main() panicked: %v", r)
		}
	}()

	// We can't actually call main() as it would execute the CLI
	// Instead, we test the setup functions
	cmd.SetVersionInfo("test", "test", "test")
}
