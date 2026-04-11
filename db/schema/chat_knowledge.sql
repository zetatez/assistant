-- Chat Knowledge: structured knowledge pages
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
);