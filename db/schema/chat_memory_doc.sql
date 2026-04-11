-- Memory Document Table for persistent session memory
CREATE TABLE IF NOT EXISTS chat_memory_doc (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(128) NOT NULL COMMENT '会话ID',
    content MEDIUMTEXT NOT NULL COMMENT '记忆文档内容(Markdown)',
    version INT NOT NULL DEFAULT 1 COMMENT '版本号',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_session_id (session_id),
    INDEX idx_updated_at (updated_at)
) COMMENT='会话记忆文档表';

-- Historical Chat Recall Table for storing recalled old conversations
CREATE TABLE IF NOT EXISTS chat_recall (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(128) NOT NULL COMMENT '会话ID',
    query VARCHAR(512) NOT NULL COMMENT '用户查询关键词',
    recalled_content TEXT NOT NULL COMMENT '召回的历史内容',
    relevance_score FLOAT COMMENT '相关性评分',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_session_id (session_id),
    INDEX idx_query (query(64)),
    INDEX idx_created_at (created_at)
) COMMENT='历史聊天召回记录表';