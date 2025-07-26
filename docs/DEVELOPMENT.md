# sqlc-gen-crystal Development Guide

This guide provides detailed information for developers working on or contributing to sqlc-gen-crystal.

## Architecture Overview

### Plugin Communication Flow

```
┌─────────┐     Protocol Buffer      ┌────────────────────┐
│  sqlc   │ ───────────────────────> │ sqlc-gen-crystal   │
│         │      (stdin)              │                    │
│         │                           │ 1. Parse Request   │
│         │                           │ 2. Generate Code   │
│         │                           │ 3. Return Response │
│         │ <─────────────────────── │                    │
└─────────┘     Protocol Buffer      └────────────────────┘
                (stdout)
```

### Project Structure

```
sqlc-gen-crystal/
├── main.go                       # Main entry point
├── go.mod                        # Go module dependencies
├── proto/
│   └── codegen.proto            # Protocol buffer definitions
├── internal/
│   ├── codegen/
│   │   └── codegen.go           # Core plugin logic & protobuf handling
│   └── crystal/
│       ├── generator.go         # Crystal code generation logic
│       ├── generator_test.go    # Generator tests
│       ├── strings.go           # String manipulation utilities
│       ├── strings_test.go      # String utility tests
│       ├── templates.go         # Go templates for code generation
│       ├── types.go             # SQL to Crystal type mappings
│       └── types_test.go        # Type mapping tests
├── test/
│   └── integration/             # Integration tests
│       ├── sqlc.yaml
│       ├── schema.sql
│       ├── queries.sql
│       └── spec/
│           └── integration_spec.cr
├── examples/
│   └── athena_example/          # Athena Framework blog API example
│       ├── sqlc.yaml
│       ├── sqlc/
│       │   ├── schema.sql
│       │   └── queries.sql
│       └── src/
│           ├── app.cr
│           └── db/              # Generated code location
└── docs/
    ├── ARCHITECTURE.md          # Technical architecture reference
    └── DEVELOPMENT.md           # This file
```

## Protocol Buffer Definitions

The plugin uses Protocol Buffers for communication with sqlc. Key messages:

### CodeGenRequest

```protobuf
message CodeGenRequest {
  Settings settings = 1;
  Catalog catalog = 2;
  repeated Query queries = 3;
  string sqlc_version = 4;
  bytes plugin_options = 5;
}

message Settings {
  string version = 1;
  string engine = 2;
  repeated string schema = 3;
  repeated string queries = 4;
  CodeGen codegen = 5;
}

message Catalog {
  repeated Schema schemas = 1;
  string default_schema = 2;
}

message Query {
  string text = 1;
  string name = 2;
  string cmd = 3;  // :one, :many, :exec, :execresult, :execrows, :execlastid, :copyfrom
  repeated Column columns = 4;
  repeated Parameter params = 5;
  repeated string comments = 6;
  string filename = 7;
}
```

### CodeGenResponse

```protobuf
message CodeGenResponse {
  repeated File files = 1;
}

message File {
  string name = 1;
  bytes contents = 2;
}
```

## Implementation Details

### Main Entry Point

The plugin is written in Go and uses the standard protobuf library for communication:

```go
// main.go
package main

import (
    "github.com/watzon/sqlc-gen-crystal/internal/codegen"
)

func main() {
    codegen.Run()
}
```

```go
// internal/codegen/codegen.go
func Run() {
    // Read protobuf request from stdin
    req, err := ReadRequest(os.Stdin)
    if err != nil {
        log.Fatal(err)
    }
    
    // Generate Crystal code
    gen := crystal.New(req)
    resp, err := gen.Generate()
    if err != nil {
        log.Fatal(err)
    }
    
    // Write protobuf response to stdout
    if err := WriteResponse(os.Stdout, resp); err != nil {
        log.Fatal(err)
    }
}
```

### Type Mapping Strategy

Type mappings are handled in `internal/crystal/types.go`:

```go
// internal/crystal/types.go
func crystalType(engine string, column *plugin.Column) string {
    baseType := mapSQLType(engine, column.Type)
    if column.NotNull {
        return baseType
    }
    return baseType + "?"
}

func mapSQLType(engine string, sqlType *plugin.Type) string {
    switch engine {
    case "postgresql":
        return postgresType(sqlType)
    case "mysql":
        return mysqlType(sqlType)
    case "sqlite":
        return sqliteType(sqlType)
    default:
        return "String"
    }
}

func postgresType(sqlType *plugin.Type) string {
    switch strings.ToLower(sqlType.Name) {
    case "int8", "bigint", "bigserial":
        return "Int64"
    case "int4", "int", "integer", "serial":
        return "Int32"
    case "int2", "smallint", "smallserial":
        return "Int16"
    case "numeric", "decimal":
        return "Float64"
    case "float4", "real":
        return "Float32"
    case "float8", "double precision":
        return "Float64"
    case "bool", "boolean":
        return "Bool"
    case "text", "varchar", "char", "bpchar":
        return "String"
    case "timestamp", "timestamptz", "date":
        return "Time"
    case "uuid":
        return "String"
    case "json", "jsonb":
        return "JSON::Any"
    case "bytea":
        return "Bytes"
    default:
        return "String"
    }
}
```

### Code Generation

The main generator logic is in `internal/crystal/generator.go`:

```go
// internal/crystal/generator.go
type Generator struct {
    req    *plugin.CodeGenRequest
    opts   Options
    engine string
}

func (g *Generator) Generate() (*plugin.CodeGenResponse, error) {
    var files []*plugin.File
    
    // Always generate database.cr as entry point
    databaseFile, err := g.generateDatabase()
    if err != nil {
        return nil, err
    }
    files = append(files, databaseFile)
    
    // Generate models if we have schema
    if len(g.req.Catalog.Schemas) > 0 {
        modelsFile, err := g.generateModels()
        if err != nil {
            return nil, err
        }
        files = append(files, modelsFile)
    }
    
    // Generate queries if we have them
    if len(g.req.Queries) > 0 {
        queriesFile, err := g.generateQueries()
        if err != nil {
            return nil, err
        }
        files = append(files, queriesFile)
    }
    
    // Generate repositories if enabled
    if g.opts.GenerateRepositories {
        repoFiles, err := g.generateRepositories()
        if err != nil {
            return nil, err
        }
        files = append(files, repoFiles...)
    }
    
    return &plugin.CodeGenResponse{Files: files}, nil
}
```

### Generated Files

The plugin generates these files:

1. **database.cr** - Entry point that requires all other files
2. **models.cr** - Crystal structs for database tables
3. **queries.cr** - Type-safe query methods
4. **repositories/** - Repository pattern classes (optional)

## Adding New Type Mappings

To add support for a new SQL type:

1. Update the appropriate type mapping function in `internal/crystal/types.go`:

```go
func postgresType(sqlType *plugin.Type) string {
    switch strings.ToLower(sqlType.Name) {
    // ... existing mappings ...
    case "my_custom_type":
        return "MyCustomCrystalType"
    // ...
    }
}
```

2. Add a test case in `internal/crystal/types_test.go`:

```go
func TestPostgresTypeMapping(t *testing.T) {
    tests := []struct {
        sqlType  string
        expected string
    }{
        // ... existing tests ...
        {"my_custom_type", "MyCustomCrystalType"},
    }
    // ...
}
```

3. If the type requires special imports, update the template imports

## Testing Strategy

### Unit Tests

Test individual components using Go's testing framework:

```go
// internal/crystal/types_test.go
func TestCrystalType(t *testing.T) {
    tests := []struct {
        name     string
        engine   string
        column   *plugin.Column
        expected string
    }{
        {
            name:   "nullable integer",
            engine: "postgresql",
            column: &plugin.Column{
                Type:    &plugin.Type{Name: "int4"},
                NotNull: false,
            },
            expected: "Int32?",
        },
        {
            name:   "not null text",
            engine: "postgresql",
            column: &plugin.Column{
                Type:    &plugin.Type{Name: "text"},
                NotNull: true,
            },
            expected: "String",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := crystalType(tt.engine, tt.column)
            if result != tt.expected {
                t.Errorf("expected %s, got %s", tt.expected, result)
            }
        })
    }
}
```

### Integration Tests

Crystal integration tests verify generated code compiles and works:

```crystal
# test/integration/spec/integration_spec.cr
require "spec"
require "../src/db"

describe "Generated Code" do
  it "creates and queries authors" do
    DB.open("sqlite3::memory:") do |db|
      # Setup schema
      db.exec(File.read("../schema.sql"))
      
      # Test generated code
      queries = Db::Queries.new(db)
      
      # Create author
      author = queries.create_author("John Doe", "john@example.com")
      author.should_not be_nil
      author.not_nil!.name.should eq("John Doe")
      
      # List authors
      authors = queries.list_authors
      authors.size.should eq(1)
    end
  end
end
```

### Running Tests

```bash
# Run Go unit tests
make test-go

# Run Crystal integration tests
make test-crystal

# Run all tests
make test
```

## Debugging Tips

### Capturing Plugin Input/Output

For debugging, you can capture the protobuf messages:

```bash
# Capture input
sqlc generate --plugin-debug > debug_input.bin

# Inspect the captured input
protoc --decode=plugin.CodeGenRequest plugin/codegen.proto < debug_input.bin
```

### Adding Debug Logging

```go
// Use stderr for debug output (won't interfere with protobuf output)
log.SetOutput(os.Stderr)
log.Printf("Debug: Processing query %s", query.Name)

// Or use conditional debug logging
if os.Getenv("SQLC_DEBUG") != "" {
    log.Printf("Debug: Generated %d files", len(files))
}
```

### Common Issues

1. **Binary/Text Mode**: Ensure stdin/stdout are in binary mode
2. **Protobuf Parsing**: Verify the protobuf schema matches sqlc's version
3. **Type Mismatches**: Check SQL type names against your mappings

## Contributing Guidelines

1. **Code Style**: Follow Go's standard formatting (`go fmt`)
2. **Tests**: Add tests for new features and bug fixes
3. **Documentation**: Update this guide for significant changes
4. **Commits**: Use clear, descriptive commit messages

## Building and Installation

### Building from Source

```bash
# Clone the repository
git clone https://github.com/watzon/sqlc-gen-crystal
cd sqlc-gen-crystal

# Build the plugin
make build

# Install to sqlc plugin directory
make install
```

### Using with sqlc

Configure in your `sqlc.yaml`:

```yaml
version: "2"
plugins:
  - name: crystal
    process:
      cmd: sqlc-gen-crystal
sql:
  - engine: "postgresql"  # or mysql, sqlite
    queries: "queries.sql"
    schema: "schema.sql"
    codegen:
      - plugin: crystal
        out: src/db
        options:
          package: "MyApp"
          emit_json_tags: true
          generate_connection_manager: true
          generate_repositories: true
```

## Examples

The `examples/` directory contains working examples:

1. **athena_example** - Full-featured blog API using Athena Framework
   - Shows repository pattern usage
   - Demonstrates connection manager
   - Includes transaction support
   - Uses keyword arguments for clarity

## Resources

- [sqlc Documentation](https://docs.sqlc.dev/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Crystal Language Reference](https://crystal-lang.org/reference/)
- [crystal-db Documentation](https://crystal-lang.github.io/crystal-db/)