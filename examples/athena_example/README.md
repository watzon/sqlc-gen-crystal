# Athena Blog API Example

A simple blog API built with [Athena Framework](https://athenaframework.org/) and [sqlc-gen-crystal](https://github.com/watzon/sqlc-gen-crystal), demonstrating how to create type-safe database interactions in a real-world web application.

## Features

- **Type-safe database queries** generated from SQL using sqlc-gen-crystal
- **REST API endpoints** built with Athena Framework
- **SQLite database** for simplicity and portability
- **JSON serialization** for API responses
- **Blog functionality** including users, posts, and tags

## API Endpoints

### Database Initialization
- `GET /init` - Initialize the database with schema

### Users
- `POST /users` - Create a new user
- `GET /users/:id` - Get user by ID

### Posts
- `GET /posts` - List all published posts (with pagination)
- `GET /posts/:slug` - Get post by slug
- `POST /posts` - Create a new post
- `POST /posts/:id/publish` - Publish a post
- `DELETE /posts/:id` - Delete a post

### Tags
- `GET /tags` - List all tags
- `POST /tags` - Create a new tag
- `GET /tags/:slug/posts` - Get posts by tag

## Getting Started

1. **Install dependencies:**
   ```bash
   shards install
   ```

2. **Generate database code from SQL:**
   ```bash
   sqlc generate
   ```

3. **Set up the database with sample data:**
   ```bash
   shards run setup_db
   ```

4. **Build and run the server:**
   ```bash
   shards run app
   ```

The database setup script will create a SQLite database with sample data including:
- A user account (john_doe)
- Two tags (Crystal, Web Development)
- Three blog posts (2 published, 1 draft)
- Post-tag associations

## Example Usage

### Create a user:
```bash
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password_hash": "hashed_password",
    "display_name": "John Doe",
    "bio": "A software developer"
  }'
```

### Create a post:
```bash
curl -X POST http://localhost:3000/posts \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "title": "My First Post",
    "slug": "my-first-post",
    "content": "This is the content of my first blog post...",
    "excerpt": "A brief excerpt",
    "published": true
  }'
```

### List posts:
```bash
curl http://localhost:3000/posts
```

### Create a tag:
```bash
curl -X POST http://localhost:3000/tags \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Crystal",
    "slug": "crystal",
    "description": "Posts about the Crystal programming language"
  }'
```
