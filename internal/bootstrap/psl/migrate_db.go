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

func MigrateDB(ctx context.Context) error {
	logger := GetLogger()
	logger.Info("migrate db...")

	if err := initSysMigrateTable(ctx); err != nil {
		return err
	}

	if err := initUserTables(ctx, userTables); err != nil {
		return err
	}

	if err := initDefaultUsers(ctx); err != nil {
		return err
	}

	logger.Info("migrate db success")
	return nil
}

func initSysMigrateTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS sys_migrate (
		id bigint NOT NULL AUTO_INCREMENT,
		gmt_create timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		gmt_modified timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		commit_id varchar(256) NOT NULL DEFAULT '',
		up_sql text NOT NULL,
		down_sql text NOT NULL,
		PRIMARY KEY (id),
		UNIQUE KEY uk_ci (commit_id)
	) COMMENT='数据库变更记录';
	`
	_, err := GetDB().ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return err
}

func initUserTables(ctx context.Context, changes []UpDownSQL) error {
	logger := GetLogger()
	logger.Infof("processing %d changes...", len(changes))

	appliedCount := 0
	skippedCount := 0

	for i, m := range changes {
		commitID := hash.SHA256([]byte(m.UpSQL + m.DownSQL))

		applied, err := alreadyApplied(ctx, commitID)
		if err != nil {
			return err
		}
		if applied {
			skippedCount++
			logger.Debugf("already applied, skipping change %d/%d", i+1, len(changes))
			continue
		}

		tx, err := GetDB().BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin transaction: %w", err)
		}

		if _, err := tx.ExecContext(ctx, m.UpSQL); err != nil {
			rbErr := tx.Rollback()
			if rbErr != nil {
				return fmt.Errorf("change %d/%d failed: %w (rollback failed: %v)\nSQL: %s", i+1, len(changes), err, rbErr, m.UpSQL)
			}
			return fmt.Errorf("change %d/%d failed: %w\nSQL: %s", i+1, len(changes), err, m.UpSQL)
		}

		recordSQL := "INSERT IGNORE INTO sys_migrate (commit_id, up_sql, down_sql) VALUES (?, ?, ?)"
		if _, err := tx.ExecContext(ctx, recordSQL, commitID, m.UpSQL, m.DownSQL); err != nil {
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

func alreadyApplied(ctx context.Context, commitID string) (bool, error) {
	const query = "SELECT 1 FROM sys_migrate WHERE commit_id = ? LIMIT 1"
	var one int
	err := GetDB().QueryRowContext(ctx, query, commitID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func sysUserExists(ctx context.Context, username string) (bool, error) {
	const q = "SELECT 1 FROM sys_user WHERE user_name = ? LIMIT 1"
	var one int
	err := GetDB().QueryRowContext(ctx, q, username).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check user '%s' exists: %w", username, err)
	}
	return true, nil
}

type SysUser struct {
	Username string
	Password string
	Email    string
}

func createSysUser(ctx context.Context, u SysUser) (bool, error) {
	if u.Username == "" {
		return false, fmt.Errorf("username is empty")
	}

	exists, err := sysUserExists(ctx, u.Username)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	password, err := hash.HashPassword(u.Password)
	if err != nil {
		return false, fmt.Errorf("hash password: %w", err)
	}

	const dml = "INSERT IGNORE INTO sys_user (user_name, password, email) VALUES (?, ?, ?)"
	result, err := GetDB().ExecContext(ctx, dml, u.Username, password, u.Email)
	if err != nil {
		return false, fmt.Errorf("insert user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("get rows affected: %w", err)
	}
	return rowsAffected > 0, nil
}

func initDefaultUsers(ctx context.Context) error {
	logger := GetLogger()
	logger.Info("initializing default users...")

	cfg := GetConfig()
	adminConfig := cfg.App.Root

	adminUser := SysUser{
		Username: adminConfig.Username,
		Password: adminConfig.Password,
		Email:    adminConfig.Email,
	}
	if adminUser.Username == "" {
		adminUser.Username = "admin"
	}
	if adminUser.Password == "" {
		adminUser.Password = "AAaa00__"
	}

	guestUser := SysUser{
		Username: "guest",
		Password: "guest",
		Email:    "guest@local",
	}

	users := []SysUser{adminUser, guestUser}

	for _, u := range users {
		created, err := createSysUser(ctx, u)
		if err != nil {
			return fmt.Errorf("init default user '%s': %w", u.Username, err)
		}
		if created {
			logger.Infof("default user '%s' created", u.Username)
		} else {
			logger.Infof("default user '%s' already exists, skip", u.Username)
		}
	}

	return nil
}
