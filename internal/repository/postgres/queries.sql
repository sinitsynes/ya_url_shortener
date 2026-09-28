-- name: CreateResource :one
INSERT INTO resource (original_url, shortened_url)
VALUES ($1, $2)
RETURNING *;

-- name: GetResourceByID :one
SELECT *
FROM resource
WHERE id = @id;

-- name: GetResourceByURL :one
SELECT *
FROM resource
WHERE shortened_url = @shortened_url;
