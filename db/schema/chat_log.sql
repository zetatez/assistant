-- Chat Operation Log
CREATE TABLE IF NOT EXISTS chat_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(64) NOT NULL,
    action VARCHAR(32) NOT NULL COMMENT 'ingest/query/lint/update',
    detail TEXT COMMENT 'JSON format detail',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_session_id (session_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at)
);