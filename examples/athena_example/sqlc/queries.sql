-- User queries
-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, display_name, bio)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: UpdateUser :exec
UPDATE users 
SET display_name = ?, bio = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- Post queries
-- name: CreatePost :one
INSERT INTO posts (user_id, title, slug, content, excerpt, published, published_at)
VALUES (?, ?, ?, ?, ?, ?, CASE WHEN sqlc.arg(set_published_at) THEN CURRENT_TIMESTAMP ELSE NULL END)
RETURNING *;

-- name: GetPost :one
SELECT 
    p.*,
    u.username as author_username,
    u.display_name as author_display_name
FROM posts p
JOIN users u ON u.id = p.user_id
WHERE p.id = ?;

-- name: GetPostBySlug :one
SELECT 
    p.*,
    u.username as author_username,
    u.display_name as author_display_name
FROM posts p
JOIN users u ON u.id = p.user_id
WHERE p.slug = ?;

-- name: ListPosts :many
SELECT 
    p.*,
    u.username as author_username,
    u.display_name as author_display_name
FROM posts p
JOIN users u ON u.id = p.user_id
WHERE p.published = TRUE
ORDER BY p.published_at DESC
LIMIT ? OFFSET ?;

-- name: ListPostsByUser :many
SELECT * FROM posts
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: UpdatePost :exec
UPDATE posts 
SET title = ?, slug = ?, content = ?, excerpt = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?;

-- name: PublishPost :exec
UPDATE posts 
SET published = TRUE, published_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?;

-- name: UnpublishPost :exec
UPDATE posts 
SET published = FALSE, published_at = NULL, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = ? AND user_id = ?;

-- Tag queries
-- name: CreateTag :one
INSERT INTO tags (name, slug, description)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetTag :one
SELECT * FROM tags WHERE id = ?;

-- name: GetTagBySlug :one
SELECT * FROM tags WHERE slug = ?;

-- name: ListTags :many
SELECT * FROM tags ORDER BY name;

-- name: UpdateTag :exec
UPDATE tags SET name = ?, slug = ?, description = ? WHERE id = ?;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = ?;

-- Post-Tag relationship queries
-- name: AddPostTag :exec
INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?);

-- name: RemovePostTag :exec
DELETE FROM post_tags WHERE post_id = ? AND tag_id = ?;

-- name: GetPostTags :many
SELECT t.* FROM tags t
JOIN post_tags pt ON pt.tag_id = t.id
WHERE pt.post_id = ?
ORDER BY t.name;

-- name: GetPostsByTag :many
SELECT 
    p.*,
    u.username as author_username,
    u.display_name as author_display_name
FROM posts p
JOIN users u ON u.id = p.user_id
JOIN post_tags pt ON pt.post_id = p.id
WHERE pt.tag_id = ? AND p.published = TRUE
ORDER BY p.published_at DESC
LIMIT ? OFFSET ?;

-- name: CountPosts :one
SELECT COUNT(*) as count FROM posts WHERE published = TRUE;