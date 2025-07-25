package crystal

import (
	"testing"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestPostgresTypeMapping(t *testing.T) {
	tests := []struct {
		sqlType  string
		expected string
	}{
		// Integer types
		{"int8", "Int64"},
		{"bigint", "Int64"},
		{"bigserial", "Int64"},
		{"int4", "Int32"},
		{"int", "Int32"},
		{"integer", "Int32"},
		{"serial", "Int32"},
		{"int2", "Int16"},
		{"smallint", "Int16"},
		{"smallserial", "Int16"},

		// Floating point types
		{"numeric", "Float64"},
		{"decimal", "Float64"},
		{"real", "Float32"},
		{"float4", "Float32"},
		{"float8", "Float64"},
		{"double precision", "Float64"},

		// Boolean
		{"bool", "Bool"},
		{"boolean", "Bool"},

		// String types
		{"text", "String"},
		{"varchar", "String"},
		{"char", "String"},
		{"bpchar", "String"},
		{"citext", "String"},
		{"name", "String"},

		// Time types
		{"timestamp", "Time"},
		{"timestamptz", "Time"},
		{"date", "Time"},
		{"time", "Time"},
		{"timetz", "Time"},
		{"interval", "Time::Span"},

		// UUID
		{"uuid", "String"},

		// JSON types
		{"json", "JSON::Any"},
		{"jsonb", "JSON::Any"},

		// Binary
		{"bytea", "Bytes"},

		// Network types
		{"inet", "String"},
		{"cidr", "String"},
		{"macaddr", "String"},

		// Unknown type
		{"unknown_type", "String"},
	}

	gen := &Generator{
		req: &plugin.GenerateRequest{
			Settings: &plugin.Settings{
				Engine: "postgresql",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.sqlType, func(t *testing.T) {
			result := gen.postgresType(tt.sqlType)
			if result != tt.expected {
				t.Errorf("postgresType(%q) = %q, want %q", tt.sqlType, result, tt.expected)
			}
		})
	}
}

func TestMySQLTypeMapping(t *testing.T) {
	tests := []struct {
		sqlType  string
		expected string
	}{
		// Integer types
		{"bigint", "Int64"},
		{"int", "Int32"},
		{"integer", "Int32"},
		{"mediumint", "Int32"},
		{"smallint", "Int16"},
		{"tinyint", "Int8"},

		// Floating point types
		{"decimal", "Float64"},
		{"numeric", "Float64"},
		{"float", "Float32"},
		{"double", "Float64"},
		{"double precision", "Float64"},
		{"real", "Float64"},

		// Boolean
		{"bit", "Bool"},
		{"bool", "Bool"},
		{"boolean", "Bool"},

		// String types
		{"char", "String"},
		{"varchar", "String"},
		{"text", "String"},
		{"tinytext", "String"},
		{"mediumtext", "String"},
		{"longtext", "String"},

		// Time types
		{"datetime", "Time"},
		{"timestamp", "Time"},
		{"date", "Time"},
		{"time", "Time::Span"},
		{"year", "Int32"},

		// JSON
		{"json", "JSON::Any"},

		// Binary types
		{"binary", "Bytes"},
		{"varbinary", "Bytes"},
		{"blob", "Bytes"},
		{"tinyblob", "Bytes"},
		{"mediumblob", "Bytes"},
		{"longblob", "Bytes"},

		// Enum and set
		{"enum", "String"},
		{"set", "String"},

		// Unknown type
		{"unknown_type", "String"},
	}

	gen := &Generator{
		req: &plugin.GenerateRequest{
			Settings: &plugin.Settings{
				Engine: "mysql",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.sqlType, func(t *testing.T) {
			result := gen.mysqlType(tt.sqlType)
			if result != tt.expected {
				t.Errorf("mysqlType(%q) = %q, want %q", tt.sqlType, result, tt.expected)
			}
		})
	}
}

func TestSQLiteTypeMapping(t *testing.T) {
	tests := []struct {
		sqlType  string
		expected string
	}{
		// Integer affinity
		{"integer", "Int64"},
		{"int", "Int64"},
		{"tinyint", "Int64"},
		{"smallint", "Int64"},
		{"mediumint", "Int64"},
		{"bigint", "Int64"},
		{"int2", "Int64"},
		{"int8", "Int64"},

		// Real affinity
		{"real", "Float64"},
		{"double", "Float64"},
		{"double precision", "Float64"},
		{"float", "Float64"},

		// Text affinity
		{"text", "String"},
		{"varchar", "String"},
		{"char", "String"},
		{"clob", "String"},

		// Blob affinity
		{"blob", "Bytes"},

		// Numeric affinity
		{"numeric", "Float64"},
		{"decimal", "Float64"},

		// Boolean
		{"boolean", "Bool"},
		{"bool", "Bool"},

		// Date/time types
		{"date", "Time"},
		{"datetime", "Time"},
		{"timestamp", "Time"},
		{"time", "Time"},

		// Unknown type
		{"unknown_type", "String"},
	}

	gen := &Generator{
		req: &plugin.GenerateRequest{
			Settings: &plugin.Settings{
				Engine: "sqlite",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.sqlType, func(t *testing.T) {
			result := gen.sqliteType(tt.sqlType)
			if result != tt.expected {
				t.Errorf("sqliteType(%q) = %q, want %q", tt.sqlType, result, tt.expected)
			}
		})
	}
}

func TestCrystalTypeWithNullability(t *testing.T) {
	tests := []struct {
		name        string
		column      *plugin.Column
		expected    string
		withPointer bool
	}{
		{
			name: "not null integer",
			column: &plugin.Column{
				Type:    &plugin.Identifier{Name: "int4"},
				NotNull: true,
			},
			expected: "Int32",
		},
		{
			name: "nullable integer",
			column: &plugin.Column{
				Type:    &plugin.Identifier{Name: "int4"},
				NotNull: false,
			},
			expected: "Int32?",
		},
		{
			name: "nullable integer with pointer",
			column: &plugin.Column{
				Type:    &plugin.Identifier{Name: "int4"},
				NotNull: false,
			},
			expected:    "Int32*",
			withPointer: true,
		},
		{
			name: "integer array",
			column: &plugin.Column{
				Type:    &plugin.Identifier{Name: "int4"},
				IsArray: true,
				NotNull: true,
			},
			expected: "Array(Int32)",
		},
		{
			name: "nullable text",
			column: &plugin.Column{
				Type:    &plugin.Identifier{Name: "text"},
				NotNull: false,
			},
			expected: "String?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &Generator{
				req: &plugin.GenerateRequest{
					Settings: &plugin.Settings{
						Engine: "postgresql",
					},
				},
				options: GeneratorOptions{
					EmitResultStructPointers: tt.withPointer,
				},
			}

			result := gen.crystalType(tt.column)
			if result != tt.expected {
				t.Errorf("crystalType() = %q, want %q", result, tt.expected)
			}
		})
	}
}
