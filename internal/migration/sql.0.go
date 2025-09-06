package migration

var tables = []Migration{
	{
		CommitID: "basic",
		UpSQL: `
		CREATE TABLE IF NOT EXISTS user (
			id BIGINT AUTO_INCREMENT,
			gmt_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			user_name VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL,
			email VARCHAR(128) NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY (user_name),
			UNIQUE KEY (email)
		) COMMENT='用户表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS user;
		`,
	},
	{
		CommitID: "basic",
		UpSQL: `
		CREATE TABLE IF NOT EXISTS server (
			id BIGINT AUTO_INCREMENT,
			gmt_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			idc VARCHAR(64) NOT NULL DEFAULT '',
			svr_ip VARCHAR(32) NOT NULL DEFAULT '127.1',
			ak VARCHAR(255) NOT NULL DEFAULT '',
			sk TEXT NOT NULL,
			svr_status VARCHAR(32) NOT NULL DEFAULT '',
			PRIMARY KEY (id),
			UNIQUE KEY (svr_ip)
		) COMMENT='服务器表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS server;
		`,
	},
	{
		CommitID: "basic",
		UpSQL: `
		CREATE TABLE IF NOT EXISTS todo_list (
			id BIGINT AUTO_INCREMENT,
			gmt_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			title VARCHAR(255) NOT NULL,
			content text NOT NULL,
			deadline VARCHAR(128) NOT NULL,
			is_done BOOLEAN DEFAULT TRUE,
			PRIMARY KEY (id)
		) COMMENT='代办事项表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS todo_list;
		`,
	},
	{
		CommitID: "basic",
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_tmpl (
			id BIGINT AUTO_INCREMENT,
			gmt_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			config JSON NOT NULL,
			is_enable BOOLEAN DEFAULT TRUE,
			PRIMARY KEY (id),
			UNIQUE KEY (name)
		) COMMENT='任务模板表, 由多个原子任务编排而成';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_tmpl;
		`,
	},
	{
		CommitID: "basic",
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_atomic (
			id BIGINT AUTO_INCREMENT,
			gmt_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			task_type VARCHAR(48) NOT NULL,
			config JSON NOT NULL,
			is_enable BOOLEAN DEFAULT TRUE,
			PRIMARY KEY (id),
			UNIQUE KEY (name)
		) COMMENT='任务原子表'
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_atomic;
		`,
	},
	{
		CommitID: "basic",
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_composite (
			id BIGINT AUTO_INCREMENT,
			gmt_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			task_tmpl_id BIGINT UNSIGNED NOT NULL,
			task_atomic_id BIGINT UNSIGNED NOT NULL,
			step_order INT NOT NULL,
			depends_on JSON NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY (task_tmpl_id, step_order)
		) COMMENT='任务模板与原子任务映射关系';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_composite;
		`,
	},
}
