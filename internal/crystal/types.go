package crystal

import "strings"

// postgresType maps PostgreSQL types to Crystal types
func (g *Generator) postgresType(sqlType string) string {
	switch sqlType {
	// Integer types
	case "int8", "bigint", "bigserial":
		return "Int64"
	case "int4", "int", "integer", "serial":
		return "Int32"
	case "int2", "smallint", "smallserial":
		return "Int16"

	// Floating point types
	case "numeric", "decimal":
		return "Float64"
	case "real", "float4":
		return "Float32"
	case "float8", "double precision":
		return "Float64"

	// Boolean type
	case "bool", "boolean":
		return "Bool"

	// String types
	case "text", "varchar", "char", "bpchar", "citext", "name":
		return "String"

	// Time types
	case "timestamp", "timestamptz", "date", "time", "timetz":
		return "Time"
	case "interval":
		return "Time::Span"

	// UUID type
	case "uuid":
		return "String"

	// JSON types
	case "json", "jsonb":
		return "JSON::Any"

	// Binary type
	case "bytea":
		return "Bytes"

	// Network types
	case "inet", "cidr", "macaddr", "macaddr8":
		return "String"

	// Geometric types
	case "point", "line", "lseg", "box", "path", "polygon", "circle":
		return "String"

	// Money type
	case "money":
		return "Float64"

	// Bit string types
	case "bit", "bit varying", "varbit":
		return "String"

	// Range types
	case "int4range", "int8range", "numrange", "tsrange", "tstzrange", "daterange":
		return "String"

	// Other types
	case "xml":
		return "String"
	case "void":
		return "Nil"

	default:
		// Default to String for unknown types
		return "String"
	}
}

// mysqlType maps MySQL types to Crystal types
func (g *Generator) mysqlType(sqlType string) string {
	switch sqlType {
	// Integer types
	case "bigint":
		return "Int64"
	case "int", "integer", "mediumint":
		return "Int32"
	case "smallint":
		return "Int16"
	case "tinyint":
		return "Int8"

	// Floating point types
	case "decimal", "numeric":
		return "Float64"
	case "float":
		return "Float32"
	case "double", "double precision", "real":
		return "Float64"

	// Boolean type
	case "bit", "bool", "boolean":
		return "Bool"

	// String types
	case "char", "varchar", "text", "tinytext", "mediumtext", "longtext":
		return "String"

	// Time types
	case "datetime", "timestamp":
		return "Time"
	case "date":
		return "Time"
	case "time":
		return "Time::Span"
	case "year":
		return "Int32"

	// JSON type
	case "json":
		return "JSON::Any"

	// Binary types
	case "binary", "varbinary", "blob", "tinyblob", "mediumblob", "longblob":
		return "Bytes"

	// Enum and set
	case "enum", "set":
		return "String"

	default:
		// Default to String for unknown types
		return "String"
	}
}

// sqliteType maps SQLite types to Crystal types
func (g *Generator) sqliteType(sqlType string) string {
	// SQLite uses type affinity, so we need to be more flexible
	sqlType = normalizeSQLiteType(sqlType)

	switch {
	// Integer affinity
	case isIntegerType(sqlType):
		return "Int64"

	// Real affinity
	case isRealType(sqlType):
		return "Float64"

	// Text affinity
	case isTextType(sqlType):
		return "String"

	// Blob affinity
	case isBlobType(sqlType):
		return "Bytes"

	// Numeric affinity (can be integer or real)
	case isNumericType(sqlType):
		return "Float64"

	// Boolean (SQLite doesn't have a native boolean type)
	case sqlType == "boolean" || sqlType == "bool":
		return "Bool"

	// Date/time types (stored as text, integer, or real in SQLite)
	case isDateTimeType(sqlType):
		return "Time"

	default:
		// Default to String for unknown types
		return "String"
	}
}

// Helper functions for SQLite type affinity

func normalizeSQLiteType(sqlType string) string {
	// Remove parentheses and contents (e.g., "VARCHAR(255)" -> "VARCHAR")
	if idx := strings.Index(sqlType, "("); idx != -1 {
		sqlType = sqlType[:idx]
	}
	return strings.ToLower(strings.TrimSpace(sqlType))
}

func isIntegerType(sqlType string) bool {
	return strings.Contains(sqlType, "int") ||
		sqlType == "integer" ||
		sqlType == "tinyint" ||
		sqlType == "smallint" ||
		sqlType == "mediumint" ||
		sqlType == "bigint" ||
		sqlType == "int2" ||
		sqlType == "int8"
}

func isRealType(sqlType string) bool {
	return sqlType == "real" ||
		sqlType == "double" ||
		sqlType == "double precision" ||
		sqlType == "float"
}

func isTextType(sqlType string) bool {
	return strings.Contains(sqlType, "char") ||
		strings.Contains(sqlType, "clob") ||
		strings.Contains(sqlType, "text") ||
		sqlType == "varchar" ||
		sqlType == "varying character" ||
		sqlType == "nchar" ||
		sqlType == "native character" ||
		sqlType == "nvarchar"
}

func isBlobType(sqlType string) bool {
	return sqlType == "blob" || sqlType == ""
}

func isNumericType(sqlType string) bool {
	return sqlType == "numeric" ||
		sqlType == "decimal" ||
		strings.Contains(sqlType, "decimal")
}

func isDateTimeType(sqlType string) bool {
	return sqlType == "date" ||
		sqlType == "datetime" ||
		sqlType == "timestamp" ||
		sqlType == "time"
}
