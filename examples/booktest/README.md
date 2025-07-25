# Booktest Example

This is an example project demonstrating the sqlc-gen-crystal plugin.

## Setup

1. Create a PostgreSQL database:
```bash
createdb booktest
```

2. Run the schema:
```bash
psql booktest < schema.sql
```

3. Generate the Crystal code:
```bash
sqlc generate
```

4. Install Crystal dependencies:
```bash
shards install
```

5. Run the example:
```bash
crystal run src/main.cr
```

## Generated Files

- `src/db/models.cr` - Contains Crystal structs for all tables and query results
- `src/db/queries.cr` - Contains the Queries class with methods for all SQL queries

## Notes

The plugin generates:
- Type-safe structs for all database tables
- Methods for each query with proper parameter and return types
- Support for nullable columns using Crystal's union types
- JSON serialization support when enabled