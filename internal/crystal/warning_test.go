package crystal

import (
	"fmt"
	"testing"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestWarningCollection(t *testing.T) {
	gen := NewGenerator(&plugin.GenerateRequest{}, "test", GeneratorOptions{})

	// Test adding warnings
	gen.addWarning("First warning")
	gen.addWarning("Second warning")

	// Check warnings are collected
	if len(gen.warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(gen.warnings))
	}

	if gen.warnings[0] != "First warning" {
		t.Errorf("Expected 'First warning', got '%s'", gen.warnings[0])
	}

	if gen.warnings[1] != "Second warning" {
		t.Errorf("Expected 'Second warning', got '%s'", gen.warnings[1])
	}
}

func TestParameterDeduplicationWithWarnings(t *testing.T) {
	// Test the warning functionality in the parameter deduplication logic
	gen := NewGenerator(&plugin.GenerateRequest{}, "test", GeneratorOptions{})

	// Simulate parameter processing with duplicates
	query := &plugin.Query{
		Name: "TestQuery",
		Params: []*plugin.Parameter{
			{
				Number: 1,
				Column: &plugin.Column{
					Name:    "created_at",
					Type:    &plugin.Identifier{Name: "timestamp"},
					NotNull: true,
				},
			},
			{
				Number: 2,
				Column: &plugin.Column{
					Name:    "created_at",
					Type:    &plugin.Identifier{Name: "timestamp"},
					NotNull: true,
				},
			},
		},
	}

	// Process parameters using the same logic as in buildCrystalQuery
	usedNames := make(map[string]int)
	for _, param := range query.Params {
		if param.Column != nil && param.Column.Name != "" {
			baseName := toSnakeCase(param.Column.Name)

			// Check for name conflicts and add suffix if needed
			if count, exists := usedNames[baseName]; exists {
				// Name already used, add numeric suffix
				usedNames[baseName] = count + 1

				// Add warning for this query
				gen.addWarning(fmt.Sprintf("Query '%s' has duplicate parameter name '%s'. Consider using sqlc.arg('%s_1') and sqlc.arg('%s_2') or @%s_1 and @%s_2 for explicit naming.",
					query.Name, baseName, baseName, baseName, baseName, baseName))
			} else {
				// First use of this name
				usedNames[baseName] = 1
			}
		}
	}

	// Check that warning was added
	if len(gen.warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(gen.warnings))
	}

	expectedWarning := "Query 'TestQuery' has duplicate parameter name 'created_at'. Consider using sqlc.arg('created_at_1') and sqlc.arg('created_at_2') or @created_at_1 and @created_at_2 for explicit naming."
	if gen.warnings[0] != expectedWarning {
		t.Errorf("Expected warning: %s\nGot: %s", expectedWarning, gen.warnings[0])
	}
}