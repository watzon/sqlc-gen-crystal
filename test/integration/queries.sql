-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: ListActiveUsers :many
SELECT * FROM users WHERE active = 1 ORDER BY name;

-- name: CreateUser :one
INSERT INTO users (email, name, age, active)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users 
SET name = ?, age = ?, active = ?
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: CountUsers :one
SELECT COUNT(*) as count FROM users;

-- name: GetUserStats :one
SELECT 
    COUNT(*) as total_users,
    COUNT(CASE WHEN active = 1 THEN 1 END) as active_users,
    AVG(age) as average_age
FROM users;

-- name: CreatePost :one
INSERT INTO posts (user_id, title, content, published, rating)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPost :one
SELECT 
    p.*,
    u.name as author_name,
    u.email as author_email
FROM posts p
JOIN users u ON u.id = p.user_id
WHERE p.id = ?;

-- name: ListPostsByUser :many
SELECT * FROM posts
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: UpdatePostViewCount :execrows
UPDATE posts 
SET view_count = view_count + 1
WHERE id = ?;

-- name: GetPopularPosts :many
SELECT 
    p.*,
    u.name as author_name
FROM posts p
JOIN users u ON u.id = p.user_id
WHERE p.published = 1
ORDER BY p.view_count DESC
LIMIT ?;

-- name: CreateTag :one
INSERT INTO tags (name) VALUES (?) RETURNING *;

-- name: AddPostTag :exec
INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?);

-- name: GetPostTags :many
SELECT t.* FROM tags t
JOIN post_tags pt ON pt.tag_id = t.id
WHERE pt.post_id = ?
ORDER BY t.name;

-- name: SearchPosts :many
SELECT * FROM posts
WHERE (title LIKE ? OR content LIKE ?)
AND published = 1
ORDER BY created_at DESC;

-- name: CountPostsByUser :many
SELECT user_id, COUNT(*) as post_count
FROM posts
GROUP BY user_id
ORDER BY user_id;

-- name: InsertPostBasic :exec
INSERT INTO posts (user_id, title, content, published)
VALUES (?, ?, ?, ?);

-- name: CreateUserReturnId :one
INSERT INTO users (email, name, age, active)
VALUES (?, ?, ?, ?)
RETURNING id;

-- name: BulkInsertUsers :copyfrom
INSERT INTO users (email, name, age, active) VALUES (?, ?, ?, ?);

-- name: DeactivateAllUsers :exec
UPDATE users SET active = 0;

-- name: UpdateUserName :exec  
UPDATE users SET name = ? WHERE id = ?;

-- name: DeleteUserById :exec
DELETE FROM users WHERE id = ?;

-- name: DeleteUsersByAge :execrows
DELETE FROM users WHERE age < ?;

-- name: DeleteAllInactivePosts :exec
DELETE FROM posts WHERE published = 0;

-- name: UpdateUserWithNamedArgs :exec
UPDATE users 
SET name = sqlc.arg(new_name), age = sqlc.arg(new_age)
WHERE id = sqlc.arg(user_id);

-- name: UpdateUserWithShorthand :exec
UPDATE users 
SET name = @new_name, age = @new_age
WHERE id = @user_id;

-- name: UpdateUserNullable :exec  
UPDATE users
SET
  name = coalesce(sqlc.narg('new_name'), name),
  age = coalesce(sqlc.narg('new_age'), age)
WHERE id = sqlc.arg('user_id');