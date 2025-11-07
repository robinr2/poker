package server

import (
	"os"
	"testing"
)

func TestWorkingDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Working directory: %s", cwd)

	// Check if web/static/index.html exists
	info, err := os.Stat("web/static/index.html")
	t.Logf("web/static/index.html stat error: %v", err)
	if err == nil {
		t.Logf("web/static/index.html exists: %v bytes", info.Size())
	}
}
