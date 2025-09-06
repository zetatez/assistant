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
			chat_id VARCHAR(128) NOT NULL COMMENT '会话ID',
			open_id VARCHAR(128) NOT NULL COMMENT '用户openid',
			username VARCHAR(128) COMMENT '用户名',
			role VARCHAR(32) NOT NULL COMMENT '角色: user/assistant',
			content TEXT NOT NULL COMMENT '消息内容',
			message_id VARCHAR(128) COMMENT '原始消息ID',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			PRIMARY KEY (id),
			INDEX idx_chat_id (chat_id),
			INDEX idx_open_id (open_id),
			INDEX idx_created_at (created_at),
			INDEX idx_chat_role (chat_id, role),
			INDEX idx_created_chat (created_at, chat_id)
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
			chat_id VARCHAR(128) NOT NULL COMMENT '会话ID',
			keyword VARCHAR(256) NOT NULL COMMENT '提取的关键词',
			summary TEXT NOT NULL COMMENT '对话摘要',
			start_time TIMESTAMP NOT NULL COMMENT '对话时间段开始',
			end_time TIMESTAMP NOT NULL COMMENT '对话时间段结束',
			message_count INT DEFAULT 0 COMMENT '消息数量',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			PRIMARY KEY (id),
			INDEX idx_chat_id (chat_id),
			INDEX idx_keyword (keyword),
			INDEX idx_created_at (created_at),
			INDEX idx_chat_keyword (chat_id, keyword(64)),
			INDEX idx_time_range (start_time, end_time)
		) COMMENT='对话长期记忆摘要表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS chat_memory;
		`,
	},
	{
		UpSQL: `
		CREATE TABLE IF NOT EXISTS wiki (
			id BIGINT AUTO_INCREMENT NOT NULL,
			gmt_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			gmt_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			title VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文档标题',
			content LONGTEXT NOT NULL COMMENT 'Markdown 内容',
			keywords VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '搜索关键词',
			content_hash VARCHAR(64) NOT NULL DEFAULT '' COMMENT '内容 MD5 哈希',
			created_by VARCHAR(128) NOT NULL DEFAULT '' COMMENT '创建者',
			PRIMARY KEY (id),
			UNIQUE KEY uk_hash (content_hash),
			FULLTEXT KEY ft_search (title, keywords, content)
		) COMMENT='知识库文档表';
		`,
		DownSQL: `
		DROP TABLE IF EXISTS wiki;
		`,
	},
}
