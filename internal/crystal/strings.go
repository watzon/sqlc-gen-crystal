package crystal

import (
	"regexp"
	"strings"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

// toSnakeCase converts a string to snake_case
func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// toPascalCase converts a string to PascalCase
func toPascalCase(str string) string {
	// If the string is already in PascalCase/camelCase, just ensure first letter is upper
	if !strings.Contains(str, "_") && !strings.Contains(str, "-") && !strings.Contains(str, " ") {
		if len(str) > 0 {
			return strings.ToUpper(string(str[0])) + str[1:]
		}
		return str
	}

	// Handle common SQL naming patterns
	str = strings.ReplaceAll(str, "_", " ")
	str = strings.ReplaceAll(str, "-", " ")

	// Title case each word
	words := strings.Fields(str)
	for i, word := range words {
		if len(word) > 0 {
			// Capitalize first letter, lowercase the rest
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, "")
}

// toConstantCase converts a string to CONSTANT_CASE
func toConstantCase(str string) string {
	// First convert to snake case
	snake := toSnakeCase(str)
	// Then uppercase
	return strings.ToUpper(snake)
}

// singularize attempts to convert a plural word to singular
// This is a simple implementation for common English pluralization rules
func singularize(word string) string {
	if len(word) <= 1 {
		return word
	}

	lower := strings.ToLower(word)
	
	// Handle irregular plurals
	irregulars := map[string]string{
		"children": "child",
		"people":   "person",
		"men":      "man",
		"women":    "woman",
		"feet":     "foot",
		"teeth":    "tooth",
		"geese":    "goose",
		"mice":     "mouse",
		"dice":     "die",
	}
	
	if singular, exists := irregulars[lower]; exists {
		// Preserve original casing
		if strings.ToUpper(word) == word {
			return strings.ToUpper(singular)
		} else if strings.Title(word) == word {
			return strings.Title(singular)
		}
		return singular
	}
	
	// Handle common suffixes
	if strings.HasSuffix(lower, "ies") && len(word) > 3 {
		// companies -> company, cities -> city
		return word[:len(word)-3] + "y"
	}
	
	if strings.HasSuffix(lower, "ves") && len(word) > 3 {
		// knives -> knife, lives -> life
		return word[:len(word)-3] + "fe"
	}
	
	if strings.HasSuffix(lower, "ses") && len(word) > 3 {
		// classes -> class, processes -> process  
		return word[:len(word)-2]
	}
	
	if strings.HasSuffix(lower, "es") && len(word) > 2 {
		// Check if it's a simple -es plural (boxes -> box)
		beforeEs := word[:len(word)-2]
		lastChar := strings.ToLower(string(beforeEs[len(beforeEs)-1]))
		if lastChar == "x" || lastChar == "s" || lastChar == "z" || 
		   strings.HasSuffix(strings.ToLower(beforeEs), "ch") || 
		   strings.HasSuffix(strings.ToLower(beforeEs), "sh") {
			return beforeEs
		}
	}
	
	if strings.HasSuffix(lower, "s") && len(word) > 1 {
		// Simple plural: books -> book, authors -> author
		return word[:len(word)-1]
	}
	
	// No change needed
	return word
}

// crystalModuleName converts a module name to proper Crystal module syntax
// Handles nested modules like "TestApp::Database" and ensures PascalCase
func crystalModuleName(name string) string {
	if name == "" {
		return "Db"
	}
	
	// Handle nested modules (separated by :: or .)
	var parts []string
	if strings.Contains(name, "::") {
		parts = strings.Split(name, "::")
	} else if strings.Contains(name, ".") {
		parts = strings.Split(name, ".")
	} else {
		parts = []string{name}
	}
	
	// Convert each part to PascalCase
	for i, part := range parts {
		parts[i] = toPascalCase(part)
	}
	
	return strings.Join(parts, "::")
}
