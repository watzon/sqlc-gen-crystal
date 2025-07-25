-- name: ListAuthorsByIDs :many
SELECT id, name, bio FROM authors
WHERE id IN (sqlc.slice('ids'));

-- name: ListBooksByAuthorIDs :many
SELECT id, author_id, title, isbn FROM books
WHERE author_id IN (sqlc.slice('author_ids'))
ORDER BY title;

-- name: DeleteAuthorsByIDs :execrows
DELETE FROM authors
WHERE id IN (sqlc.slice('ids'));

-- name: GetAuthorNames :many
SELECT name FROM authors
WHERE id IN (sqlc.slice('ids'))
ORDER BY name;