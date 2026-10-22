-- name: CreateResource :one
INSERT INTO resource (original_url, short_url, user_id)
VALUES (@original_url, @short_url, @user_id)
RETURNING id, original_url, short_url, correlation_id, user_id;

-- name: CreateBatch :many
INSERT INTO resource (original_url, short_url, correlation_id, user_id)
SELECT
    UNNEST(@original_urls::TEXT[]),
    UNNEST(@short_urls::TEXT[]),
    UNNEST(@correlation_ids::TEXT[]),
    UNNEST(@user_ids::UUID[])
RETURNING id, original_url, short_url, correlation_id, user_id;

-- name: GetResourceByID :one
SELECT id, original_url, short_url, correlation_id, user_id
FROM resource
WHERE id = @id;

-- name: GetResourceByShortURL :one
SELECT id, original_url, short_url, correlation_id, user_id, is_deleted
FROM resource
WHERE short_url = @short_url;

-- name: GetResourceByOriginalURL :one
SELECT id, original_url, short_url, correlation_id, user_id
FROM resource
WHERE original_url = @original_url;

-- name: GetResourcesByUserID :many
SELECT id, original_url, short_url, correlation_id, user_id
FROM resource
WHERE user_id = @user_id;

-- name: SoftDeleteResources :exec
UPDATE resource AS r
SET is_deleted = TRUE
FROM (
    SELECT
        UNNEST(@user_ids::UUID[])   AS user_id,
        UNNEST(@short_urls::TEXT[]) AS short_url
) AS v
WHERE r.user_id = v.user_id
  AND r.short_url = v.short_url;
