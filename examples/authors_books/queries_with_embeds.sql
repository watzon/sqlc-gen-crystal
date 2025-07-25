-- Example queries demonstrating sqlc.embed() support

-- name: GetAuthorWithAllBooks :many
-- Get an author with all their books using LEFT JOIN
SELECT sqlc.embed(authors), sqlc.embed(books)
FROM authors
LEFT JOIN books ON books.author_id = authors.id
WHERE authors.id = $1
ORDER BY books.published DESC;

-- name: GetRecentBooksWithAuthors :many
-- Get recent books with their authors using INNER JOIN
SELECT sqlc.embed(books), sqlc.embed(authors)
FROM books
INNER JOIN authors ON books.author_id = authors.id
WHERE books.published > CURRENT_DATE - INTERVAL '1 year'
ORDER BY books.published DESC;

-- name: GetAuthorsWithBookStats :many
-- Get authors with aggregated book statistics
SELECT 
  sqlc.embed(authors),
  COUNT(books.id) as total_books,
  COUNT(CASE WHEN books.published > CURRENT_DATE - INTERVAL '1 year' THEN 1 END) as recent_books
FROM authors
LEFT JOIN books ON books.author_id = authors.id
GROUP BY authors.id, authors.name, authors.bio
ORDER BY total_books DESC;