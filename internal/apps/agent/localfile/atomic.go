// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package localfile persists agent identity and configuration without truncating existing files.
package localfile

import (
	"os"
	"path/filepath"
)

// WriteAtomic replaces path only after its complete contents have been synced.
func WriteAtomic(path string, data []byte, perm os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()           // Best effort after an earlier write/close error.
		_ = os.Remove(file.Name()) // Best effort; rename normally removes the temporary path.
	}()
	if err = file.Chmod(perm); err != nil {
		return err
	}
	if _, err = file.Write(data); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
