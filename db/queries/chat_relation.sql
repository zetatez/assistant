-- name: CreateChatRelation :execresult
INSERT INTO chat_relations (session_id, from_entity_id, to_entity_id, relation_type, context)
VALUES (?, ?, ?, ?, ?);

-- name: GetChatRelations :many
SELECT id, session_id, from_entity_id, to_entity_id, relation_type, context, created_at
FROM chat_relations
WHERE session_id = ?
ORDER BY created_at DESC;

-- name: GetEntityRelations :many
SELECT r.id, r.session_id, r.from_entity_id, r.to_entity_id, r.relation_type, r.context, r.created_at
FROM chat_relations r
WHERE r.from_entity_id = ? OR r.to_entity_id = ?
ORDER BY relation_type;

-- name: SearchChatRelations :many
SELECT r.id, r.session_id, r.from_entity_id, r.to_entity_id, r.relation_type, r.context, r.created_at
FROM chat_relations r
JOIN chat_entities e ON r.from_entity_id = e.id
WHERE e.entity_name LIKE ? OR r.relation_type LIKE ?
LIMIT ?;

-- name: DeleteChatRelationsByChatID :execresult
DELETE FROM chat_relations WHERE session_id = ?;

-- name: DeleteOldChatRelations :execresult
DELETE FROM chat_relations WHERE created_at < ?;