-- Chat Relations: relationships between entities
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
);