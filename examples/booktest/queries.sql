-- name: GetAuthor :one
SELECT * FROM authors WHERE id = $1;

-- name: ListAuthors :many
SELECT * FROM authors ORDER BY name;

-- name: CreateAuthor :one
INSERT INTO authors (name, bio) VALUES ($1, $2) RETURNING *;

-- name: UpdateAuthor :exec
UPDATE authors SET name = $2, bio = $3 WHERE id = $1;

-- name: DeleteAuthor :exec
DELETE FROM authors WHERE id = $1;

-- name: GetBook :one
SELECT b.*, a.name as author_name
FROM books b
JOIN authors a ON a.id = b.author_id
WHERE b.id = $1;

-- name: ListBooks :many
SELECT b.*, a.name as author_name
FROM books b
JOIN authors a ON a.id = b.author_id
ORDER BY b.created_at DESC;

-- name: ListBooksByAuthor :many
SELECT * FROM books
WHERE author_id = $1
ORDER BY published DESC;

-- name: CreateBook :one
INSERT INTO books (
  author_id, title, description, price, isbn, published
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateBookPrice :execrows
UPDATE books SET price = $2 WHERE id = $1;

-- name: CountBooksByAuthor :one
SELECT COUNT(*) FROM books WHERE author_id = $1;

-- name: GetAverageRating :one
SELECT AVG(rating)::FLOAT as avg_rating
FROM reviews
WHERE book_id = $1;

-- name: CreateReview :one
INSERT INTO reviews (book_id, reviewer, rating, comment)
VALUES ($1, $2, $3, $4)
RETURNING *;