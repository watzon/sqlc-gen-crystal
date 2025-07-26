# sqlc-gen-crystal

> [!CAUTION]
> Early development - expect breaking changes and rough edges. Please report issues!

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Getting Started](#getting-started)
  - [Real-world Usage](#real-world-usage)
- [Options](#options)
  - [Generated Files](#generated-files)
- [Query Annotations](#query-annotations)
- [Advanced Features](#advanced-features)
  - [Struct Deduplication](#struct-deduplication)
  - [JOIN Queries with sqlc.embed()](#join-queries-with-sqlcembed)
- [Supported Engines](#supported-engines)
- [Type Mappings](#type-mappings)
  - [PostgreSQL](#postgresql)
  - [MySQL](#mysql)
  - [SQLite](#sqlite)
- [Transactions](#transactions)
  - [Manual Transaction Handling](#manual-transaction-handling)
  - [Repository Transaction Support](#repository-transaction-support)
- [Roadmap](#roadmap)
- [Development](#development)
  - [Building](#building)
  - [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Installation

Install the plugin using Go:

```bash
go install github.com/watzon/sqlc-gen-crystal@latest
```

The `sqlc-gen-crystal` binary will be installed to your `$GOPATH/bin` directory. Make sure this directory is in your `$PATH`.

## Usage

```yaml
version: "2"
plugins:
  - name: crystal
    process:
      cmd: sqlc-gen-crystal
sql:
  - schema: "schema.sql"
    queries: "query.sql"
    engine: postgresql
    codegen:
      - out: src/db
        plugin: crystal
        options:
          module: "MyApp" # Note: 'module' instead of 'package'
          emit_json_tags: false
          emit_boolean_question_getters: true # Use `getter?` for boolean fields
          generate_connection_manager: true
          generate_repositories: true
```

## Getting Started

Install sqlc following the [official documentation](https://docs.sqlc.dev/en/latest/overview/install.html).

Create a `schema.sql`:

```sql
CREATE TABLE authors (
  id   BIGSERIAL PRIMARY KEY,
  name text      NOT NULL,
  bio  text
);
```

Create a `query.sql`:

```sql
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

-- name: DeleteAuthor :exec
DELETE FROM authors
WHERE id = $1;
```

Generate Crystal code:

```shell
$ sqlc generate
```

Use the generated code:

```crystal
require "db"
require "pg"
require "./src/db"

DB.open("postgres://localhost/myapp") do |db|
  queries = MyApp::Queries.new(db)

  # Create an author
  author = queries.create_author("John Doe", "A great writer")

  # List all authors
  authors = queries.list_authors
  authors.each do |author|
    puts "#{author.name}: #{author.bio || "No bio"}"
  end

  # Delete an author
  queries.delete_author(author.id) if author
end
```

### Real-world Usage

With the `generate_connection_manager` and `generate_repositories` options enabled (as shown in the configuration above), you get a clean, ready-to-use API:

```crystal
require "./src/db/database"

# Simple repository usage
author = MyApp::AuthorsRepository.create(name: "Jane Doe", bio: "A prolific writer")
all_authors = MyApp::AuthorsRepository.all

# With transactions
MyApp::AuthorsRepository.transaction do |repo|
  author1 = repo.create(name: "Author 1", bio: "First author")
  author2 = repo.create(name: "Author 2", bio: "Second author")
  # Both inserts succeed or both fail
end

# Direct database access when needed
MyApp::Database.connection.exec("VACUUM ANALYZE authors")
```

## Options

| Option                         | Default    | Description                                              |
| ------------------------------ | ---------- | -------------------------------------------------------- |
| module                         | (required) | Crystal module name (supports nested: "MyApp::Database") |
| emit_json_tags                 | false      | Add JSON::Serializable annotations to structs            |
| emit_yaml_tags                 | false      | Add YAML::Serializable annotations to structs            |
| emit_db_tags                   | true       | Add DB::Serializable annotations to structs              |
| emit_msgpack_tags              | false      | Add MessagePack::Serializable annotations to structs     |
| emit_boolean_question_getters  | false      | Generate `getter?` methods for boolean fields            |
| generate_connection_manager    | false      | Generate a Database class for connection management      |
| generate_repositories          | false      | Generate repository classes for each table               |

### Generated Files

The plugin always generates:

- `database.cr` - Entry point that requires all other generated files
- `models.cr` - Crystal structs for database tables
- `queries.cr` - Type-safe query methods

With `generate_connection_manager: true`, the `database.cr` file also includes:

- Singleton database connection manager
- Transaction support methods

With `generate_repositories: true`, additional files are generated:

- `repositories/[table]_repository.cr` - Repository class for each table
- Repository methods that wrap the underlying queries
- Transaction support at the repository level

## Query Annotations

- `:one` - Returns 0 or 1 row as `T?`
- `:many` - Returns 0 to n rows as `Array(T)`
- `:exec` - Executes query without returning rows
- `:execrows` - Returns number of affected rows as `Int64`
- `:execresult` - Returns `DB::ExecResult`

## Advanced Features

### Struct Deduplication

The plugin automatically detects when multiple queries return the same set of columns and reuses the same struct instead of generating duplicates. This reduces code size and improves maintainability.

```sql
-- Both queries return the same columns
-- name: GetAuthor :one
SELECT id, name, bio FROM authors WHERE id = $1;

-- name: CreateAuthor :one
INSERT INTO authors (name, bio) VALUES ($1, $2)
RETURNING id, name, bio;
```

Generated code will use the same `Author` struct for both queries instead of creating `GetAuthorRow` and `CreateAuthorRow`.

### JOIN Queries with sqlc.embed()

The plugin supports sqlc's `embed()` function for JOIN queries, creating nested structs that maintain table relationships:

```sql
-- name: GetAuthorWithBooks :many
SELECT sqlc.embed(authors), sqlc.embed(books)
FROM authors
LEFT JOIN books ON books.author_id = authors.id
WHERE authors.id = $1;

-- name: GetBookWithAuthor :one
SELECT sqlc.embed(books), sqlc.embed(authors)
FROM books
INNER JOIN authors ON books.author_id = authors.id
WHERE books.id = $1;

-- name: GetAuthorsWithStats :many
SELECT
  sqlc.embed(authors),
  COUNT(books.id) as book_count,
  MAX(books.published) as latest_book
FROM authors
LEFT JOIN books ON books.author_id = authors.id
GROUP BY authors.id, authors.name, authors.bio;
```

This generates clean, nested structs:

```crystal
struct GetAuthorWithBooksRow
  include DB::Serializable
  getter author : Author      # Non-nullable (from FROM table)
  getter book : Book?         # Nullable (from LEFT JOIN)
end

struct GetBookWithAuthorRow
  include DB::Serializable
  getter book : Book          # Non-nullable (INNER JOIN)
  getter author : Author      # Non-nullable (INNER JOIN)
end

struct GetAuthorsWithStatsRow
  include DB::Serializable
  getter author : Author      # Embedded author struct
  getter book_count : Int64   # Aggregate column
  getter latest_book : Time?  # Aggregate column (nullable)
end
```

Usage:

```crystal
# Fetch author with their books
rows = queries.get_author_with_books(author_id)
rows.each do |row|
  puts "Author: #{row.author.name}"
  if book = row.book
    puts "  Book: #{book.title}"
  else
    puts "  No books"
  end
end

# Fetch authors with statistics
stats = queries.get_authors_with_stats
stats.each do |row|
  puts "#{row.author.name}: #{row.book_count} books"
  if latest = row.latest_book
    puts "  Latest: #{latest}"
  end
end
```

The plugin automatically handles:

- Nullable embedded structs for LEFT/RIGHT JOINs
- Non-nullable embedded structs for INNER JOINs
- Mixed queries with both embedded tables and aggregate columns
- Proper type safety throughout

## Supported Engines

- PostgreSQL via [crystal-pg](https://github.com/will/crystal-pg)
- MySQL via [crystal-mysql](https://github.com/crystal-lang/crystal-mysql)
- SQLite3 via [crystal-sqlite3](https://github.com/crystal-lang/crystal-sqlite3)

## Type Mappings

### PostgreSQL

| PostgreSQL Type          | Crystal Type | Nullable Crystal Type |
| ------------------------ | ------------ | --------------------- |
| bigint, int8             | Int64        | Int64?                |
| integer, int4, int       | Int32        | Int32?                |
| smallint, int2           | Int16        | Int16?                |
| numeric, decimal         | BigDecimal   | BigDecimal?           |
| real, float4             | Float32      | Float32?              |
| double precision, float8 | Float64      | Float64?              |
| boolean, bool            | Bool         | Bool?                 |
| text, varchar, char      | String       | String?               |
| bytea                    | Bytes        | Bytes?                |
| timestamp, timestamptz   | Time         | Time?                 |
| date                     | Time         | Time?                 |
| json, jsonb              | JSON::Any    | JSON::Any?            |
| uuid                     | UUID         | UUID?                 |

### MySQL

| MySQL Type          | Crystal Type | Nullable Crystal Type |
| ------------------- | ------------ | --------------------- |
| bigint              | Int64        | Int64?                |
| int, integer        | Int32        | Int32?                |
| smallint            | Int16        | Int16?                |
| tinyint             | Int8         | Int8?                 |
| decimal, numeric    | BigDecimal   | BigDecimal?           |
| float               | Float32      | Float32?              |
| double              | Float64      | Float64?              |
| boolean, bool       | Bool         | Bool?                 |
| varchar, text, char | String       | String?               |
| blob, binary        | Bytes        | Bytes?                |
| datetime, timestamp | Time         | Time?                 |
| date                | Time         | Time?                 |
| json                | JSON::Any    | JSON::Any?            |

### SQLite

| SQLite Type         | Crystal Type | Nullable Crystal Type |
| ------------------- | ------------ | --------------------- |
| integer, int        | Int64        | Int64?                |
| real, float         | Float64      | Float64?              |
| text, varchar       | String       | String?               |
| blob                | Bytes        | Bytes?                |
| boolean, bool       | Bool         | Bool?                 |
| datetime, timestamp | Time         | Time?                 |

## Transactions

### Manual Transaction Handling

```crystal
DB.transaction do |tx|
  queries = MyApp::Queries.new(tx.connection)

  author = queries.create_author("Jane Doe", "Another writer")
  queries.create_post(author.id, "My First Post", "Content here...")

  # Automatically commits on success, rolls back on exception
end
```

### Repository Transaction Support

When using `generate_repositories: true`, repositories automatically support transactions:

```crystal
# Single repository transaction
MyApp::AuthorRepository.transaction do |author_repo|
  author = author_repo.create("Jane Doe", "A writer")
  author_repo.update(author.id, "Jane Smith", "Updated bio")
  # Automatically commits on success, rolls back on exception1
end

# Multiple repositories in same transaction
MyApp::Database.transaction do |tx_queries|
  author_repo = MyApp::AuthorRepository.with_transaction(tx_queries)
  book_repo = MyApp::BookRepository.with_transaction(tx_queries)

  author = author_repo.create("Jane Doe", "A writer")
  book = book_repo.create(author.id, "Her First Book", "978-1234567890")
  # All operations in same transaction
end
```

## Roadmap

### Core SQLC Features (Completed)

- [x] **Query Generation**
  - [x] `:one` queries (return single optional result)
  - [x] `:many` queries (return array of results)
  - [x] `:exec` queries (no return value)
  - [x] `:execrows` queries (return affected row count)
  - [x] `:execresult` queries (return DB::ExecResult)

- [x] **SQL Operations**
  - [x] SELECT queries with complex WHERE clauses
  - [x] INSERT queries with RETURNING support
  - [x] UPDATE queries with parameters
  - [x] DELETE queries with parameters
  - [x] Aggregate functions (COUNT, AVG, etc.)
  - [x] GROUP BY and ORDER BY clauses

- [x] **Type System**
  - [x] PostgreSQL type mappings
  - [x] MySQL type mappings
  - [x] SQLite type mappings
  - [x] Nullable types with Crystal union syntax
  - [x] Array parameter handling for IN clauses

- [x] **Code Generation**
  - [x] Struct generation with DB::Serializable
  - [x] Query method generation
  - [x] Parameter handling with optional defaults
  - [x] JSON/YAML serialization support
  - [x] Struct deduplication for identical column sets

- [x] **Advanced Features**
  - [x] JOIN queries with sqlc.embed() support
  - [x] Nested struct generation for embeds
  - [x] Connection manager generation
  - [x] Repository pattern generation
  - [x] Transaction support

### SQLC Features (Partial/In Progress)

- [x] **Bulk Operations**
  - [x] `:copyfrom` placeholder (returns stub with TODO)
  - [ ] Full bulk insert implementation for Crystal drivers

### Missing SQLC Features

- [ ] **Prepared Statements**
  - [ ] `emit_prepared_queries` configuration option
  - [ ] `Prepare()` method generation
  - [ ] `WithTx()` transaction support for prepared statements
- [ ] **Type Overrides**
  - [ ] Custom type mappings via overrides configuration
  - [ ] Database type overrides (db_type)
  - [ ] Column-specific overrides
- [ ] **Field and Struct Renaming**
  - [ ] Custom field name mappings via rename configuration
  - [ ] Table struct name customization
  - [ ] Column name customization
  - [ ] Prepared statement lifecycle management
  - [ ] Performance optimizations for repeated queries

- [ ] **Advanced Query Features**
  - [ ] Dynamic SQL support
  - [ ] Custom scalar types
  - [ ] Enum generation from database constraints
  - [ ] Custom type mappings

- [ ] **Configuration Options**
  - [ ] `emit_exact_table_names` support
  - [ ] `emit_empty_slices` support
  - [ ] Custom naming conventions
  - [ ] Override built-in type mappings

## Development

### Building

```bash
# Install dependencies
go mod download

# Build the plugin
go build -o bin/sqlc-gen-crystal

# Run tests
go test ./...
make test-integration
```

### Testing

The project includes comprehensive test coverage:

- Unit tests for type mappings and code generation
- Integration tests with real Crystal code compilation
- Support for PostgreSQL, MySQL, and SQLite

## Contributing

Pull requests welcome! Please:

- Add tests for new functionality
- Update documentation as needed
- Follow existing code style

## License

MIT - See [LICENSE](LICENSE) for details.
