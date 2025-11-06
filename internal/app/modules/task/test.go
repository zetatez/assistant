package task

import (
	"database/sql"
	"fmt"
)

func CanSetParent(db *sql.DB, taskID, newParentID int64) (bool, error) {
	if taskID == newParentID {
		return false, fmt.Errorf("不能将自己设置为父任务")
	}

	query := `
	WITH RECURSIVE descendants AS (
		SELECT id FROM task WHERE id = ?
		UNION ALL
		SELECT t.id FROM task t
		INNER JOIN descendants d ON t.parent_id = d.id
	)
	SELECT COUNT(*) FROM descendants WHERE id = ?;
	`

	var count int
	if err := db.QueryRow(query, taskID, newParentID).Scan(&count); err != nil {
		return false, err
	}
	if count > 0 {
		return false, fmt.Errorf("检测到循环依赖，禁止形成环")
	}
	return true, nil
}
