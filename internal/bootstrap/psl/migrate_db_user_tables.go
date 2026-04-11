package psl

var userTables = []UpDownSQL{
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS sys_user (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			user_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'user login name',
			password VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'user password hash',
			email VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'user email address',
			is_internal TINYINT NOT NULL DEFAULT 0 COMMENT 'internal user: 1=yes,0=no',
			PRIMARY KEY (id),
			UNIQUE KEY uk_un (user_name),
			UNIQUE KEY uk_e (email)
		) COMMENT='用户表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS sys_user;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS sys_server (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			idc VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'server idc location',
			svr_ip VARCHAR(32) NOT NULL DEFAULT '127.1' COMMENT 'server ip address',
			svr_status VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'server running status',
			cpu_usage FLOAT NOT NULL DEFAULT 0 COMMENT 'cpu usage percent',
			mem_usage FLOAT NOT NULL DEFAULT 0 COMMENT 'memory usage percent',
			PRIMARY KEY (id),
			UNIQUE KEY uk_si (svr_ip)
		) COMMENT='服务器表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS sys_server;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS chat_messages (
			id BIGINT AUTO_INCREMENT NOT NULL,
			session_id VARCHAR(128) NOT NULL COMMENT '会话ID',
			open_id VARCHAR(128) NOT NULL COMMENT '用户openid',
			username VARCHAR(128) COMMENT '用户名',
			role VARCHAR(32) NOT NULL COMMENT '角色: user/assistant',
			content TEXT NOT NULL COMMENT '消息内容',
			message_id VARCHAR(128) COMMENT '原始消息ID',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			PRIMARY KEY (id),
			INDEX idx_session_id (session_id),
			INDEX idx_open_id (open_id),
			INDEX idx_created_at (created_at),
			INDEX idx_session_role (session_id, role),
			INDEX idx_created_chat (created_at, session_id)
		) COMMENT='对话消息记录表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS chat_messages;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS chat_memory (
			id BIGINT AUTO_INCREMENT NOT NULL,
			session_id VARCHAR(128) NOT NULL COMMENT '会话ID',
			keyword VARCHAR(256) NOT NULL COMMENT '提取的关键词',
			summary TEXT NOT NULL COMMENT 'Historical conversation summary for long-term retrieval',
			memory_type VARCHAR(32) NOT NULL DEFAULT 'historical' COMMENT 'memory type: historical (historical memories, searchable)',
			start_time TIMESTAMP NOT NULL COMMENT '对话时间段开始',
			end_time TIMESTAMP NOT NULL COMMENT '对话时间段结束',
			message_count INT DEFAULT 0 COMMENT '消息数量',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			PRIMARY KEY (id),
			INDEX idx_session_id (session_id),
			INDEX idx_keyword (keyword),
			INDEX idx_created_at (created_at),
			INDEX idx_session_keyword (session_id, keyword(64)),
			INDEX idx_time_range (start_time, end_time),
			INDEX idx_session_type (session_id, memory_type)
		) COMMENT='Historical long-term memories - periodic snapshots for retrieval';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS chat_memory;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS chat_entities (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			session_id VARCHAR(64) NOT NULL,
			entity_type VARCHAR(32) NOT NULL COMMENT 'person/topic/concept/event',
			entity_name VARCHAR(256) NOT NULL,
			description TEXT COMMENT 'LLM generated description',
			keywords VARCHAR(512),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_session_id (session_id),
			INDEX idx_entity_name (entity_name(128)),
			INDEX idx_entity_type (entity_type)
		) COMMENT='对话实体提取表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS chat_entities;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS chat_relations (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			session_id VARCHAR(64) NOT NULL,
			from_entity_id BIGINT NOT NULL,
			to_entity_id BIGINT NOT NULL,
			relation_type VARCHAR(64) NOT NULL COMMENT 'related_to/depends_on/contradicts/part_of',
			context TEXT COMMENT 'context of this relation',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_session_id (session_id),
			INDEX idx_from_entity (from_entity_id),
			INDEX idx_to_entity (to_entity_id),
			INDEX idx_relation_type (relation_type),
			FOREIGN KEY (from_entity_id) REFERENCES chat_entities(id) ON DELETE CASCADE,
			FOREIGN KEY (to_entity_id) REFERENCES chat_entities(id) ON DELETE CASCADE
		) COMMENT='实体关系表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS chat_relations;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS chat_knowledge (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			session_id VARCHAR(64) NOT NULL,
			entity_id BIGINT COMMENT 'linked entity',
			title VARCHAR(256) NOT NULL,
			content TEXT NOT NULL COMMENT 'Markdown format content',
			source_messages TEXT COMMENT 'source message IDs, comma separated',
			version INT DEFAULT 1,
			is_draft BOOLEAN DEFAULT FALSE COMMENT 'pending LLM confirmation',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_session_id (session_id),
			INDEX idx_entity_id (entity_id),
			INDEX idx_title (title(128)),
			INDEX idx_updated_at (updated_at)
		) COMMENT='知识页面表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS chat_knowledge;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS chat_log (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			session_id VARCHAR(64) NOT NULL,
			action VARCHAR(32) NOT NULL COMMENT 'ingest/query/lint/update',
			detail TEXT COMMENT 'JSON format detail',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_session_id (session_id),
			INDEX idx_action (action),
			INDEX idx_created_at (created_at)
		) COMMENT='操作日志表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS chat_log;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS chat_session (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			session_id VARCHAR(64) NOT NULL,
			summary TEXT COMMENT 'Running session summary - current conversation understanding',
			pending_tasks TEXT COMMENT 'JSON array of pending tasks',
			context TEXT COMMENT 'Important context from conversation',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_session_id (session_id)
		) COMMENT='Current session state - updated on every message';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS chat_session;
		`,
	},
}
