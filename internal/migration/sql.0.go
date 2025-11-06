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
		CREATE TABLE IF NOT EXISTS task_def (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			name VARCHAR(255) NOT NULL COMMENT '任务名称',
			description TEXT COMMENT '任务描述',
			task_type ENUM('ATOMIC', 'COMPOSITE') NOT NULL DEFAULT 'ATOMIC' COMMENT '任务类型',
			config JSON DEFAULT NULL COMMENT '任务配置, 例如脚本, HTTP请求等',
			retry_policy JSON DEFAULT NULL COMMENT '重试策略配置',
			PRIMARY KEY (id),
			UNIQUE KEY uk_ (name),
			KEY idx_n_tt (name, task_type)
		) COMMENT='任务定义表: 可复用模板';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_def;
		`,
	},
	{
		CommitID: "basic",
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_def_relation (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			parent_id BIGINT NOT NULL,
			child_id BIGINT NOT NULL,
			ord INT DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY uniq_pi_o (parent_id, ord)
		) COMMENT='任务定义层级关系表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_def_relation;
		`,
	},
	{
		CommitID: "basic",
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_instance (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			task_def_id BIGINT NOT NULL COMMENT '关联的任务定义ID',
			parent_instance_id BIGINT DEFAULT NULL COMMENT '父实例ID',
			ord INT DEFAULT 0 COMMENT '同级实例顺序',
			status ENUM('PENDING','RUNNING','SUCCESS','FAILED','CANCELLED') DEFAULT 'PENDING' COMMENT '执行状态',
			result JSON DEFAULT NULL COMMENT '执行结果',
			err_msg TEXT DEFAULT NULL COMMENT '错误信息',
			FOREIGN KEY (task_def_id) REFERENCES task_def(id),
			FOREIGN KEY (parent_instance_id) REFERENCES task_instance(id) ON DELETE CASCADE,
			PRIMARY KEY (id),
			INDEX idx_pii_s (parent_instance_id, ord)
		) COMMENT='任务实例表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_instance;
		`,
	},
	{
		CommitID: "basic",
		UpSQL: `
   CREATE TABLE IF NOT EXISTS task_exec (
     id BIGINT AUTO_INCREMENT NOT NULL,
     gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
     gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
     task_instance_id BIGINT NOT NULL COMMENT '任务实例ID',
     gmt_start TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     gmt_end TIMESTAMP DEFAULT NULL,
     status ENUM('RUNNING','SUCCESS','FAILED') DEFAULT 'RUNNING',
     log TEXT COMMENT '执行日志',
     FOREIGN KEY (task_instance_id) REFERENCES task_instance(id) ON DELETE CASCADE,
     PRIMARY KEY (id),
     KEY idx_ti_s (task_instance_id, status)
   ) COMMENT='任务执行记录';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_exec;
		`,
	},
}
