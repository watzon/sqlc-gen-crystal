# Authors & Books Example

This example demonstrates the connection manager and repository features of sqlc-gen-crystal.

## Features Demonstrated

- **Connection Manager**: Singleton database connection management
- **Repository Pattern**: Clean repository classes for each table
- **Smart Method Names**: Simplified method names (find, all, create, update, delete)
- **Multiple Tables**: Shows how queries are grouped by table

## Usage

```bash
# Generate the Crystal code
sqlc generate

# The generated files will be in src/db/
```

## Generated Structure

```
src/db/
├── models.cr           # Data structs
├── queries.cr          # Raw query methods
├── database.cr         # Connection manager
└── repositories/
    ├── authors_repository.cr
    └── books_repository.cr
```

## Example Usage

### Normal Usage (Auto-commit)

```crystal
require "./src/db"

# Using repositories with auto-commit
author_repo = TestApp::AuthorsRepository.new
book_repo = TestApp::BooksRepository.new

# Create an author
author = author_repo.create("Jane Doe", "A great writer")

# Create a book
book = book_repo.create(author.id.to_s, "Crystal Programming", "978-1234567890", Time.utc)

# Find books by author
books = book_repo.by_author(author.id.to_s)
```

### Transaction Usage - Method 1: Repository.transaction

```crystal
require "./src/db"

# Single repository with transaction
TestApp::AuthorsRepository.transaction do |author_repo|
  # All operations within this block are in the same transaction
  author = author_repo.create("Jane Doe", "A great writer")
  author_repo.update(author.id, "Jane Smith", "Updated bio")
  # Automatically commits on success, rolls back on exception
end
```

### Transaction Usage - Method 2: Database.transaction

```crystal
require "./src/db"

# Multiple repositories in same transaction
TestApp::Database.transaction do |tx_queries|
  author_repo = TestApp::AuthorsRepository.with_transaction(tx_queries)
  book_repo = TestApp::BooksRepository.with_transaction(tx_queries)
  
  # Create author and book in same transaction
  author = author_repo.create("Jane Doe", "A great writer")
  if author
    book = book_repo.create(author.id.to_s, "Crystal Programming", "978-1234567890", Time.utc)
    puts "Created author #{author.name} and book #{book.try(&.title)}"
  end
  # Automatically commits on success, rolls back on exception
end
```

### Mixed Usage

```crystal
require "./src/db"

# You can mix normal and transaction usage
author_repo = TestApp::AuthorsRepository.new

# Normal operation
authors = author_repo.all()

# Transaction operation when needed
TestApp::AuthorsRepository.transaction do |tx_author_repo|
  # Bulk operations in transaction
  tx_author_repo.create("Author 1", "Bio 1")
  tx_author_repo.create("Author 2", "Bio 2")
  tx_author_repo.create("Author 3", "Bio 3")
end
```