// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"path/filepath"
	"testing"

	db "github.com/Rain-kl/Wavelet/internal/infra/persistence"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestExportSQLiteSnapshotIncludesWAL(t *testing.T) {
	dir := t.TempDir()
	source, err := gorm.Open(sqlite.Open(filepath.Join(dir, "source.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := source.DB()
	if err != nil {
		t.Fatal(err)
	}
	previous := db.DB(context.Background())
	db.SetDB(source)
	t.Cleanup(func() {
		db.SetDB(previous)
		if err := sqlDB.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, query := range []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA wal_autocheckpoint = 0",
		"CREATE TABLE snapshot_data (value TEXT)",
		"INSERT INTO snapshot_data VALUES ('latest committed data')",
	} {
		if err := source.Exec(query).Error; err != nil {
			t.Fatal(err)
		}
	}
	target := filepath.Join(dir, "snapshot.db")
	if err := ExportSQLiteSnapshot(context.Background(), target); err != nil {
		t.Fatal(err)
	}
	snapshot, err := gorm.Open(sqlite.Open(target), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	snapshotDB, err := snapshot.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := snapshotDB.Close(); err != nil {
			t.Error(err)
		}
	})
	var value string
	if err := snapshot.Raw("SELECT value FROM snapshot_data").Scan(&value).Error; err != nil {
		t.Fatal(err)
	}
	if value != "latest committed data" {
		t.Fatalf("unexpected snapshot content: %q", value)
	}
	var integrity string
	if err := snapshot.Raw("PRAGMA integrity_check").Scan(&integrity).Error; err != nil {
		t.Fatal(err)
	}
	if integrity != "ok" {
		t.Fatalf("snapshot integrity: %s", integrity)
	}
}
