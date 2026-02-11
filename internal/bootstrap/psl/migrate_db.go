package psl

import (
	"assistant/pkg/hash"
	"context"
	"database/sql"
	"fmt"
)

type UpDownSQL struct {
	UpSQL   string
	DownSQL string
}

func MigrateDB() {
	logger := GetLogger()
	logger.Info("migrate db...")

	if err := initSysMigrateTable(); err != nil {
		logger.Fatalf("migrate failed: %v", err)
	}

	if err := initUserTables(userTables); err != nil {
		logger.Fatalf("migrate failed: %v", err)
	}

	if err := initAdmin(); err != nil {
		logger.Fatalf("migrate failed: %v", err)
	}

	logger.Info("migrate db success")
}

func initSysMigrateTable() error {
	ctx := context.Background()

	query := `
	CREATE TABLE IF NOT EXISTS sys_migrate (
		id bigint NOT NULL AUTO_INCREMENT,
		gmt_create timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		gmt_modified timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		commit_id varchar(256) NOT NULL,
		up_sql text NOT NULL,
		down_sql text NOT NULL,
		PRIMARY KEY (id),
		UNIQUE KEY uk_ci (commit_id)
	) COMMENT='数据库变更记录';
	`
	_, err := GetDB().ExecContext(ctx, query)
	return err
}

func initUserTables(changes []UpDownSQL) error {
	logger := GetLogger()
	logger.Infof("processing %d changes...", len(changes))

	appliedCount := 0
	skippedCount := 0

	for i, m := range changes {
		commitID := hash.SHA256([]byte(m.UpSQL + m.DownSQL))

		applied, err := alreadyApplied(commitID)
		if err != nil {
			return err
		}
		if applied {
			skippedCount++
			logger.Debugf("skipping change %d/%d (already applied)", i+1, len(changes))
			continue
		}

		tx, err := GetDB().Begin()
		if err != nil {
			return fmt.Errorf("begin transaction: %w", err)
		}

		if _, err := tx.Exec(m.UpSQL); err != nil {
			rbErr := tx.Rollback()
			if rbErr != nil {
				return fmt.Errorf("change %d/%d failed: %w (rollback failed: %v)\nSQL: %s", i+1, len(changes), err, rbErr, m.UpSQL)
			}
			return fmt.Errorf("change %d/%d failed: %w\nSQL: %s", i+1, len(changes), err, m.UpSQL)
		}

		recordSQL := "INSERT IGNORE INTO sys_migrate (commit_id, up_sql, down_sql) VALUES (?, ?, ?)"
		if _, err := tx.Exec(recordSQL, commitID, m.UpSQL, m.DownSQL); err != nil {
			rbErr := tx.Rollback()
			if rbErr != nil {
				return fmt.Errorf("record change %d/%d: %w (rollback failed: %v)\nSQL: %s", i+1, len(changes), err, rbErr, recordSQL)
			}
			return fmt.Errorf("record change %d/%d: %w\nSQL: %s", i+1, len(changes), err, recordSQL)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit transaction: %w", err)
		}

		appliedCount++
		logger.Infof("applied change %d/%d", i+1, len(changes))
	}

	logger.Infof("changes completed: %d applied, %d skipped", appliedCount, skippedCount)
	return nil
}

func alreadyApplied(commitID string) (bool, error) {
	const query = "SELECT 1 FROM sys_migrate WHERE commit_id = ? LIMIT 1"
	var one int
	err := GetDB().QueryRow(query, commitID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func adminUserExists(ctx context.Context, username string) (bool, error) {
	const q = "SELECT 1 FROM user WHERE user_name = ? LIMIT 1"
	var one int
	err := GetDB().QueryRowContext(ctx, q, username).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check admin user exists: %w", err)
	}
	return true, nil
}

func initAdmin() error {
	logger := GetLogger()
	logger.Info("initializing admin user...")
	ctx := context.Background()

	cfg := GetConfig()
	adminConfig := cfg.App.Root

	if adminConfig.Username == "" {
		adminConfig.Username = "admin"
	}
	if adminConfig.Password == "" {
		adminConfig.Password = "AAaa00__"
	}

	exists, err := adminUserExists(ctx, adminConfig.Username)
	if err != nil {
		return err
	}
	if exists {
		logger.Infof("admin user '%s' already exists, skip", adminConfig.Username)
		return nil
	}

	password, err := hash.HashPassword(adminConfig.Password)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	dml := "INSERT IGNORE INTO user (user_name, password, email) VALUES (?, ?, ?)"
	result, err := GetDB().Exec(dml, adminConfig.Username, password, adminConfig.Email)
	if err != nil {
		return fmt.Errorf("failed to insert admin user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected > 0 {
		logger.Infof("admin user '%s' created", adminConfig.Username)
	} else {
		logger.Infof("admin user '%s' already exists", adminConfig.Username)
	}

	return nil
}
