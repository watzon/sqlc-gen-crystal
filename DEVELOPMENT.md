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
├── src/
│   ├── sqlc_gen_crystal.cr       # Main entry point
│   ├── plugin/
│   │   ├── codegen.pb.cr         # Generated protobuf definitions
│   │   └── codegen.proto         # Protobuf schema (reference)
│   ├── generator/
│   │   ├── generator.cr          # Core code generation logic
│   │   ├── type_mapper.cr        # SQL to Crystal type mappings
│   │   └── templates/            # Code generation templates
│   │       ├── models.ecr        # Model struct template
│   │       └── queries.ecr       # Query methods template
│   └── drivers/
│       ├── postgres.cr           # PostgreSQL-specific logic
│       ├── mysql.cr              # MySQL-specific logic
│       └── sqlite.cr             # SQLite-specific logic
├── spec/                         # Test files
├── examples/                     # Example projects
└── shard.yml                     # Crystal dependencies
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

```crystal
# src/sqlc_gen_crystal.cr
require "./plugin/codegen.pb"
require "./generator/generator"

module SqlcGenCrystal
  def self.run
    # Read binary input from stdin
    input = STDIN.gets_to_end
    
    # Parse protobuf request
    request = Plugin::CodeGenRequest.from_protobuf(IO::Memory.new(input.to_slice))
    
    # Generate code
    generator = Generator.new(request)
    response = generator.generate
    
    # Write binary output to stdout
    STDOUT.write(response.to_protobuf)
    STDOUT.flush
  end
end

SqlcGenCrystal.run
```

### Type Mapping Strategy

```crystal
# src/generator/type_mapper.cr
module SqlcGenCrystal
  class TypeMapper
    def initialize(@engine : String)
    end
    
    def crystal_type(column : Plugin::Column) : String
      base_type = map_base_type(column.type)
      column.not_null ? base_type : "#{base_type}?"
    end
    
    private def map_base_type(sql_type : Plugin::Type) : String
      case @engine
      when "postgresql"
        postgres_type_map(sql_type)
      when "mysql"
        mysql_type_map(sql_type)
      when "sqlite"
        sqlite_type_map(sql_type)
      else
        raise "Unsupported engine: #{@engine}"
      end
    end
    
    private def postgres_type_map(sql_type : Plugin::Type) : String
      case sql_type.name.downcase
      when "int8", "bigint", "bigserial"
        "Int64"
      when "int4", "int", "integer", "serial"
        "Int32"
      when "int2", "smallint", "smallserial"
        "Int16"
      when "numeric", "decimal"
        "Float64"
      when "float4", "real"
        "Float32"
      when "float8", "double precision"
        "Float64"
      when "bool", "boolean"
        "Bool"
      when "text", "varchar", "char", "bpchar"
        "String"
      when "timestamp", "timestamptz", "date", "time", "timetz"
        "Time"
      when "uuid"
        "String"
      when "json", "jsonb"
        "JSON::Any"
      when "bytea"
        "Bytes"
      when "inet", "cidr"
        "String"
      else
        "String" # Default fallback
      end
    end
  end
end
```

### Code Generation

```crystal
# src/generator/generator.cr
module SqlcGenCrystal
  class Generator
    def initialize(@request : Plugin::CodeGenRequest)
      @type_mapper = TypeMapper.new(@request.settings.engine)
      @options = parse_options(@request.plugin_options)
    end
    
    def generate : Plugin::CodeGenResponse
      files = [] of Plugin::File
      
      # Generate models file
      if has_models?
        files << generate_models_file
      end
      
      # Generate queries file
      if has_queries?
        files << generate_queries_file
      end
      
      Plugin::CodeGenResponse.new(files: files)
    end
    
    private def generate_models_file : Plugin::File
      content = ECR.render("src/generator/templates/models.ecr")
      Plugin::File.new(
        name: "models.cr",
        contents: content.to_slice
      )
    end
    
    private def generate_queries_file : Plugin::File
      content = ECR.render("src/generator/templates/queries.ecr")
      Plugin::File.new(
        name: "queries.cr",
        contents: content.to_slice
      )
    end
  end
end
```

## Adding New Type Mappings

To add support for a new SQL type:

1. Update the appropriate type mapping method in `type_mapper.cr`:

```crystal
private def postgres_type_map(sql_type : Plugin::Type) : String
  case sql_type.name.downcase
  # ... existing mappings ...
  when "my_custom_type"
    "MyCustomCrystalType"
  # ...
  end
end
```

2. Add any necessary imports or custom type definitions to the templates

3. Add tests for the new type mapping

## Testing Strategy

### Unit Tests

Test individual components:

```crystal
# spec/type_mapper_spec.cr
describe SqlcGenCrystal::TypeMapper do
  it "maps PostgreSQL types correctly" do
    mapper = SqlcGenCrystal::TypeMapper.new("postgresql")
    
    column = Plugin::Column.new(
      type: Plugin::Type.new(name: "int8"),
      not_null: true
    )
    
    mapper.crystal_type(column).should eq("Int64")
  end
end
```

### Integration Tests

Test the full plugin flow:

```crystal
# spec/integration_spec.cr
describe "Plugin Integration" do
  it "generates correct code for a simple query" do
    request = build_test_request(
      queries: [simple_select_query],
      schema: [authors_table]
    )
    
    generator = SqlcGenCrystal::Generator.new(request)
    response = generator.generate
    
    response.files.size.should eq(2)
    response.files[0].name.should eq("models.cr")
    response.files[1].name.should eq("queries.cr")
  end
end
```

### End-to-End Tests

Test with actual sqlc:

```bash
# test/e2e/test.sh
#!/bin/bash
cd examples/booktest
sqlc generate
crystal build src/main.cr
./main
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

```crystal
# Use STDERR for debug output (won't interfere with protobuf output)
STDERR.puts "Debug: Processing query #{query.name}"
```

### Common Issues

1. **Binary/Text Mode**: Ensure stdin/stdout are in binary mode
2. **Protobuf Parsing**: Verify the protobuf schema matches sqlc's version
3. **Type Mismatches**: Check SQL type names against your mappings

## Contributing Guidelines

1. **Code Style**: Follow Crystal's standard formatting (`crystal tool format`)
2. **Tests**: Add tests for new features and bug fixes
3. **Documentation**: Update this guide for significant changes
4. **Commits**: Use clear, descriptive commit messages

## Release Process

1. Update version in `shard.yml`
2. Update CHANGELOG.md
3. Run all tests
4. Build release binary: `shards build --release`
5. Create GitHub release with binary artifacts
6. Consider publishing to Crystal shards registry

## Resources

- [sqlc Documentation](https://docs.sqlc.dev/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Crystal Language Reference](https://crystal-lang.org/reference/)
- [crystal-db Documentation](https://crystal-lang.github.io/crystal-db/)