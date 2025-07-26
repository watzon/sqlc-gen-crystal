# sqlc-gen-crystal Technical Architecture

This document serves as an internal reference for the sqlc-gen-crystal plugin implementation.

## Core Components

### 1. Protocol Buffer Interface

The plugin communicates with sqlc using Protocol Buffers v3. The schema is defined in `plugin/codegen.proto`.

Key message types:
- `CodeGenRequest`: Input from sqlc containing SQL queries, schema, and settings
- `CodeGenResponse`: Output containing generated Crystal files
- `Query`: Individual SQL query with metadata
- `Column`: Column information including type and nullability
- `Parameter`: Query parameter information

### 2. Type System Mapping

#### SQL to Crystal Type Conversion Rules

The type mapper follows these principles:
1. **Nullability**: SQL nullable columns map to Crystal union types (e.g., `String?`)
2. **Precision**: Maintain numeric precision where possible
3. **Native Types**: Use Crystal stdlib types when appropriate
4. **Database Compatibility**: Handle driver-specific type differences

#### Type Mapping Tables

**PostgreSQL Types**
```
int8, bigint, bigserial     -> Int64
int4, integer, serial       -> Int32
int2, smallint              -> Int16
numeric, decimal            -> Float64
real, float4                -> Float32
double precision, float8    -> Float64
boolean                     -> Bool
text, varchar, char         -> String
timestamp, timestamptz      -> Time
date                        -> Time
uuid                        -> String
json, jsonb                 -> JSON::Any
bytea                       -> Bytes
array types                 -> Array(T)
```

**MySQL Types**
```
bigint                      -> Int64
int, integer                -> Int32
smallint                    -> Int16
tinyint                     -> Int8
decimal, numeric            -> Float64
float                       -> Float32
double                      -> Float64
bit, boolean                -> Bool
varchar, text, char         -> String
datetime, timestamp         -> Time
date                        -> Time
time                        -> Time::Span
json                        -> JSON::Any
blob, binary                -> Bytes
```

**SQLite Types**
```
integer                     -> Int64
real                        -> Float64
text                        -> String
blob                        -> Bytes
numeric                     -> Float64
boolean                     -> Bool
datetime, timestamp         -> Time
```

### 3. Code Generation Strategy

#### Query Classification

Queries are classified by their command type (`:cmd` field):
- `:one` - Returns zero or one row
- `:many` - Returns zero or more rows
- `:exec` - Executes without returning data
- `:execresult` - Returns execution result metadata
- `:execrows` - Returns number of affected rows
- `:execlastid` - Returns last inserted ID
- `:copyfrom` - Bulk insert operation

#### Method Generation Patterns

**Query :one**
```crystal
def method_name(param1 : Type1, param2 : Type2) : ReturnType?
  @db.query_one?(
    "SQL QUERY HERE",
    param1, param2,
    as: ReturnType
  )
end
```

**Query :many**
```crystal
def method_name(param1 : Type1) : Array(ReturnType)
  @db.query_all(
    "SQL QUERY HERE",
    param1,
    as: ReturnType
  )
end
```

**Query :exec**
```crystal
def method_name(param1 : Type1) : Nil
  @db.exec(
    "SQL QUERY HERE",
    param1
  )
end
```

### 4. Template System

Uses Crystal's ECR (Embedded Crystal) for code generation templates.

**models.ecr Structure**
```ecr
module <%= module_name %>
  <% structs.each do |struct| %>
  struct <%= struct.name %>
    include DB::Serializable
    <% if emit_json_tags %>
    include JSON::Serializable
    <% end %>

    <% struct.fields.each do |field| %>
    <% if emit_json_tags && field.json_name %>
    @[JSON::Field(key: "<%= field.json_name %>")]
    <% end %>
    <% if emit_db_tags && field.db_name != field.name %>
    @[DB::Field(key: "<%= field.db_name %>")]
    <% end %>
    getter <%= field.name %> : <%= field.type %>
    <% end %>
  end
  <% end %>
end
```

### 5. Driver-Specific Handling

Each database driver has specific requirements:

**PostgreSQL**
- Parameter placeholders: `$1`, `$2`, etc.
- Array type support
- Custom type handling (e.g., ENUM types)

**MySQL**
- Parameter placeholders: `?`
- Different timestamp handling
- No array types

**SQLite**
- Parameter placeholders: `?` or `?NNN`
- Limited type system
- Type affinity rules

### 6. Error Handling Strategy

The plugin implements defensive error handling:

1. **Input Validation**: Verify protobuf message structure
2. **Type Safety**: Fail fast on unknown SQL types
3. **Clear Error Messages**: Provide context for debugging
4. **Graceful Degradation**: Use sensible defaults when possible

Example:
```crystal
def map_type(sql_type : String) : String
  mapped = TYPE_MAP[sql_type]?
  unless mapped
    STDERR.puts "Warning: Unknown SQL type '#{sql_type}', defaulting to String"
    return "String"
  end
  mapped
end
```

### 7. Performance Considerations

1. **Memory Efficiency**: Stream processing for large schemas
2. **String Building**: Use String::Builder for concatenation
3. **Template Caching**: Compile ECR templates once
4. **Minimal Dependencies**: Keep the plugin lightweight

### 8. Generated File Structure

The plugin generates the following files:

1. **database.cr** - Main entry point (always generated)
   - Requires all other generated files
   - Contains connection manager (if enabled)
   - Provides centralized access to all database functionality

2. **models.cr** - Model structs with DB::Serializable
   - Generated from database schema
   - Includes JSON/YAML serialization (if enabled)
   - Type-safe representations of database tables

3. **queries.cr** - Query methods
   - Generated from SQL queries
   - Type-safe method signatures
   - Handles all database operations

4. **repositories/** - Repository pattern classes (if enabled)
   - One repository per table/model
   - Encapsulates database operations
   - Supports transaction handling

**database.cr Structure**
```crystal
# Main entry point for generated database code
# Always require models and queries
require "./models"
require "./queries"

<% if generate_repositories %>
# Require all repository files
require "./repositories/*"
<% end %>

<% if generate_connection_manager %>
module <%= package %>
  class Database
    # Singleton connection manager
    @@instance : DB::Database?
    
    def self.connection : DB::Database
      @@instance ||= DB.open(ENV["DATABASE_URL"])
    end
    
    def self.queries : Queries
      @@queries ||= Queries.new(connection)
    end
    
    # Transaction support
    def self.transaction(&block : Queries ->)
      connection.transaction do |tx|
        tx_queries = Queries.new(tx)
        yield tx_queries
      end
    end
  end
end
<% end %>
```

### 9. Configuration Options

Plugin accepts options via JSON in the protobuf request:

```json
{
  "package": "DB",
  "emit_json_tags": true,
  "emit_db_tags": true,
  "emit_result_struct_pointers": false,
  "emit_boolean_question_getters": false,
  "generate_connection_manager": false,
  "generate_repositories": false
}
```

Options are parsed and validated:
```go
type Options struct {
    Package                     string `json:"package"`
    EmitJSONTags                bool   `json:"emit_json_tags"`
    EmitDBTags                  bool   `json:"emit_db_tags"`
    EmitResultStructPointers    bool   `json:"emit_result_struct_pointers"`
    EmitBooleanQuestionGetters  bool   `json:"emit_boolean_question_getters"`
    GenerateConnectionManager   bool   `json:"generate_connection_manager"`
    GenerateRepositories        bool   `json:"generate_repositories"`
}
```

### 10. Testing Architecture

**Unit Tests**
- Type mapper correctness
- Template rendering
- Option parsing

**Integration Tests**
- Full query generation
- Multi-file output
- Edge cases

**Smoke Tests**
- Basic sqlc integration
- Database driver compatibility
- Generated code compilation

### 11. Repository Pattern Implementation

When `generate_repositories` is enabled, the plugin generates repository classes that provide a higher-level abstraction over the raw query methods.

**Repository Features**
- One repository per table/model
- Encapsulates CRUD operations
- Transaction support
- Consistent naming conventions
- Type-safe method signatures

**Example Repository Structure**
```crystal
module Blog
  class PostsRepository
    def initialize(@queries : Queries? = nil)
    end

    private def queries
      @queries || Database.queries
    end

    # CRUD operations
    def create(user_id : Int64, title : String, ...) : Post?
      queries.create_post(user_id, title, ...)
    end

    def find(id : Int64) : Post?
      queries.get_post(id)
    end

    def all(limit : Int64, offset : Int64) : Array(Post)
      queries.list_posts(limit, offset)
    end

    def update(id : Int64, title : String, ...) : Nil
      queries.update_post(id, title, ...)
    end

    def delete(id : Int64) : Nil
      queries.delete_post(id)
    end

    # Transaction support
    def self.transaction(&block : PostsRepository ->)
      Database.transaction do |tx_queries|
        repo = new(tx_queries)
        yield repo
      end
    end
  end
end
```

### 12. Future Considerations

1. **Streaming Queries**: Support for result streaming
2. **Custom Types**: User-defined type mappings
3. **Migrations**: Integration with migration tools
4. **Validation**: Query validation at generation time
5. **Optimizations**: Query batching support
