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
RETURNING correlation_id, short_url;

-- name: GetResourceByID :one
SELECT *
FROM resource
WHERE id = @id;

-- name: GetResourceByURL :one
SELECT *
FROM resource
WHERE short_url = @short_url;
