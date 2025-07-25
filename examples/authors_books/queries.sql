-- name: GetAuthor :one
SELECT * FROM authors
WHERE id = $1 LIMIT 1;

-- name: ListAuthors :many
SELECT * FROM authors
ORDER BY name;

-- name: CreateAuthor :one
INSERT INTO authors (
  name, bio
) VALUES (
  $1, $2
)
RETURNING *;

-- name: UpdateAuthor :exec
UPDATE authors
SET name = $2, bio = $3
WHERE id = $1;

-- name: DeleteAuthor :exec
DELETE FROM authors
WHERE id = $1;

-- name: GetBook :one
SELECT * FROM books
WHERE id = $1 LIMIT 1;

-- name: ListBooksByAuthor :many
SELECT * FROM books
WHERE author_id = $1
ORDER BY published DESC;

-- name: CreateBook :one
INSERT INTO books (
  author_id, title, isbn, published
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;