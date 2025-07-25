# SELECT Query Features Verification

Based on the SQLC documentation at https://docs.sqlc.dev/en/latest/howto/select.html, here's the verification of all supported SELECT features in sqlc-gen-crystal:

## ✅ Basic SELECT queries

### Single row queries (:one)
- Returns `T?` (nullable type)
- Handles both single and multiple column results

### Multiple row queries (:many)
- Returns `Array(T)`
- Handles both single and multiple column results

## ✅ Single column SELECT
- **Fixed**: Template bug where single column array results were using incorrect type extraction
- Now correctly returns the column type directly without a struct
- Example: `SELECT name FROM authors` returns `String?` for `:one` or `Array(String)` for `:many`

## ✅ Multiple column SELECT
- Generates structs with proper field mappings
- Deduplicates structs with identical field signatures
- Supports table-based and query-specific structs

## ✅ PostgreSQL array parameters
- Supports `ANY()` operator with native array types
- Example: `WHERE id = ANY($1::int[])` accepts `Array(Int32)`
- Arrays are passed directly to the database driver

## ✅ MySQL/SQLite slice parameters (sqlc.slice())
- **Implemented**: Full support for `sqlc.slice()` meta-function
- Detects `IsSqlcSlice` parameter metadata
- Generates engine-specific code:
  - PostgreSQL: Direct array passing
  - MySQL/SQLite: Runtime SQL transformation to expand placeholders
- Handles empty slice validation
- Flattens arrays for proper parameter passing

## Implementation Details

### Engine-specific code generation
- PostgreSQL: Uses native array support with `ANY()` operator
- MySQL/SQLite: Dynamically expands `/*SLICE:name*/?` markers to multiple `?` placeholders
- Compile-time engine detection determines which approach to use

### Generated code structure
```crystal
# PostgreSQL (direct array passing)
@db.query_all(SQL_QUERIES[:QUERY_NAME], array_param, as: ResultType)

# MySQL/SQLite (runtime expansion)
sql = SQL_QUERIES[:QUERY_NAME]
placeholders = array_param.size.times.map { "?" }.join(", ")
sql = sql.gsub("/*SLICE:param_name*/?", placeholders)
query_params = [] of DB::Any
query_params.concat(array_param.map { |v| v.as(DB::Any) })
@db.query_all(sql, args: query_params, as: ResultType)
```

## Conclusion
All SELECT query features documented in SQLC are now fully supported in sqlc-gen-crystal, with engine-specific optimizations for array/slice parameters.