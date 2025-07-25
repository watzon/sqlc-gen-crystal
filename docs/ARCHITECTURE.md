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

### 8. Configuration Options

Plugin accepts options via JSON in the protobuf request:

```json
{
  "package": "db",
  "emit_json_tags": true,
  "emit_db_tags": true,
  "emit_result_struct_pointers": false,
  "naming_convention": "snake_case"
}
```

Options are parsed and validated:
```crystal
struct Options
  include JSON::Serializable

  @[JSON::Field(key: "package")]
  property package : String = "db"

  @[JSON::Field(key: "emit_json_tags")]
  property emit_json_tags : Bool = false

  @[JSON::Field(key: "emit_db_tags")]
  property emit_db_tags : Bool = true

  @[JSON::Field(key: "emit_result_struct_pointers")]
  property emit_result_struct_pointers : Bool = false

  @[JSON::Field(key: "naming_convention")]
  property naming_convention : String = "snake_case"
end
```

### 9. Testing Architecture

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

### 10. Future Considerations

1. **Streaming Queries**: Support for result streaming
2. **Custom Types**: User-defined type mappings
3. **Migrations**: Integration with migration tools
4. **Validation**: Query validation at generation time
5. **Optimizations**: Query batching support
