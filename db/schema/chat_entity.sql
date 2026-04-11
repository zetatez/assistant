-- Chat Entities: extracted entities from conversations
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
);