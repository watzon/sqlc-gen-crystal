package crystal

import "testing"

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GetAuthor", "get_author"},
		{"ListAuthors", "list_authors"},
		{"CreateAuthor", "create_author"},
		{"HTTPRequest", "http_request"},
		{"XMLParser", "xml_parser"},
		{"IOError", "io_error"},
		{"get_author", "get_author"}, // Already snake_case
		{"GetAuthorByID", "get_author_by_id"},
		{"ID", "id"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"get_author", "GetAuthor"},
		{"list_authors", "ListAuthors"},
		{"create_author", "CreateAuthor"},
		{"http_request", "HttpRequest"},
		{"xml_parser", "XmlParser"},
		{"io_error", "IoError"},
		{"get_author_by_id", "GetAuthorById"},
		{"id", "Id"},
		{"ID", "ID"},
		{"GetAuthor", "GetAuthor"},  // Input is already PascalCase
		{"get-author", "GetAuthor"}, // Hyphenated
		{"get author", "GetAuthor"}, // Space separated
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToConstantCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GetAuthor", "GET_AUTHOR"},
		{"ListAuthors", "LIST_AUTHORS"},
		{"CreateAuthor", "CREATE_AUTHOR"},
		{"HTTPRequest", "HTTP_REQUEST"},
		{"XMLParser", "XML_PARSER"},
		{"IOError", "IO_ERROR"},
		{"get_author", "GET_AUTHOR"},
		{"GetAuthorByID", "GET_AUTHOR_BY_ID"},
		{"ID", "ID"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toConstantCase(tt.input)
			if result != tt.expected {
				t.Errorf("toConstantCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSingularize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Simple plurals
		{"books", "book"},
		{"authors", "author"},
		{"users", "user"},
		{"posts", "post"},
		
		// -ies endings
		{"companies", "company"},
		{"cities", "city"},
		{"stories", "story"},
		
		// -es endings
		{"boxes", "box"},
		{"classes", "class"},
		{"processes", "process"},
		{"dishes", "dish"},
		{"churches", "church"},
		
		// -ves endings
		{"knives", "knife"},
		{"lives", "life"},
		{"wives", "wife"},
		
		// Irregular plurals
		{"children", "child"},
		{"people", "person"},
		{"men", "man"},
		{"women", "woman"},
		{"feet", "foot"},
		{"teeth", "tooth"},
		{"geese", "goose"},
		{"mice", "mouse"},
		
		// Already singular
		{"book", "book"},
		{"author", "author"},
		{"person", "person"},
		
		// Edge cases
		{"", ""},
		{"a", "a"},
		{"data", "data"}, // No change for ambiguous cases
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := singularize(test.input)
			if result != test.expected {
				t.Errorf("singularize(%q) = %q; want %q", test.input, result, test.expected)
			}
		})
	}
}
