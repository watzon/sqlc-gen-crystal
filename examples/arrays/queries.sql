-- Basic array parameter queries for PostgreSQL

-- name: ListAuthorsByIDs :many
SELECT * FROM authors
WHERE id = ANY($1::int[]);

-- name: ListAuthorsByBirthYears :many
SELECT * FROM authors  
WHERE birth_year = ANY($1::int[])
ORDER BY name;

-- name: ListBooksByTags :many
SELECT * FROM books
WHERE tags && $1::text[]
ORDER BY published DESC;

-- name: ListBooksWithAnyTag :many
SELECT * FROM books
WHERE tags @> ARRAY[$1::text]
ORDER BY published DESC;

-- name: UpdateBookTags :exec
UPDATE books
SET tags = $2::text[]
WHERE id = $1;

-- Array queries for MySQL/SQLite using sqlc.slice()

-- name: ListAuthorsByIDsMySQL :many
SELECT * FROM authors
WHERE id IN (sqlc.slice('ids'));

-- name: ListBooksByAuthorIDs :many
SELECT * FROM books
WHERE author_id IN (sqlc.slice('author_ids'))
ORDER BY published DESC;

-- name: DeleteAuthorsByIDs :execrows
DELETE FROM authors
WHERE id IN (sqlc.slice('ids'));

-- Single column query with array param
-- name: GetAuthorNames :many
SELECT name FROM authors
WHERE id = ANY($1::int[])
ORDER BY name;

-- Mixed parameter types
-- name: SearchBooksWithTags :many
SELECT * FROM books
WHERE author_id = $1
  AND tags && $2::text[]
  AND published >= $3
ORDER BY published DESC;