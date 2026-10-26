// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"encoding/hex"
	"path/filepath"
	"testing"
)

func TestRegistrationKeySurvivesStaleSnapshotAndRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := NewStore(path)
	stale, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	key, err := store.EnsureRegistrationKey()
	if err != nil {
		t.Fatal(err)
	}
	if err = store.SetRegisteredNodeID("node-server-assigned"); err != nil {
		t.Fatal(err)
	}
	decoded, err := hex.DecodeString(key)
	if err != nil || len(decoded) != 32 {
		t.Fatalf("EnsureRegistrationKey() length = %d, error %v; want 32, nil", len(decoded), err)
	}
	stale.LastError = "background sync failed"
	if err = store.Save(stale); err != nil {
		t.Fatal(err)
	}
	restartedKey, err := NewStore(path).EnsureRegistrationKey()
	if err != nil {
		t.Fatal(err)
	}
	if restartedKey != key {
		t.Error("EnsureRegistrationKey(after stale save and restart) changed enrollment identity")
	}
	snapshot, err := NewStore(path).Load()
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.NodeID != "node-server-assigned" {
		t.Errorf("Load(after stale save).NodeID = %q, want server-assigned ID", snapshot.NodeID)
	}
}
