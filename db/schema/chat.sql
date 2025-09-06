-- Chat messages table for storing all conversation records
CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    chat_id VARCHAR(128) NOT NULL COMMENT '会话ID',
    open_id VARCHAR(128) NOT NULL COMMENT '用户openid',
    username VARCHAR(128) COMMENT '用户名',
    role VARCHAR(32) NOT NULL COMMENT '角色: user/assistant',
    content TEXT NOT NULL COMMENT '消息内容',
    message_id VARCHAR(128) COMMENT '原始消息ID',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_chat_id (chat_id),
    INDEX idx_open_id (open_id),
    INDEX idx_created_at (created_at),
    INDEX idx_chat_role (chat_id, role),
    INDEX idx_created_chat (created_at, chat_id)
) COMMENT='对话消息记录表';

-- Chat memory table for long-term memory (summaries)
CREATE TABLE IF NOT EXISTS chat_memory (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    chat_id VARCHAR(128) NOT NULL COMMENT '会话ID',
    keyword VARCHAR(256) NOT NULL COMMENT '提取的关键词',
    summary TEXT NOT NULL COMMENT '对话摘要',
    start_time TIMESTAMP NOT NULL COMMENT '对话时间段开始',
    end_time TIMESTAMP NOT NULL COMMENT '对话时间段结束',
    message_count INT DEFAULT 0 COMMENT '消息数量',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_chat_id (chat_id),
    INDEX idx_keyword (keyword),
    INDEX idx_created_at (created_at),
    INDEX idx_chat_keyword (chat_id, keyword(64)),
    INDEX idx_time_range (start_time, end_time)
) COMMENT='对话长期记忆摘要表';
