package psl

var userTables = []Change{
	{
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
		UpSQL: `
		CREATE TABLE IF NOT EXISTS atomic_task_def (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			name VARCHAR(128) NOT NULL COMMENT '原子任务名称',
			description TEXT COMMENT '任务描述',
			task_category ENUM('SCRIPT', 'HTTP_API', 'BUILTIN') NOT NULL DEFAULT 'SCRIPT' COMMENT '任务类别',
			script_type ENUM('SHELL', 'PYTHON', 'LUA', 'JAVASCRIPT', 'OTHER') DEFAULT NULL COMMENT '脚本类型',
			script_content LONGTEXT NOT NULL COMMENT '脚本内容',
			rollback_script_type ENUM('SHELL', 'PYTHON', 'LUA', 'JAVASCRIPT', 'OTHER') DEFAULT NULL COMMENT '回滚脚本类型',
			rollback_script_content LONGTEXT DEFAULT NULL COMMENT '回滚脚本内容',
			http_config JSON DEFAULT NULL COMMENT 'HTTP API 配置 (method, url, headers, body)',
			builtin_config JSON DEFAULT NULL COMMENT '内置任务配置',
			timeout INT DEFAULT 300 COMMENT '超时时间(秒)',
			retry_count INT DEFAULT 0 COMMENT '重试次数',
			retry_interval INT DEFAULT 5 COMMENT '重试间隔(秒)',
			is_rollback_supported BOOLEAN DEFAULT TRUE COMMENT '是否支持回滚',
			parameters JSON DEFAULT NULL COMMENT '参数定义 (schema)',
			output_schema JSON DEFAULT NULL COMMENT '输出 schema',
			env_vars JSON DEFAULT NULL COMMENT '环境变量',
			working_dir VARCHAR(512) DEFAULT NULL COMMENT '工作目录',
			status ENUM('ENABLED', 'DISABLED') DEFAULT 'ENABLED' COMMENT '状态',
			PRIMARY KEY (id),
			UNIQUE KEY uk_name (name),
			KEY idx_cat_st (task_category, script_type)
		) COMMENT='原子任务定义表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS atomic_task_def;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_workflow_def (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			name VARCHAR(128) NOT NULL COMMENT '工作流名称',
			description TEXT COMMENT '工作流描述',
			version INT DEFAULT 1 COMMENT '版本号',
			workflow_type ENUM('SEQUENTIAL', 'PARALLEL', 'DAG', 'CONDITIONAL') NOT NULL DEFAULT 'DAG' COMMENT '工作流类型',
			graph_config JSON NOT NULL COMMENT 'DAG 图配置: 节点和边',
			parameters JSON DEFAULT NULL COMMENT '输入参数定义',
			timeout INT DEFAULT 3600 COMMENT '整体超时时间(秒)',
			on_error_strategy ENUM('STOP', 'CONTINUE', 'ROLLBACK') DEFAULT 'STOP' COMMENT '错误处理策略',
			notification_config JSON DEFAULT NULL COMMENT '通知配置',
			status ENUM('DRAFT', 'ENABLED', 'DISABLED', 'DEPRECATED') DEFAULT 'DRAFT' COMMENT '状态',
			created_by VARCHAR(128) DEFAULT NULL COMMENT '创建人',
			PRIMARY KEY (id),
			UNIQUE KEY uk_name_version (name, version),
			KEY idx_status (status)
		) COMMENT='任务工作流定义表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_workflow_def;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_workflow_node (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			workflow_id BIGINT NOT NULL COMMENT '关联工作流定义ID',
			node_id VARCHAR(64) NOT NULL COMMENT '节点唯一标识',
			node_type ENUM('ATOMIC', 'SUB_WORKFLOW', 'CONDITION', 'PARALLEL_SPLIT', 'PARALLEL_JOIN', 'DELAY', 'MANUAL_APPROVAL') NOT NULL COMMENT '节点类型',
			display_name VARCHAR(128) NOT NULL COMMENT '节点显示名称',
			atomic_task_def_id BIGINT DEFAULT NULL COMMENT '引用的原子任务ID',
			sub_workflow_id BIGINT DEFAULT NULL COMMENT '引用的子工作流ID',
			condition_expr TEXT DEFAULT NULL COMMENT '条件表达式 (EL)',
			node_config JSON DEFAULT NULL COMMENT '节点额外配置',
			retry_policy JSON DEFAULT NULL COMMENT '节点级重试策略',
			timeout INT DEFAULT NULL COMMENT '节点超时时间',
			ord INT DEFAULT 0 COMMENT '执行顺序',
			PRIMARY KEY (id),
			UNIQUE KEY uk_wf_nid (workflow_id, node_id),
			KEY idx_wf_ord (workflow_id, ord)
		) COMMENT='工作流节点定义表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_workflow_node;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_workflow_edge (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			workflow_id BIGINT NOT NULL COMMENT '关联工作流定义ID',
			from_node_id VARCHAR(64) NOT NULL COMMENT '源节点ID',
			to_node_id VARCHAR(64) NOT NULL COMMENT '目标节点ID',
			edge_type ENUM('SEQUENTIAL', 'CONDITION_TRUE', 'CONDITION_FALSE', 'PARALLEL') DEFAULT 'SEQUENTIAL' COMMENT '边类型',
			condition_expr TEXT DEFAULT NULL COMMENT '条件表达式',
			PRIMARY KEY (id),
			UNIQUE KEY uk_wf_edge (workflow_id, from_node_id, to_node_id)
		) COMMENT='工作流边定义表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_workflow_edge;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_workflow_instance (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			workflow_def_id BIGINT NOT NULL COMMENT '关联工作流定义ID',
			workflow_def_version INT DEFAULT 1 COMMENT '工作流定义版本',
			trigger_type ENUM('MANUAL', 'SCHEDULED', 'API', 'EVENT', 'DEPENDENCY') NOT NULL DEFAULT 'MANUAL' COMMENT '触发类型',
			trigger_id VARCHAR(128) DEFAULT NULL COMMENT '触发源ID (调度ID/API请求ID等)',
			input_params JSON DEFAULT NULL COMMENT '输入参数',
			status ENUM('PENDING', 'RUNNING', 'PAUSED', 'SUCCESS', 'FAILED', 'CANCELLED', 'ROLLING_BACK') NOT NULL DEFAULT 'PENDING' COMMENT '执行状态',
			status_reason VARCHAR(512) DEFAULT NULL COMMENT '状态原因',
			current_node_id VARCHAR(64) DEFAULT NULL COMMENT '当前执行节点',
			execution_mode ENUM('SYNCHRONOUS', 'ASYNCHRONOUS') DEFAULT 'ASYNCHRONOUS' COMMENT '执行模式',
			priority INT DEFAULT 5 COMMENT '调度优先级 (1-10)',
			gmt_start TIMESTAMP DEFAULT NULL COMMENT '开始时间',
			gmt_end TIMESTAMP DEFAULT NULL COMMENT '结束时间',
			gmt_paused TIMESTAMP DEFAULT NULL COMMENT '暂停时间',
			total_nodes INT DEFAULT 0 COMMENT '总节点数',
			completed_nodes INT DEFAULT 0 COMMENT '已完成节点数',
			failed_nodes INT DEFAULT 0 COMMENT '失败节点数',
			result_summary JSON DEFAULT NULL COMMENT '执行结果汇总',
			error_info JSON DEFAULT NULL COMMENT '错误信息',
			created_by VARCHAR(128) DEFAULT NULL COMMENT '创建人',
			PRIMARY KEY (id),
			KEY idx_wf_def (workflow_def_id),
			KEY idx_status (status),
			KEY idx_priority (status, priority DESC)
		) COMMENT='工作流实例表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_workflow_instance;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_node_instance (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			workflow_instance_id BIGINT NOT NULL COMMENT '关联工作流实例ID',
			node_def_id BIGINT NOT NULL COMMENT '关联节点定义ID',
			node_id VARCHAR(64) NOT NULL COMMENT '节点标识',
			status ENUM('PENDING', 'RUNNING', 'SUCCESS', 'FAILED', 'SKIPPED', 'CANCELLED', 'ROLLING_BACK', 'ROLLED_BACK') NOT NULL DEFAULT 'PENDING' COMMENT '执行状态',
			status_reason VARCHAR(512) DEFAULT NULL COMMENT '状态原因',
			input_params JSON DEFAULT NULL COMMENT '输入参数',
			output_result JSON DEFAULT NULL COMMENT '输出结果',
			execution_log TEXT DEFAULT NULL COMMENT '执行日志',
			error_log TEXT DEFAULT NULL COMMENT '错误日志',
			retry_count INT DEFAULT 0 COMMENT '已重试次数',
			gmt_start TIMESTAMP DEFAULT NULL COMMENT '开始时间',
			gmt_end TIMESTAMP DEFAULT NULL COMMENT '结束时间',
			duration_ms INT DEFAULT NULL COMMENT '执行耗时(毫秒)',
			worker_id VARCHAR(64) DEFAULT NULL COMMENT '执行 worker ID',
			PRIMARY KEY (id),
			UNIQUE KEY uk_wi_nid (workflow_instance_id, node_id),
			KEY idx_wi_status (workflow_instance_id, status)
		) COMMENT='任务节点实例表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_node_instance;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS atomic_task_instance (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			node_instance_id BIGINT NOT NULL COMMENT '关联节点实例ID',
			atomic_task_def_id BIGINT NOT NULL COMMENT '关联原子任务定义ID',
			status ENUM('PENDING', 'RUNNING', 'SUCCESS', 'FAILED', 'ROLLED_BACK') NOT NULL DEFAULT 'PENDING' COMMENT '执行状态',
			input_params JSON DEFAULT NULL COMMENT '输入参数',
			output_result JSON DEFAULT NULL COMMENT '输出结果',
			execution_log TEXT DEFAULT NULL COMMENT '执行日志',
			error_log TEXT DEFAULT NULL COMMENT '错误日志',
			rollback_log TEXT DEFAULT NULL COMMENT '回滚日志',
			rollback_result JSON DEFAULT NULL COMMENT '回滚结果',
			timeout INT DEFAULT NULL COMMENT '超时时间',
			gmt_start TIMESTAMP DEFAULT NULL COMMENT '开始时间',
			gmt_end TIMESTAMP DEFAULT NULL COMMENT '结束时间',
			duration_ms INT DEFAULT NULL COMMENT '执行耗时(毫秒)',
			worker_id VARCHAR(64) DEFAULT NULL COMMENT '执行 worker ID',
			PRIMARY KEY (id),
			KEY idx_ni_status (node_instance_id, status),
			KEY idx_atd (atomic_task_def_id)
		) COMMENT='原子任务实例表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS atomic_task_instance;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_schedule (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			name VARCHAR(128) NOT NULL COMMENT '调度名称',
			description TEXT COMMENT '调度描述',
			workflow_def_id BIGINT NOT NULL COMMENT '关联工作流定义ID',
			schedule_type ENUM('CRON', 'INTERVAL', 'ONCE') NOT NULL COMMENT '调度类型',
			cron_expr VARCHAR(64) DEFAULT NULL COMMENT 'Cron 表达式',
			interval_seconds INT DEFAULT NULL COMMENT '间隔秒数',
			execute_at TIMESTAMP DEFAULT NULL COMMENT '单次执行时间',
			input_params JSON DEFAULT NULL COMMENT '默认输入参数',
			status ENUM('ENABLED', 'DISABLED', 'PAUSED') DEFAULT 'ENABLED' COMMENT '状态',
			last_execute_id BIGINT DEFAULT NULL COMMENT '最近执行的实例ID',
			next_execute_at TIMESTAMP DEFAULT NULL COMMENT '下次执行时间',
			created_by VARCHAR(128) DEFAULT NULL COMMENT '创建人',
			PRIMARY KEY (id),
			KEY idx_wf_status (workflow_def_id, status),
			KEY idx_next_exec (next_execute_at)
		) COMMENT='任务调度表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_schedule;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS task_execution_log (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			workflow_instance_id BIGINT NOT NULL COMMENT '关联工作流实例ID',
			node_instance_id BIGINT DEFAULT NULL COMMENT '关联节点实例ID',
			atomic_task_instance_id BIGINT DEFAULT NULL COMMENT '关联原子任务实例ID',
			log_level ENUM('DEBUG', 'INFO', 'WARN', 'ERROR') NOT NULL DEFAULT 'INFO' COMMENT '日志级别',
			log_type ENUM('EXECUTION', 'ROLLBACK', 'RETRY', 'SYSTEM') NOT NULL DEFAULT 'EXECUTION' COMMENT '日志类型',
			message TEXT NOT NULL COMMENT '日志内容',
			context JSON DEFAULT NULL COMMENT '上下文信息',
			PRIMARY KEY (id),
			KEY idx_wi (workflow_instance_id),
			KEY idx_wi_ni (workflow_instance_id, node_instance_id)
		) COMMENT='任务执行日志表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS task_execution_log;
		`,
	},
}
