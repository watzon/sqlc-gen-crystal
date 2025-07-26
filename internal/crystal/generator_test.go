package crystal

import (
	"context"
	"strings"
	"testing"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestGenerateModels(t *testing.T) {
	req := &plugin.GenerateRequest{
		Settings: &plugin.Settings{
			Engine: "postgresql",
		},
		Catalog: &plugin.Catalog{
			Schemas: []*plugin.Schema{
				{
					Name: "public",
					Tables: []*plugin.Table{
						{
							Rel: &plugin.Identifier{
								Name: "authors",
							},
							Columns: []*plugin.Column{
								{
									Name:    "id",
									Type:    &plugin.Identifier{Name: "int4"},
									NotNull: true,
								},
								{
									Name:    "name",
									Type:    &plugin.Identifier{Name: "text"},
									NotNull: true,
								},
								{
									Name:    "bio",
									Type:    &plugin.Identifier{Name: "text"},
									NotNull: false,
								},
							},
						},
					},
				},
			},
		},
		Queries: []*plugin.Query{
			{
				Name: "GetAuthor",
				Cmd:  ":one",
				Columns: []*plugin.Column{
					{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
					{Name: "name", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
					{Name: "bio", Type: &plugin.Identifier{Name: "text"}, NotNull: false},
				},
			},
		},
	}

	gen := NewGenerator(req, "db", GeneratorOptions{
		EmitJSONTags: true,
		EmitDBTags:   true,
	})

	resp, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(resp.Files) != 3 {
		t.Fatalf("Expected 3 files, got %d", len(resp.Files))
	}

	// Check models file
	modelsFile := resp.Files[0]
	if modelsFile.Name != "models.cr" {
		t.Errorf("Expected models.cr, got %s", modelsFile.Name)
	}

	modelsContent := string(modelsFile.Contents)

	// Check for module declaration
	if !strings.Contains(modelsContent, "module Db") {
		t.Error("Models file should contain module Db")
	}

	// Check for Author struct - should use singular form of table name
	if !strings.Contains(modelsContent, "struct Author") {
		t.Errorf("Models file should contain Author struct, got:\n%s", modelsContent)
	}

	// Check for JSON tags
	if !strings.Contains(modelsContent, `@[JSON::Field(key: "id")]`) {
		t.Error("Models file should contain JSON field annotations")
	}

	// Check for proper types
	if !strings.Contains(modelsContent, "getter id : Int32") {
		t.Error("Models file should have proper type for id field")
	}

	if !strings.Contains(modelsContent, "getter bio : String?") {
		t.Error("Models file should have nullable type for bio field")
	}

	// Check database file (should be third)
	databaseFile := resp.Files[2]
	if databaseFile.Name != "database.cr" {
		t.Errorf("Expected database.cr, got %s", databaseFile.Name)
	}

	databaseContent := string(databaseFile.Contents)
	if !strings.Contains(databaseContent, `require "./models"`) {
		t.Error("Database file should require models")
	}

	if !strings.Contains(databaseContent, `require "./queries"`) {
		t.Error("Database file should require queries")
	}
}

func TestGenerateQueries(t *testing.T) {
	req := &plugin.GenerateRequest{
		Settings: &plugin.Settings{
			Engine: "postgresql",
		},
		Queries: []*plugin.Query{
			{
				Name: "GetAuthor",
				Text: "SELECT * FROM authors WHERE id = $1",
				Cmd:  ":one",
				Params: []*plugin.Parameter{
					{
						Number: 1,
						Column: &plugin.Column{
							Name:    "id",
							Type:    &plugin.Identifier{Name: "int4"},
							NotNull: true,
						},
					},
				},
				Columns: []*plugin.Column{
					{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
					{Name: "name", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
					{Name: "bio", Type: &plugin.Identifier{Name: "text"}, NotNull: false},
				},
			},
			{
				Name: "ListAuthors",
				Text: "SELECT * FROM authors ORDER BY name",
				Cmd:  ":many",
				Columns: []*plugin.Column{
					{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
					{Name: "name", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
					{Name: "bio", Type: &plugin.Identifier{Name: "text"}, NotNull: false},
				},
			},
			{
				Name: "DeleteAuthor",
				Text: "DELETE FROM authors WHERE id = $1",
				Cmd:  ":exec",
				Params: []*plugin.Parameter{
					{
						Number: 1,
						Column: &plugin.Column{
							Name:    "id",
							Type:    &plugin.Identifier{Name: "int4"},
							NotNull: true,
						},
					},
				},
			},
			{
				Name: "CountAuthors",
				Text: "SELECT COUNT(*) FROM authors",
				Cmd:  ":one",
				Columns: []*plugin.Column{
					{Name: "count", Type: &plugin.Identifier{Name: "int8"}, NotNull: true},
				},
			},
		},
	}

	gen := NewGenerator(req, "db", GeneratorOptions{})

	resp, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(resp.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(resp.Files))
	}

	// Check queries file (should be first)
	queriesFile := resp.Files[0]
	if queriesFile.Name != "queries.cr" {
		t.Errorf("Expected queries.cr, got %s", queriesFile.Name)
	}

	// Check database file (should be second)
	databaseFile := resp.Files[1]
	if databaseFile.Name != "database.cr" {
		t.Errorf("Expected database.cr, got %s", databaseFile.Name)
	}

	queriesContent := string(queriesFile.Contents)

	// Check for module declaration
	if !strings.Contains(queriesContent, "module Db") {
		t.Error("Queries file should contain module Db")
	}

	// Check for Queries class
	if !strings.Contains(queriesContent, "class Queries") {
		t.Error("Queries file should contain Queries class")
	}

	// Check for SQL constants
	if !strings.Contains(queriesContent, "SQL_4_QUERIES") {
		t.Error("Queries file should contain SQL constants")
	}

	// Check for method signatures
	if !strings.Contains(queriesContent, "def get_author(id : Int32) : GetAuthorRow?") {
		t.Error("Queries file should contain get_author method with correct signature")
	}

	if !strings.Contains(queriesContent, "def list_authors() : Array(ListAuthorsRow)") {
		t.Error("Queries file should contain list_authors method with correct signature")
	}

	if !strings.Contains(queriesContent, "def delete_author(id : Int32) : Nil") {
		t.Error("Queries file should contain delete_author method with correct signature")
	}

	// Check for single column query handling
	if !strings.Contains(queriesContent, "def count_authors() : Int64?") {
		t.Error("Queries file should contain count_authors method returning single column type")
	}

	if !strings.Contains(queriesContent, "rs.read(Int64)") {
		t.Error("Queries file should use rs.read for single column queries")
	}
}

func TestSkipSystemTables(t *testing.T) {
	req := &plugin.GenerateRequest{
		Settings: &plugin.Settings{
			Engine: "postgresql",
		},
		Catalog: &plugin.Catalog{
			Schemas: []*plugin.Schema{
				{
					Name: "information_schema",
					Tables: []*plugin.Table{
						{
							Rel: &plugin.Identifier{
								Name:   "columns",
								Schema: "information_schema",
							},
							Columns: []*plugin.Column{
								{Name: "column_name", Type: &plugin.Identifier{Name: "text"}},
							},
						},
					},
				},
				{
					Name: "pg_catalog",
					Tables: []*plugin.Table{
						{
							Rel: &plugin.Identifier{
								Name:   "pg_class",
								Schema: "pg_catalog",
							},
							Columns: []*plugin.Column{
								{Name: "relname", Type: &plugin.Identifier{Name: "text"}},
							},
						},
					},
				},
				{
					Name: "public",
					Tables: []*plugin.Table{
						{
							Rel: &plugin.Identifier{
								Name: "pg_stat_statements",
							},
							Columns: []*plugin.Column{
								{Name: "query", Type: &plugin.Identifier{Name: "text"}},
							},
						},
						{
							Rel: &plugin.Identifier{
								Name: "users",
							},
							Columns: []*plugin.Column{
								{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
								{Name: "email", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
							},
						},
					},
				},
			},
		},
	}

	gen := NewGenerator(req, "db", GeneratorOptions{})

	resp, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(resp.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(resp.Files))
	}

	modelsContent := string(resp.Files[0].Contents)

	// Should not contain system tables
	if strings.Contains(modelsContent, "struct Columns") {
		t.Error("Should not generate struct for information_schema tables")
	}

	if strings.Contains(modelsContent, "struct PgClass") {
		t.Error("Should not generate struct for pg_catalog tables")
	}

	if strings.Contains(modelsContent, "struct PgStatStatements") {
		t.Error("Should not generate struct for pg_ prefixed tables")
	}

	// Should contain user table (singularized)
	if !strings.Contains(modelsContent, "struct User") {
		t.Error("Should generate struct for user tables")
	}
}

func TestNullableParameterDefaults(t *testing.T) {
	req := &plugin.GenerateRequest{
		Settings: &plugin.Settings{
			Engine: "postgresql",
		},
		Queries: []*plugin.Query{
			{
				Name: "CreateAuthor",
				Text: "INSERT INTO authors (name, bio) VALUES ($1, $2) RETURNING *",
				Cmd:  ":one",
				Params: []*plugin.Parameter{
					{
						Number: 1,
						Column: &plugin.Column{
							Name:    "name",
							Type:    &plugin.Identifier{Name: "text"},
							NotNull: true,
						},
					},
					{
						Number: 2,
						Column: &plugin.Column{
							Name:    "bio",
							Type:    &plugin.Identifier{Name: "text"},
							NotNull: false, // This makes it nullable
						},
					},
				},
				Columns: []*plugin.Column{
					{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
					{Name: "name", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
					{Name: "bio", Type: &plugin.Identifier{Name: "text"}, NotNull: false},
				},
			},
		},
	}

	gen := NewGenerator(req, "db", GeneratorOptions{})

	resp, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(resp.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(resp.Files))
	}

	queriesContent := string(resp.Files[0].Contents)

	// Check that nullable parameter has default nil value
	expectedSignature := "def create_author(name : String, bio : String? = nil) : CreateAuthorRow?"
	if !strings.Contains(queriesContent, expectedSignature) {
		t.Errorf("Expected method signature with default nil for nullable parameter:\n%s\nGot:\n%s", expectedSignature, queriesContent)
	}

	// Check that the parameter order in the actual query call is correct (original SQL order)
	expectedCall := "name, bio,"
	if !strings.Contains(queriesContent, expectedCall) {
		t.Errorf("Expected parameters in SQL order in query call: %s\nGot:\n%s", expectedCall, queriesContent)
	}
}

func TestParameterOrdering(t *testing.T) {
	req := &plugin.GenerateRequest{
		Settings: &plugin.Settings{
			Engine: "postgresql",
		},
		Queries: []*plugin.Query{
			{
				Name: "CreateBook",
				Text: "INSERT INTO books (author_id, title, description, price, isbn, published) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *",
				Cmd:  ":one",
				Params: []*plugin.Parameter{
					{
						Number: 1,
						Column: &plugin.Column{
							Name:    "author_id",
							Type:    &plugin.Identifier{Name: "int4"},
							NotNull: true, // Required
						},
					},
					{
						Number: 2,
						Column: &plugin.Column{
							Name:    "title",
							Type:    &plugin.Identifier{Name: "text"},
							NotNull: true, // Required
						},
					},
					{
						Number: 3,
						Column: &plugin.Column{
							Name:    "description",
							Type:    &plugin.Identifier{Name: "text"},
							NotNull: false, // Nullable - should come after required params
						},
					},
					{
						Number: 4,
						Column: &plugin.Column{
							Name:    "price",
							Type:    &plugin.Identifier{Name: "numeric"},
							NotNull: true, // Required
						},
					},
					{
						Number: 5,
						Column: &plugin.Column{
							Name:    "isbn",
							Type:    &plugin.Identifier{Name: "varchar"},
							NotNull: false, // Nullable - should come after required params
						},
					},
					{
						Number: 6,
						Column: &plugin.Column{
							Name:    "published",
							Type:    &plugin.Identifier{Name: "date"},
							NotNull: false, // Nullable - should come after required params
						},
					},
				},
				Columns: []*plugin.Column{
					{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
					{Name: "author_id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
					{Name: "title", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
					{Name: "description", Type: &plugin.Identifier{Name: "text"}, NotNull: false},
					{Name: "price", Type: &plugin.Identifier{Name: "numeric"}, NotNull: true},
					{Name: "isbn", Type: &plugin.Identifier{Name: "varchar"}, NotNull: false},
					{Name: "published", Type: &plugin.Identifier{Name: "date"}, NotNull: false},
				},
			},
		},
	}

	gen := NewGenerator(req, "db", GeneratorOptions{})

	resp, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(resp.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(resp.Files))
	}

	queriesContent := string(resp.Files[0].Contents)

	// Check that required parameters come first, followed by optional parameters with defaults
	// Required: author_id, title, price
	// Optional: description, isbn, published (all with = nil)
	expectedSignature := "def create_book(author_id : Int32, title : String, price : Float64, description : String? = nil, isbn : String? = nil, published : Time? = nil) : CreateBookRow?"
	if !strings.Contains(queriesContent, expectedSignature) {
		t.Errorf("Expected method signature with correct parameter ordering:\n%s\nGot:\n%s", expectedSignature, queriesContent)
	}

	// Check that the parameter order in the actual query call matches original SQL order
	// Original SQL order: $1=author_id, $2=title, $3=description, $4=price, $5=isbn, $6=published
	expectedCallOrder := "author_id, title, description, price, isbn, published,"
	if !strings.Contains(queriesContent, expectedCallOrder) {
		t.Errorf("Expected parameters in original SQL order in query call: %s\nGot:\n%s", expectedCallOrder, queriesContent)
	}
}

func TestAllRequiredParameters(t *testing.T) {
	req := &plugin.GenerateRequest{
		Settings: &plugin.Settings{
			Engine: "postgresql",
		},
		Queries: []*plugin.Query{
			{
				Name: "UpdateAuthor",
				Text: "UPDATE authors SET name = $2, email = $3 WHERE id = $1",
				Cmd:  ":exec",
				Params: []*plugin.Parameter{
					{
						Number: 1,
						Column: &plugin.Column{
							Name:    "id",
							Type:    &plugin.Identifier{Name: "int4"},
							NotNull: true,
						},
					},
					{
						Number: 2,
						Column: &plugin.Column{
							Name:    "name",
							Type:    &plugin.Identifier{Name: "text"},
							NotNull: true,
						},
					},
					{
						Number: 3,
						Column: &plugin.Column{
							Name:    "email",
							Type:    &plugin.Identifier{Name: "text"},
							NotNull: true,
						},
					},
				},
			},
		},
	}

	gen := NewGenerator(req, "db", GeneratorOptions{})

	resp, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(resp.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(resp.Files))
	}

	queriesContent := string(resp.Files[0].Contents)

	// When all parameters are required, none should have default values
	expectedSignature := "def update_author(id : Int32, name : String, email : String) : Nil"
	if !strings.Contains(queriesContent, expectedSignature) {
		t.Errorf("Expected method signature without default values for all required parameters:\n%s\nGot:\n%s", expectedSignature, queriesContent)
	}

	// Should not contain any " = nil" for required parameters
	if strings.Contains(queriesContent, "String = nil") || strings.Contains(queriesContent, "Int32 = nil") {
		t.Error("Required parameters should not have default nil values")
	}
}

func TestAllOptionalParameters(t *testing.T) {
	req := &plugin.GenerateRequest{
		Settings: &plugin.Settings{
			Engine: "postgresql",
		},
		Queries: []*plugin.Query{
			{
				Name: "SearchAuthors",
				Text: "SELECT * FROM authors WHERE name ILIKE $1 AND bio ILIKE $2 AND location ILIKE $3",
				Cmd:  ":many",
				Params: []*plugin.Parameter{
					{
						Number: 1,
						Column: &plugin.Column{
							Name:    "name",
							Type:    &plugin.Identifier{Name: "text"},
							NotNull: false,
						},
					},
					{
						Number: 2,
						Column: &plugin.Column{
							Name:    "bio",
							Type:    &plugin.Identifier{Name: "text"},
							NotNull: false,
						},
					},
					{
						Number: 3,
						Column: &plugin.Column{
							Name:    "location",
							Type:    &plugin.Identifier{Name: "text"},
							NotNull: false,
						},
					},
				},
				Columns: []*plugin.Column{
					{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
					{Name: "name", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
					{Name: "bio", Type: &plugin.Identifier{Name: "text"}, NotNull: false},
					{Name: "location", Type: &plugin.Identifier{Name: "text"}, NotNull: false},
				},
			},
		},
	}

	gen := NewGenerator(req, "db", GeneratorOptions{})

	resp, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(resp.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(resp.Files))
	}

	queriesContent := string(resp.Files[0].Contents)

	// When all parameters are optional, all should have default nil values
	expectedSignature := "def search_authors(name : String? = nil, bio : String? = nil, location : String? = nil) : Array(SearchAuthorsRow)"
	if !strings.Contains(queriesContent, expectedSignature) {
		t.Errorf("Expected method signature with default nil for all optional parameters:\n%s\nGot:\n%s", expectedSignature, queriesContent)
	}
}

func TestParamListFunction(t *testing.T) {
	tests := []struct {
		name     string
		params   []crystalParam
		expected string
	}{
		{
			name:     "empty parameters",
			params:   []crystalParam{},
			expected: "",
		},
		{
			name: "all required parameters",
			params: []crystalParam{
				{Name: "id", Type: "Int32", Position: 1},
				{Name: "name", Type: "String", Position: 2},
			},
			expected: "id : Int32, name : String",
		},
		{
			name: "all nullable parameters",
			params: []crystalParam{
				{Name: "bio", Type: "String?", Position: 1},
				{Name: "location", Type: "String?", Position: 2},
			},
			expected: "bio : String? = nil, location : String? = nil",
		},
		{
			name: "mixed parameters",
			params: []crystalParam{
				{Name: "name", Type: "String", Position: 1},
				{Name: "bio", Type: "String?", Position: 2},
				{Name: "age", Type: "Int32", Position: 3},
				{Name: "location", Type: "String?", Position: 4},
			},
			expected: "name : String, bio : String? = nil, age : Int32, location : String? = nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := paramList(tt.params)
			if result != tt.expected {
				t.Errorf("paramList() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestParamNamesFunction(t *testing.T) {
	tests := []struct {
		name     string
		params   []crystalParam
		expected string
	}{
		{
			name:     "empty parameters",
			params:   []crystalParam{},
			expected: "",
		},
		{
			name: "parameters in order",
			params: []crystalParam{
				{Name: "name", Type: "String", Position: 1},
				{Name: "bio", Type: "String?", Position: 2},
				{Name: "age", Type: "Int32", Position: 3},
			},
			expected: "name, bio, age",
		},
		{
			name: "parameters out of order (should be sorted by position)",
			params: []crystalParam{
				{Name: "age", Type: "Int32", Position: 3},
				{Name: "name", Type: "String", Position: 1},
				{Name: "bio", Type: "String?", Position: 2},
			},
			expected: "name, bio, age",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := paramNames(tt.params)
			if result != tt.expected {
				t.Errorf("paramNames() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestBooleanQuestionGetters(t *testing.T) {
	req := &plugin.GenerateRequest{
		Settings: &plugin.Settings{
			Engine: "postgresql",
		},
		Catalog: &plugin.Catalog{
			Schemas: []*plugin.Schema{
				{
					Name: "public",
					Tables: []*plugin.Table{
						{
							Rel: &plugin.Identifier{
								Name: "users",
							},
							Columns: []*plugin.Column{
								{
									Name:    "id",
									Type:    &plugin.Identifier{Name: "int4"},
									NotNull: true,
								},
								{
									Name:    "name",
									Type:    &plugin.Identifier{Name: "text"},
									NotNull: true,
								},
								{
									Name:    "is_active",
									Type:    &plugin.Identifier{Name: "bool"},
									NotNull: true,
								},
								{
									Name:    "is_verified",
									Type:    &plugin.Identifier{Name: "bool"},
									NotNull: false, // nullable boolean
								},
							},
						},
					},
				},
			},
		},
	}

	t.Run("with boolean question getters disabled", func(t *testing.T) {
		gen := NewGenerator(req, "db", GeneratorOptions{
			EmitBooleanQuestionGetters: false,
		})

		resp, err := gen.Generate(context.Background())
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if len(resp.Files) != 2 {
			t.Fatalf("Expected 2 files, got %d", len(resp.Files))
		}

		modelsContent := string(resp.Files[0].Contents)

		// Should use regular getters for boolean fields
		if !strings.Contains(modelsContent, "getter is_active : Bool") {
			t.Error("Should use regular getter for boolean field when option is disabled")
		}

		if !strings.Contains(modelsContent, "getter is_verified : Bool?") {
			t.Error("Should use regular getter for nullable boolean field when option is disabled")
		}

		// Should not contain question mark getters
		if strings.Contains(modelsContent, "getter? is_active") {
			t.Error("Should not use question mark getter when option is disabled")
		}

		if strings.Contains(modelsContent, "getter? is_verified") {
			t.Error("Should not use question mark getter when option is disabled")
		}
	})

	t.Run("with boolean question getters enabled", func(t *testing.T) {
		gen := NewGenerator(req, "db", GeneratorOptions{
			EmitBooleanQuestionGetters: true,
		})

		resp, err := gen.Generate(context.Background())
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if len(resp.Files) != 2 {
			t.Fatalf("Expected 2 files, got %d", len(resp.Files))
		}

		modelsContent := string(resp.Files[0].Contents)

		// Should use question mark getters for boolean fields
		if !strings.Contains(modelsContent, "getter? is_active : Bool") {
			t.Error("Should use question mark getter for boolean field when option is enabled")
		}

		if !strings.Contains(modelsContent, "getter? is_verified : Bool?") {
			t.Error("Should use question mark getter for nullable boolean field when option is enabled")
		}

		// Should not contain regular getters for boolean fields
		if strings.Contains(modelsContent, "getter is_active : Bool") {
			t.Error("Should not use regular getter for boolean field when option is enabled")
		}

		if strings.Contains(modelsContent, "getter is_verified : Bool?") {
			t.Error("Should not use regular getter for nullable boolean field when option is enabled")
		}

		// Non-boolean fields should still use regular getters
		if !strings.Contains(modelsContent, "getter id : Int32") {
			t.Error("Non-boolean fields should use regular getters")
		}

		if !strings.Contains(modelsContent, "getter name : String") {
			t.Error("Non-boolean fields should use regular getters")
		}
	})
}

func TestIsBooleanType(t *testing.T) {
	tests := []struct {
		name         string
		crystalType  string
		expected     bool
	}{
		{"bool type", "Bool", true},
		{"nullable bool type", "Bool?", true},
		{"string type", "String", false},
		{"nullable string type", "String?", false},
		{"int type", "Int32", false},
		{"nullable int type", "Int32?", false},
		{"empty string", "", false},
		{"boolean lowercase", "bool", false}, // Crystal uses Bool, not bool
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBooleanType(tt.crystalType)
			if result != tt.expected {
				t.Errorf("isBooleanType(%s) = %v, expected %v", tt.crystalType, result, tt.expected)
			}
		})
	}
}
