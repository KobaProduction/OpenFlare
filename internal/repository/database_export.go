// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"errors"

	db "github.com/Rain-kl/Wavelet/internal/infra/persistence"
)

// ExportSQLiteSnapshot 生成包含 WAL 中已提交数据的一致性快照，目标文件须不存在或为空。
func ExportSQLiteSnapshot(ctx context.Context, path string) error {
	database := db.DB(ctx)
	if database == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	return database.Exec("VACUUM INTO ?", path).Error
}
