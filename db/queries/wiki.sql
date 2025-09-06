-- name: CreateWiki :execresult
INSERT INTO wiki (title, content, keywords, content_hash, created_by)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  title = VALUES(title),
  content = VALUES(content),
  keywords = VALUES(keywords),
  gmt_modified = CURRENT_TIMESTAMP;

-- name: DeleteWikiByID :execresult
DELETE FROM wiki WHERE id = ? LIMIT 1;

-- name: GetWikiByID :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  keywords,
  content_hash,
  created_by
FROM wiki
WHERE id = ?
LIMIT 1;

-- name: GetWikiByHash :one
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  keywords,
  content_hash,
  created_by
FROM wiki
WHERE content_hash = ?
LIMIT 1;

-- name: ListWiki :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  keywords,
  content_hash,
  created_by
FROM wiki
ORDER BY gmt_modified DESC
LIMIT ? OFFSET ?;

-- name: SearchWiki :many
SELECT
  id,
  gmt_create,
  gmt_modified,
  title,
  content,
  keywords,
  content_hash,
  created_by
FROM wiki
WHERE MATCH(title, keywords, content) AGAINST(? IN NATURAL LANGUAGE MODE)
ORDER BY MATCH(title, keywords, content) AGAINST(? IN NATURAL LANGUAGE MODE) DESC
LIMIT ?;

-- name: CountWiki :one
SELECT count(*) ct FROM wiki;
