// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package localfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAtomicReplacesCompleteFileAndCleansFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "identity.json")
	if err := WriteAtomic(path, []byte("old identity"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomic(path, []byte("new identity"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "new identity" {
		t.Errorf("WriteAtomic replacement = %q, %v; want complete new identity", data, err)
	}
	blocked := filepath.Join(dir, "directory")
	if err = os.Mkdir(blocked, 0o700); err != nil {
		t.Fatal(err)
	}
	if err = WriteAtomic(blocked, []byte("cannot replace directory"), 0o600); err == nil {
		t.Fatal("WriteAtomic(directory) succeeded, want rename failure")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("WriteAtomic(after failure) directory entries = %d, want 2 with no temporary file", len(entries))
	}
}
