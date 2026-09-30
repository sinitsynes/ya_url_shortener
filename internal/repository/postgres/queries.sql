-- name: CreateResource :one
INSERT INTO resource (original_url, short_url)
VALUES ($1, $2)
RETURNING *;

-- name: CreateBatch :many
INSERT INTO resource (original_url, short_url, correlation_id)
SELECT
    UNNEST(@original_urls::TEXT[]),
    UNNEST(@short_urls::TEXT[]),
    UNNEST(@correlation_ids::UUID[])
RETURNING *;

-- name: GetResourceByID :one
SELECT *
FROM resource
WHERE id = @id;

-- name: GetResourceByShortURL :one
SELECT *
FROM resource
WHERE short_url = @short_url;

-- name: GetResourceByOriginalURL :one
SELECT *
FROM resource
WHERE original_url = @original_url;
