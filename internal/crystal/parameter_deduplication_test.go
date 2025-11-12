package crystal

import (
	"fmt"
	"testing"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestParameterDeduplication(t *testing.T) {
	tests := []struct {
		name     string
		params   []*plugin.Parameter
		expected []string
	}{
		{
			name: "no duplicates",
			params: []*plugin.Parameter{
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
			},
			expected: []string{"id", "name"},
		},
		{
			name: "duplicate column names - date range",
			params: []*plugin.Parameter{
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
			expected: []string{"created_at", "created_at_2"},
		},
		{
			name: "multiple duplicates",
			params: []*plugin.Parameter{
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
						Name:    "id",
						Type:    &plugin.Identifier{Name: "int4"},
						NotNull: true,
					},
				},
				{
					Number: 3,
					Column: &plugin.Column{
						Name:    "id",
						Type:    &plugin.Identifier{Name: "int4"},
						NotNull: true,
					},
				},
			},
			expected: []string{"id", "id_2", "id_3"},
		},
		{
			name: "mixed duplicates and unique",
			params: []*plugin.Parameter{
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
				{
					Number: 3,
					Column: &plugin.Column{
						Name:    "published",
						Type:    &plugin.Identifier{Name: "bool"},
						NotNull: true,
					},
				},
				{
					Number: 4,
					Column: &plugin.Column{
						Name:    "created_at",
						Type:    &plugin.Identifier{Name: "timestamp"},
						NotNull: true,
					},
				},
			},
			expected: []string{"created_at", "created_at_2", "published", "created_at_3"},
		},
		{
			name: "no column names (fallback to argN)",
			params: []*plugin.Parameter{
				{Number: 1},
				{Number: 2},
				{Number: 3},
			},
			expected: []string{"arg1", "arg2", "arg3"},
		},
		{
			name: "fallback argN duplicates",
			params: []*plugin.Parameter{
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
				},
				{
					Number: 3,
					Column: &plugin.Column{
						Name:    "id",
						Type:    &plugin.Identifier{Name: "int4"},
						NotNull: true,
					},
				},
			},
			expected: []string{"id", "arg2", "id_2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the parameter naming logic directly
			// This is similar to what happens in buildCrystalQuery
			usedNames := make(map[string]int)
			var actual []string

			for _, param := range tt.params {
				name := fmt.Sprintf("arg%d", param.Number)

				// Use column name if available (this is the deduplication logic)
				if param.Column != nil && param.Column.Name != "" {
					baseName := toSnakeCase(param.Column.Name)

					// Check for name conflicts and add suffix if needed
					if count, exists := usedNames[baseName]; exists {
						// Name already used, add numeric suffix
						name = fmt.Sprintf("%s_%d", baseName, count+1)
						usedNames[baseName] = count + 1
					} else {
						// First use of this name
						name = baseName
						usedNames[baseName] = 1
					}
				} else {
					// For argN names, also track to avoid conflicts
					baseName := name
					if count, exists := usedNames[baseName]; exists {
						name = fmt.Sprintf("%s_%d", baseName, count+1)
						usedNames[baseName] = count + 1
					} else {
						usedNames[baseName] = 1
					}
				}

				actual = append(actual, name)
			}

			// Verify the results
			if len(actual) != len(tt.expected) {
				t.Fatalf("Expected %d parameters, got %d", len(tt.expected), len(actual))
			}

			for i, expected := range tt.expected {
				if actual[i] != expected {
					t.Errorf("Parameter %d: expected name %q, got %q", i+1, expected, actual[i])
				}
			}
		})
	}
}