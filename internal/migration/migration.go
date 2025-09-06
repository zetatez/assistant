package migration

import (
	"assistant/pkg/hash"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

type Migration struct {
	CommitID string
	UpSQL    string
	DownSQL  string
}

func Migrate(log *logrus.Logger, db *sql.DB) {
	err := createMigrationTable(db)
	if err != nil {
		log.Fatalf("❌ migrate failed: %v", err)
	}

	// tables
	for _, v := range tables {
		v.CommitID = "table: " + hash.Sha1(v.UpSQL+v.DownSQL)

		ok, err := alreadyApplied(db, v.CommitID)
		if err != nil {
			log.Fatalf("❌ migrate failed: %v", err)
		}
		if ok {
			continue
		}

		if _, err := db.Exec(v.UpSQL); err != nil {
			log.Fatalf("❌ migrate table failed: %v, %s", err, v.UpSQL)
		}

		if err = recordMigration(db, v); err != nil {
			log.Fatalf("❌ migrate failed: %v", err)
		}
	}

	// changes
	for _, v := range changes {
		v.CommitID = hash.Sha256(v.UpSQL + v.DownSQL)

		ok, err := alreadyApplied(db, v.CommitID)
		if err != nil {
			log.Fatalf("❌ migrate failed: %v", err)
		}
		if ok {
			continue
		}

		if _, err := db.Exec(v.UpSQL); err != nil {
			log.Fatalf("❌ migrate changes failed: %v, %s", err, v.UpSQL)
		}

		if err = recordMigration(db, v); err != nil {
			log.Fatalf("❌ migrate failed: %v", err)
		}
	}

	// init admin
	err = initAdmin(db)
	if err != nil {
		log.Fatalf("❌ migrate failed: init admin failed, %v", err)
	}
}

func createMigrationTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS migration (
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
	if _, err := db.Exec(query); err != nil {
		return err
	}
	return nil
}

func alreadyApplied(db *sql.DB, commit_id string) (bool, error) {
	var count int
	query := "SELECT COUNT(1) FROM migration WHERE commit_id = ?"
	err := db.QueryRow(query, commit_id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func recordMigration(db *sql.DB, v Migration) error {
	query := "INSERT INTO migration (commit_id, up_sql, down_sql) VALUES (?, ?, ?)"
	_, err := db.Exec(query, v.CommitID, v.UpSQL, v.DownSQL)
	if err != nil {
		return fmt.Errorf("failed to record migration")
	}
	return nil
}

func initAdmin(db *sql.DB) error {
	password, _ := hash.HashPassword("AAaa00__")
	user := struct {
		UserName string
		Password string
		Email    string
	}{UserName: "admin", Password: password, Email: ""}
	_, err := db.Exec(
		"insert ignore into user (user_name, password, email) values(?, ?, ?)",
		user.UserName,
		user.Password,
		user.Email,
	)
	if err != nil {
		return err
	}
	return nil
}
