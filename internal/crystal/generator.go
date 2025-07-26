package crystal

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// GeneratorOptions configures the Crystal code generator
type GeneratorOptions struct {
	EmitJSONTags              bool
	EmitDBTags                bool
	EmitResultStructPointers  bool
	GenerateConnectionManager bool
	GenerateRepositories      bool
	EmitBooleanQuestionGetters bool
}

// Generator generates Crystal code from SQL queries
type Generator struct {
	req                *plugin.GenerateRequest
	pkg                string
	options            GeneratorOptions
	signatureToStruct  map[string]string // Maps field signatures to struct names
}

// NewGenerator creates a new Crystal code generator
func NewGenerator(req *plugin.GenerateRequest, pkg string, options GeneratorOptions) *Generator {
	return &Generator{
		req:               req,
		pkg:               pkg,
		options:           options,
		signatureToStruct: make(map[string]string),
	}
}

// Generate generates Crystal code from the SQL queries
func (g *Generator) Generate(ctx context.Context) (*plugin.GenerateResponse, error) {
	var resp plugin.GenerateResponse

	// Generate models if there are any tables
	if g.req.Catalog != nil && len(g.req.Catalog.Schemas) > 0 {
		modelsFile, err := g.generateModels()
		if err != nil {
			return nil, fmt.Errorf("failed to generate models: %w", err)
		}
		if modelsFile != nil {
			resp.Files = append(resp.Files, modelsFile)
		}
	}

	// Generate queries if there are any
	if len(g.req.Queries) > 0 {
		queriesFile, err := g.generateQueries()
		if err != nil {
			return nil, fmt.Errorf("failed to generate queries: %w", err)
		}
		if queriesFile != nil {
			resp.Files = append(resp.Files, queriesFile)
		}
	}

	// Generate connection manager if enabled
	if g.options.GenerateConnectionManager {
		connMgrFile, err := g.generateConnectionManager()
		if err != nil {
			return nil, fmt.Errorf("failed to generate connection manager: %w", err)
		}
		if connMgrFile != nil {
			resp.Files = append(resp.Files, connMgrFile)
		}
	}

	// Generate repositories if enabled
	if g.options.GenerateRepositories {
		repoFiles, err := g.generateRepositories()
		if err != nil {
			return nil, fmt.Errorf("failed to generate repositories: %w", err)
		}
		resp.Files = append(resp.Files, repoFiles...)
	}

	return &resp, nil
}

// generateModels generates the models.cr file
func (g *Generator) generateModels() (*plugin.File, error) {
	// Collect all unique structs, prioritizing table-based structs over query-specific ones
	structs := make(map[string]*crystalStruct)
	// Track field signatures to detect duplicates
	fieldSignatures := make(map[string]string)

	// First pass: Add structs for tables (prioritize table-based names)
	for _, schema := range g.req.Catalog.Schemas {
		// Skip information_schema and pg_catalog schemas
		if schema.Name == "information_schema" || schema.Name == "pg_catalog" {
			continue
		}

		for _, table := range schema.Tables {
			// Skip system tables
			if strings.HasPrefix(table.Rel.Name, "pg_") ||
				strings.HasPrefix(table.Rel.Name, "sql_") ||
				table.Rel.Schema == "information_schema" ||
				table.Rel.Schema == "pg_catalog" {
				continue
			}

			structName := toPascalCase(singularize(table.Rel.Name))

			cs := &crystalStruct{
				Name:      structName,
				TableName: table.Rel.Name,
			}

			var fieldSig strings.Builder
			for _, col := range table.Columns {
				field := crystalField{
					Name:   toSnakeCase(col.Name),
					DBName: col.Name,
					Type:   g.crystalType(col),
				}

				if g.options.EmitJSONTags {
					field.JSONName = col.Name
				}

				cs.Fields = append(cs.Fields, field)
				// Build field signature for deduplication
				fieldSig.WriteString(fmt.Sprintf("%s:%s;", field.Name, field.Type))
			}

			if len(cs.Fields) > 0 {
				structs[structName] = cs
				fieldSignatures[fieldSig.String()] = structName
				g.signatureToStruct[fieldSig.String()] = structName
			}
		}
	}

	// Second pass: Add structs from query results, but only if they're unique
	for _, query := range g.req.Queries {
		if len(query.Columns) == 0 {
			continue
		}

		// Skip queries that don't return data
		switch query.Cmd {
		case ":exec", ":execresult", ":execrows", ":copyfrom":
			continue
		}

		// Group columns by table to detect embeds
		tableColumns := make(map[string][]*plugin.Column)
		var standaloneColumns []*plugin.Column
		
		// Check if any column has EmbedTable set - if so, this is an embed query
		hasEmbeds := false
		for _, col := range query.Columns {
			if col.EmbedTable != nil {
				hasEmbeds = true
				// Each embedded column represents ALL columns from that table
				tableColumns[col.EmbedTable.Name] = append(tableColumns[col.EmbedTable.Name], col)
			} else if col.Table != nil && col.Table.Name != "" {
				// Regular column with table info
				key := col.Table.Name
				if col.TableAlias != "" {
					key = col.TableAlias
				}
				tableColumns[key] = append(tableColumns[key], col)
			} else {
				// Standalone column (like COUNT, etc)
				standaloneColumns = append(standaloneColumns, col)
			}
		}
		
		// Use embeds if we found any EmbedTable columns
		usesEmbeds := hasEmbeds
		
		
		// Build struct fields
		var fieldSig strings.Builder
		var fields []crystalField
		
		if usesEmbeds {
			// This is a query with embeds - create embedded struct fields
			for tableName, cols := range tableColumns {
				// For embed columns, the column itself has the embed info
				if len(cols) > 0 && cols[0].EmbedTable != nil {
					// Find the table struct name
					var structName string
					if g.req.Catalog != nil {
						for _, schema := range g.req.Catalog.Schemas {
							for _, table := range schema.Tables {
								if table.Rel.Name == tableName {
									structName = toPascalCase(singularize(table.Rel.Name))
									break
								}
							}
							if structName != "" {
								break
							}
						}
					}
					
					if structName == "" {
						structName = toPascalCase(singularize(tableName))
					}
					
					// For embedded tables in LEFT/RIGHT JOINs, check nullability from SQL
					nullable := strings.Contains(strings.ToUpper(query.Text), "LEFT JOIN") ||
						      strings.Contains(strings.ToUpper(query.Text), "RIGHT JOIN")
					
					fieldType := structName
					if nullable {
						// Check if this specific table is on the nullable side
						// For LEFT JOIN, right table is nullable
						// For INNER JOIN, neither is nullable
						if strings.Contains(strings.ToUpper(query.Text), "LEFT JOIN") {
							// In "FROM authors LEFT JOIN books", books is nullable
							if !strings.Contains(strings.ToUpper(query.Text), "FROM " + strings.ToUpper(tableName)) {
								fieldType += "?"
							}
						}
					}
					
					field := crystalField{
						Name:   toSnakeCase(singularize(tableName)),
						DBName: toSnakeCase(singularize(tableName)), // Set DBName to avoid empty key
						Type:   fieldType,
					}
					fields = append(fields, field)
					fieldSig.WriteString(fmt.Sprintf("%s:%s;", field.Name, field.Type))
				}
			}
			
			// Add standalone columns (like COUNT, etc)
			for _, col := range standaloneColumns {
				field := crystalField{
					Name:   toSnakeCase(col.Name),
					DBName: col.Name,
					Type:   g.crystalType(col),
				}

				if g.options.EmitJSONTags {
					field.JSONName = col.Name
				}

				fields = append(fields, field)
				fieldSig.WriteString(fmt.Sprintf("%s:%s;", field.Name, field.Type))
			}
		} else {
			// Regular query - create normal fields
			for _, col := range query.Columns {
				field := crystalField{
					Name:   toSnakeCase(col.Name),
					DBName: col.Name,
					Type:   g.crystalType(col),
				}

				if g.options.EmitJSONTags {
					field.JSONName = col.Name
				}

				fields = append(fields, field)
				fieldSig.WriteString(fmt.Sprintf("%s:%s;", field.Name, field.Type))
			}
		}

		signature := fieldSig.String()

		// Check if we already have a struct with this exact field signature
		if _, exists := fieldSignatures[signature]; exists {
			// Reuse existing struct - this query will use the existing struct
			continue
		}

		// This is a unique struct, create it
		structName := toPascalCase(query.Name)
		if query.Cmd == ":one" || query.Cmd == ":many" {
			structName = structName + "Row"
		}

		cs := &crystalStruct{
			Name:   structName,
			Fields: fields,
		}

		structs[structName] = cs
		fieldSignatures[signature] = structName
		g.signatureToStruct[signature] = structName
	}

	if len(structs) == 0 {
		return nil, nil
	}

	// Sort structs by name for consistent output
	var structList []*crystalStruct
	for _, s := range structs {
		structList = append(structList, s)
	}
	sort.Slice(structList, func(i, j int) bool {
		return structList[i].Name < structList[j].Name
	})

	// Generate the models file
	var buf bytes.Buffer
	err := modelsTemplate.Execute(&buf, templateData{
		Package:                   g.pkg,
		Structs:                   structList,
		EmitJSONTags:              g.options.EmitJSONTags,
		EmitDBTags:                g.options.EmitDBTags,
		EmitBooleanQuestionGetters: g.options.EmitBooleanQuestionGetters,
	})
	if err != nil {
		return nil, err
	}

	return &plugin.File{
		Name:     "models.cr",
		Contents: buf.Bytes(),
	}, nil
}

// getStructNameForQuery determines the appropriate struct name for a query based on field matching
func (g *Generator) getStructNameForQuery(query *plugin.Query) string {
	if len(query.Columns) <= 1 {
		return "" // Single column queries don't need structs
	}

	// Build field signature for this query
	var fieldSig strings.Builder
	for _, col := range query.Columns {
		field := crystalField{
			Name:   toSnakeCase(col.Name),
			DBName: col.Name,
			Type:   g.crystalType(col),
		}
		fieldSig.WriteString(fmt.Sprintf("%s:%s;", field.Name, field.Type))
	}
	signature := fieldSig.String()

	// Check if we have a struct with this signature from the deduplication process
	if structName, exists := g.signatureToStruct[signature]; exists {
		return structName
	}

	// Fallback: generate query-specific name (shouldn't happen if models were generated first)
	structName := toPascalCase(query.Name)
	if query.Cmd == ":one" || query.Cmd == ":many" {
		structName = structName + "Row"
	}
	return structName
}

// generateQueries generates the queries.cr file
func (g *Generator) generateQueries() (*plugin.File, error) {
	var queries []crystalQuery

	for _, query := range g.req.Queries {
		cq := crystalQuery{
			Name:         toSnakeCase(query.Name),
			SQL:          query.Text,
			SourceName:   query.Name,
			Cmd:          query.Cmd,
			Comments:     query.Comments,
			ConstantName: toConstantCase(query.Name),
		}

		// Build parameter list
		hasSlice := false
		for _, param := range query.Params {
			p := crystalParam{
				Name:     fmt.Sprintf("arg%d", param.Number),
				Type:     g.crystalType(param.Column),
				Position: int(param.Number),
			}

			// Try to get a better name from the column
			if param.Column.Name != "" {
				p.Name = toSnakeCase(param.Column.Name)
			}
			
			// Check if this parameter is used with sqlc.slice()
			if param.Column.IsSqlcSlice {
				// Force the type to be an array
				baseType := g.baseType(param.Column)
				p.Type = fmt.Sprintf("Array(%s)", baseType)
				hasSlice = true
				
				// Add to slice params for template processing
				cq.SliceParams = append(cq.SliceParams, sqlcSliceParam{
					Name:        p.Name,
					Placeholder: fmt.Sprintf("$%d", param.Number),
				})
			}

			cq.Params = append(cq.Params, p)
		}
		
		// Mark query as using sqlc.slice() if any param has it
		if hasSlice {
			cq.UsesSQLCSlice = true
		}

		// Sort parameters to put required parameters before optional ones
		// Required parameters are those without '?' suffix (non-nullable)
		// Optional parameters are those with '?' suffix (nullable)
		sort.SliceStable(cq.Params, func(i, j int) bool {
			iIsOptional := strings.HasSuffix(cq.Params[i].Type, "?")
			jIsOptional := strings.HasSuffix(cq.Params[j].Type, "?")

			// If one is optional and the other isn't, required comes first
			if iIsOptional != jIsOptional {
				return !iIsOptional // required (non-optional) comes first
			}

			// If both are the same type (both required or both optional),
			// maintain original order by position
			return cq.Params[i].Position < cq.Params[j].Position
		})

		// Determine return type using deduplicated struct names
		switch query.Cmd {
		case ":one":
			if len(query.Columns) == 1 {
				cq.ReturnType = g.crystalType(query.Columns[0])
			} else {
				cq.ReturnType = g.getStructNameForQuery(query)
			}
			cq.ReturnType += "?"
		case ":many":
			if len(query.Columns) == 1 {
				cq.ReturnType = "Array(" + g.crystalType(query.Columns[0]) + ")"
				cq.SingleColumnType = g.crystalType(query.Columns[0])
			} else {
				cq.ReturnType = "Array(" + g.getStructNameForQuery(query) + ")"
			}
		case ":exec":
			cq.ReturnType = "Nil"
		case ":execresult":
			cq.ReturnType = "DB::ExecResult"
		case ":execrows":
			cq.ReturnType = "Int64"
		case ":execlastid":
			cq.ReturnType = "Int64"
		case ":copyfrom":
			cq.ReturnType = "Int64"  // Will return number of rows copied
		}

		// Set the result struct name if needed
		if len(query.Columns) > 1 && (query.Cmd == ":one" || query.Cmd == ":many") {
			// Use deduplicated struct name
			cq.ResultStruct = g.getStructNameForQuery(query)
		} else if len(query.Columns) == 1 && (query.Cmd == ":one" || query.Cmd == ":many") {
			// For single column queries, store the actual type (without ? or Array)
			if query.Cmd == ":one" {
				cq.SingleColumnType = g.crystalType(query.Columns[0])
			}
		}

		queries = append(queries, cq)
	}

	// Generate the queries file
	var buf bytes.Buffer
	err := queriesTemplate.Execute(&buf, templateData{
		Package: g.pkg,
		Queries: queries,
		Engine:  g.req.Settings.Engine,
	})
	if err != nil {
		return nil, err
	}

	return &plugin.File{
		Name:     "queries.cr",
		Contents: buf.Bytes(),
	}, nil
}

// crystalType converts a SQL column to a Crystal type
func (g *Generator) crystalType(col *plugin.Column) string {
	typ := g.baseType(col)

	if col.IsArray {
		typ = fmt.Sprintf("Array(%s)", typ)
	}

	if !col.NotNull && !col.IsArray {
		if g.options.EmitResultStructPointers {
			typ = typ + "*"
		} else {
			typ = typ + "?"
		}
	}

	return typ
}

// baseType returns the base Crystal type for a SQL type
func (g *Generator) baseType(col *plugin.Column) string {
	// Get the base SQL type name
	typeName := ""
	if col.Type != nil {
		typeName = strings.ToLower(col.Type.Name)
	}

	// Handle engine-specific type mappings
	switch g.req.Settings.Engine {
	case "postgresql":
		return g.postgresType(typeName)
	case "mysql":
		return g.mysqlType(typeName)
	case "sqlite":
		return g.sqliteType(typeName)
	default:
		// Fallback to postgres types
		return g.postgresType(typeName)
	}
}

// Type definitions for templates
type crystalStruct struct {
	Name      string
	TableName string
	Fields    []crystalField
}

type crystalField struct {
	Name     string
	DBName   string
	JSONName string
	Type     string
}

type crystalQuery struct {
	Name             string
	SQL              string
	SourceName       string
	Cmd              string
	Comments         []string
	ConstantName     string
	Params           []crystalParam
	ReturnType       string
	ResultStruct     string
	SingleColumnType string
	UsesSQLCSlice    bool
	SliceParams      []sqlcSliceParam
}

type crystalParam struct {
	Name     string
	Type     string
	Position int
}

type sqlcSliceParam struct {
	Name        string
	Placeholder string
}

type templateData struct {
	Package                   string
	Structs                   []*crystalStruct
	Queries                   []crystalQuery
	EmitJSONTags              bool
	EmitDBTags                bool
	EmitBooleanQuestionGetters bool
	Engine                    string
}

// generateConnectionManager generates the database.cr file with connection management
func (g *Generator) generateConnectionManager() (*plugin.File, error) {
	tmpl, err := template.New("connectionManager").Funcs(template.FuncMap{
		"crystalModule": crystalModuleName,
	}).Parse(connectionManagerTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection manager template: %w", err)
	}

	data := struct {
		Package string
	}{
		Package: g.pkg,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute connection manager template: %w", err)
	}

	return &plugin.File{
		Name:     "database.cr",
		Contents: buf.Bytes(),
	}, nil
}

// generateRepositories generates repository files for each table
func (g *Generator) generateRepositories() ([]*plugin.File, error) {
	var files []*plugin.File

	// Group queries by table
	tableQueries := make(map[string][]crystalQuery)
	for _, q := range g.req.Queries {
		crystalQ := g.buildCrystalQuery(q)

		// Try to determine the primary table from the query
		tableName := g.extractTableName(q)
		if tableName != "" {
			tableQueries[tableName] = append(tableQueries[tableName], crystalQ)
		}
	}

	// Generate a repository for each table
	for tableName, queries := range tableQueries {
		file, err := g.generateRepositoryForTable(tableName, queries)
		if err != nil {
			return nil, fmt.Errorf("failed to generate repository for %s: %w", tableName, err)
		}
		files = append(files, file)
	}

	return files, nil
}

// generateRepositoryForTable generates a repository file for a specific table
func (g *Generator) generateRepositoryForTable(tableName string, queries []crystalQuery) (*plugin.File, error) {
	tmpl, err := template.New("repository").Funcs(template.FuncMap{
		"paramList":     paramList,
		"paramNames":    paramNames,
		"crystalModule": crystalModuleName,
	}).Parse(repositoryTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repository template: %w", err)
	}

	// Filter and adapt queries for this table
	var methods []struct {
		crystalQuery
		IsTableSpecific bool
		MethodName      string
	}

	for _, q := range queries {
		methods = append(methods, struct {
			crystalQuery
			IsTableSpecific bool
			MethodName      string
		}{
			crystalQuery:    q,
			IsTableSpecific: true,
			MethodName:      g.simplifyMethodName(q.Name, tableName),
		})
	}

	data := struct {
		Package   string
		TableName string
		Methods   any
	}{
		Package:   g.pkg,
		TableName: toPascalCase(tableName),
		Methods:   methods,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute repository template: %w", err)
	}

	return &plugin.File{
		Name:     fmt.Sprintf("repositories/%s_repository.cr", toSnakeCase(tableName)),
		Contents: buf.Bytes(),
	}, nil
}

// extractTableName attempts to extract the primary table name from a query
func (g *Generator) extractTableName(q *plugin.Query) string {
	// Simple heuristic: look for table name patterns in the query
	// This is a simplified implementation - could be improved with proper SQL parsing
	sql := strings.ToLower(q.Text)

	// Try to extract from INSERT INTO
	if strings.Contains(sql, "insert into") {
		parts := strings.Split(sql, "insert into")
		if len(parts) > 1 {
			tablePart := strings.TrimSpace(parts[1])
			fields := strings.Fields(tablePart)
			if len(fields) > 0 {
				return strings.Trim(fields[0], "(")
			}
		}
	}

	// Try to extract from UPDATE
	if strings.HasPrefix(strings.TrimSpace(sql), "update ") {
		parts := strings.Split(sql, "update")
		if len(parts) > 1 {
			tablePart := strings.TrimSpace(parts[1])
			fields := strings.Fields(tablePart)
			if len(fields) > 0 {
				return fields[0]
			}
		}
	}

	// Try to extract from DELETE FROM
	if strings.Contains(sql, "delete from") {
		parts := strings.Split(sql, "delete from")
		if len(parts) > 1 {
			tablePart := strings.TrimSpace(parts[1])
			fields := strings.Fields(tablePart)
			if len(fields) > 0 {
				return fields[0]
			}
		}
	}

	// Try to extract from FROM clause in SELECT
	if strings.Contains(sql, "from") {
		parts := strings.Split(sql, "from")
		if len(parts) > 1 {
			tablePart := strings.TrimSpace(parts[1])
			fields := strings.Fields(tablePart)
			if len(fields) > 0 {
				// Remove any alias
				table := fields[0]
				if strings.Contains(table, " as ") {
					table = strings.Split(table, " as ")[0]
				}
				return table
			}
		}
	}

	return ""
}

// simplifyMethodName removes redundant table name from method name
func (g *Generator) simplifyMethodName(methodName, tableName string) string {
	// Convert to snake case for comparison
	snakeMethod := toSnakeCase(methodName)
	snakeTable := toSnakeCase(tableName)
	singularTable := strings.TrimSuffix(snakeTable, "s")

	// Handle specific patterns
	if strings.HasPrefix(snakeMethod, "get_"+singularTable) {
		return "find" + snakeMethod[len("get_"+singularTable):]
	}
	if strings.HasPrefix(snakeMethod, "list_"+snakeTable+"_by_") {
		return "by_" + snakeMethod[len("list_"+snakeTable+"_by_"):]
	}
	if snakeMethod == "list_"+snakeTable {
		return "all"
	}
	if strings.HasPrefix(snakeMethod, "create_"+singularTable) {
		return "create"
	}
	if strings.HasPrefix(snakeMethod, "update_"+singularTable) {
		return "update"
	}
	if strings.HasPrefix(snakeMethod, "delete_"+singularTable) {
		return "delete"
	}

	// Generic removal of table name prefix
	if strings.HasPrefix(snakeMethod, snakeTable+"_") {
		return snakeMethod[len(snakeTable)+1:]
	}
	if strings.HasPrefix(snakeMethod, singularTable+"_") {
		return snakeMethod[len(singularTable)+1:]
	}

	return methodName
}

// buildCrystalQuery converts a plugin query to a crystalQuery
func (g *Generator) buildCrystalQuery(query *plugin.Query) crystalQuery {
	cq := crystalQuery{
		Name:         toSnakeCase(query.Name),
		ConstantName: toConstantCase(query.Name),
		Cmd:          query.Cmd,
		SQL:          query.Text,
		Comments:     query.Comments,
	}

	// Build parameters
	hasSlice := false
	for i, param := range query.Params {
		p := crystalParam{
			Name:     fmt.Sprintf("arg%d", i+1),
			Type:     g.crystalType(param.Column),
			Position: i + 1,
		}

		// Use column name if available
		if param.Column != nil && param.Column.Name != "" {
			p.Name = toSnakeCase(param.Column.Name)
		}
		
		// Check if this parameter is used with sqlc.slice()
		if param.Column != nil && param.Column.IsSqlcSlice {
			// Force the type to be an array
			baseType := g.baseType(param.Column)
			p.Type = fmt.Sprintf("Array(%s)", baseType)
			hasSlice = true
			
			// Add to slice params for template processing
			cq.SliceParams = append(cq.SliceParams, sqlcSliceParam{
				Name:        p.Name,
				Placeholder: fmt.Sprintf("$%d", param.Number),
			})
		}

		cq.Params = append(cq.Params, p)
	}
	
	// Mark query as using sqlc.slice() if any param has it
	if hasSlice {
		cq.UsesSQLCSlice = true
	}

	// Sort parameters to put required parameters before optional ones
	// Required parameters are those without '?' suffix (non-nullable)
	// Optional parameters are those with '?' suffix (nullable)
	sort.SliceStable(cq.Params, func(i, j int) bool {
		iIsOptional := strings.HasSuffix(cq.Params[i].Type, "?")
		jIsOptional := strings.HasSuffix(cq.Params[j].Type, "?")

		// If one is optional and the other isn't, required comes first
		if iIsOptional != jIsOptional {
			return !iIsOptional // required (non-optional) comes first
		}

		// If both are the same type (both required or both optional),
		// maintain original order by position
		return cq.Params[i].Position < cq.Params[j].Position
	})

	// Determine return type using deduplicated struct names (same logic as generateQueries)
	switch query.Cmd {
	case ":one":
		if len(query.Columns) == 1 {
			cq.ReturnType = g.crystalType(query.Columns[0])
		} else {
			cq.ReturnType = g.getStructNameForQuery(query)
		}
		cq.ReturnType += "?"
	case ":many":
		if len(query.Columns) == 1 {
			cq.ReturnType = "Array(" + g.crystalType(query.Columns[0]) + ")"
			cq.SingleColumnType = g.crystalType(query.Columns[0])
		} else {
			cq.ReturnType = "Array(" + g.getStructNameForQuery(query) + ")"
		}
	case ":exec":
		cq.ReturnType = "Nil"
	case ":execresult":
		cq.ReturnType = "DB::ExecResult"
	case ":execrows":
		cq.ReturnType = "Int64"
	case ":execlastid":
		cq.ReturnType = "Int64"
	case ":copyfrom":
		cq.ReturnType = "Int64"  // Will return number of rows copied
	}

	// Set the result struct name if needed using deduplicated names
	if len(query.Columns) > 1 && (query.Cmd == ":one" || query.Cmd == ":many") {
		// Use deduplicated struct name
		cq.ResultStruct = g.getStructNameForQuery(query)
	} else if len(query.Columns) == 1 && (query.Cmd == ":one" || query.Cmd == ":many") {
		// For single column queries, store the actual type (without ? or Array)
		if query.Cmd == ":one" {
			cq.SingleColumnType = g.crystalType(query.Columns[0])
		}
	}

	return cq
}
