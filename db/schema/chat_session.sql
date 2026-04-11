-- Chat Session: per-session dynamic state
CREATE TABLE IF NOT EXISTS chat_session (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(64) NOT NULL,
    summary TEXT COMMENT 'Running session summary',
    pending_tasks TEXT COMMENT 'JSON array of pending tasks',
    context TEXT COMMENT 'Important context from conversation',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_session_id (session_id)
);