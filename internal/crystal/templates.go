package crystal

import (
	"fmt"
	"sort"
	"strings"
	"text/template"
)

var modelsTemplate = template.Must(template.New("models").Funcs(template.FuncMap{
	"crystalModule": crystalModuleName,
	"isBooleanType": isBooleanType,
}).Parse(modelsTemplateStr))
var queriesTemplate = template.Must(template.New("queries").Funcs(template.FuncMap{
	"paramNames":          paramNames,
	"paramList":           paramList,
	"crystalModule":       crystalModuleName,
	"join":                strings.Join,
	"printf":              fmt.Sprintf,
	"len":                 func(v any) int { return len(v.([]crystalQuery)) },
	"trimSuffix":          strings.TrimSuffix,
	"trimPrefix":          strings.TrimPrefix,
	"joinComments":        joinComments,
	"expandSliceParams":   expandSliceParams,
	"needsSliceExpansion": needsSliceExpansion,
	"contains":            strings.Contains,
}).Parse(queriesTemplateStr))

const modelsTemplateStr = `module {{ .Package | crystalModule }}
{{- range .Structs }}
  struct {{ .Name }}
    include DB::Serializable
    {{- if $.EmitJSONTags }}
    include JSON::Serializable
    {{- end }}

    {{- range .Fields }}
    {{- if and $.EmitJSONTags .JSONName }}
    @[JSON::Field(key: {{ .JSONName | printf "%q" }})]
    {{- end }}
    {{- if and $.EmitDBTags (ne .DBName .Name) }}
    @[DB::Field(key: {{ .DBName | printf "%q" }})]
    {{- end }}
    {{- if and $.EmitBooleanQuestionGetters (isBooleanType .Type) }}
    getter? {{ .Name }} : {{ .Type }}
    {{- else }}
    getter {{ .Name }} : {{ .Type }}
    {{- end }}
    {{- end }}
  end
{{ end -}}
end
`

const queriesTemplateStr = `require "db"

module {{ .Package | crystalModule }}
  class Queries
    SQL_{{ .Queries | len | printf "%d_QUERIES" }} = {
      {{- range .Queries }}
      {{ .ConstantName }}: {{ .SQL | printf "%q" }},
      {{- end }}
    }

    def initialize(@db : DB::Database)
    end

    {{- range .Queries }}
    {{- if .Comments }}

    # {{ .Comments | joinComments }}
    {{- end }}
    def {{ .Name }}({{ .Params | paramList }}) : {{ .ReturnType }}
      {{- if needsSliceExpansion . $.Engine }}
      sql = SQL_{{ len $.Queries | printf "%d_QUERIES" }}[{{ .ConstantName | printf ":%s" }}]
      {{ expandSliceParams .Params .SliceParams }}

      # Flatten array parameters for execution
      query_params = [] of DB::Any
      {{- range .Params }}
      {{- if contains .Type "Array(" }}
      query_params.concat({{ .Name }}.map { |v| v.as(DB::Any) })
      {{- else }}
      query_params << {{ .Name }}.as(DB::Any)
      {{- end }}
      {{- end }}

      {{- if eq .Cmd ":one" }}
      {{- if .ResultStruct }}
      @db.query_one?(sql, args: query_params, as: {{ .ResultStruct }})
      {{- else }}
      @db.query_one?(sql, args: query_params) do |rs|
        rs.read({{ .SingleColumnType }})
      end
      {{- end }}
      {{- else if eq .Cmd ":many" }}
      {{- if .ResultStruct }}
      @db.query_all(sql, args: query_params, as: {{ .ResultStruct }})
      {{- else }}
      results = [] of {{ .SingleColumnType }}
      @db.query(sql, args: query_params) do |rs|
        rs.each do
          results << rs.read({{ .SingleColumnType }})
        end
      end
      results
      {{- end }}
      {{- else if eq .Cmd ":exec" }}
      @db.exec(sql, args: query_params)
      nil
      {{- else if eq .Cmd ":execrows" }}
      result = @db.exec(sql, args: query_params)
      result.rows_affected
      {{- end }}
      {{- else if eq .Cmd ":one" }}
      {{- if .ResultStruct }}
      @db.query_one?(
        SQL_{{ len $.Queries | printf "%d_QUERIES" }}[{{ .ConstantName | printf ":%s" }}],{{ if .Params }}
        {{ .Params | paramNames }},{{ end }}
        as: {{ .ResultStruct }}
      )
      {{- else }}
      result = @db.query_one?(
        SQL_{{ len $.Queries | printf "%d_QUERIES" }}[{{ .ConstantName | printf ":%s" }}]{{ if .Params }},
        {{ .Params | paramNames }}{{ end }}
      ) do |rs|
        rs.read({{ .SingleColumnType }})
      end
      result
      {{- end }}
      {{- else if eq .Cmd ":many" }}
      {{- if .ResultStruct }}
      @db.query_all(
        SQL_{{ len $.Queries | printf "%d_QUERIES" }}[{{ .ConstantName | printf ":%s" }}],{{ if .Params }}
        {{ .Params | paramNames }},{{ end }}
        as: {{ .ResultStruct }}
      )
      {{- else }}
      results = [] of {{ .SingleColumnType }}
      @db.query(
        SQL_{{ len $.Queries | printf "%d_QUERIES" }}[{{ .ConstantName | printf ":%s" }}]{{ if .Params }},
        {{ .Params | paramNames }}{{ end }}
      ) do |rs|
        rs.each do
          results << rs.read({{ .SingleColumnType }})
        end
      end
      results
      {{- end }}
      {{- else if eq .Cmd ":exec" }}
      @db.exec(
        SQL_{{ len $.Queries | printf "%d_QUERIES" }}[{{ .ConstantName | printf ":%s" }}]{{ if .Params }},
        {{ .Params | paramNames }}{{ end }}
      )
      nil
      {{- else if eq .Cmd ":execresult" }}
      @db.exec(
        SQL_{{ len $.Queries | printf "%d_QUERIES" }}[{{ .ConstantName | printf ":%s" }}]{{ if .Params }},
        {{ .Params | paramNames }}{{ end }}
      )
      {{- else if eq .Cmd ":execrows" }}
      result = @db.exec(
        SQL_{{ len $.Queries | printf "%d_QUERIES" }}[{{ .ConstantName | printf ":%s" }}]{{ if .Params }},
        {{ .Params | paramNames }}{{ end }}
      )
      result.rows_affected
      {{- else if eq .Cmd ":execlastid" }}
      result = @db.exec(
        SQL_{{ len $.Queries | printf "%d_QUERIES" }}[{{ .ConstantName | printf ":%s" }}]{{ if .Params }},
        {{ .Params | paramNames }}{{ end }}
      )
      result.last_insert_id
      {{- else if eq .Cmd ":copyfrom" }}
      # TODO: Implement bulk insert functionality
      # Crystal doesn't have direct support for PostgreSQL COPY or MySQL LOAD DATA
      # For now, this returns 0 as a placeholder
      # Consider implementing batch inserts using transactions and prepared statements
      0_i64
      {{- end }}
    end
    {{- end }}
  end
end
`

// Helper functions for templates

func paramNames(params []crystalParam) string {
	if len(params) == 0 {
		return ""
	}

	// Create a copy of params and sort by original position for SQL parameter order
	paramsCopy := make([]crystalParam, len(params))
	copy(paramsCopy, params)
	sort.SliceStable(paramsCopy, func(i, j int) bool {
		return paramsCopy[i].Position < paramsCopy[j].Position
	})

	names := make([]string, len(paramsCopy))
	for i, p := range paramsCopy {
		names[i] = p.Name
	}
	return strings.Join(names, ", ")
}

func paramList(params []crystalParam) string {
	if len(params) == 0 {
		return ""
	}

	parts := make([]string, len(params))
	for i, p := range params {
		// Add default nil value for nullable parameters
		if strings.HasSuffix(p.Type, "?") {
			parts[i] = p.Name + " : " + p.Type + " = nil"
		} else {
			parts[i] = p.Name + " : " + p.Type
		}
	}
	return strings.Join(parts, ", ")
}

func joinComments(comments []string) string {
	if len(comments) == 0 {
		return ""
	}
	return strings.Join(comments, "\n    # ")
}

func needsSliceExpansion(query crystalQuery, engine string) bool {
	return query.UsesSQLCSlice && (engine == "mysql" || engine == "sqlite")
}

func expandSliceParams(params []crystalParam, sliceParams []sqlcSliceParam) string {
	if len(sliceParams) == 0 {
		return ""
	}

	var expansions []string
	for _, sp := range sliceParams {
		// Find the corresponding parameter
		for _, p := range params {
			if p.Name == sp.Name {
				// Replace the /*SLICE:name*/? marker with proper question marks
				expansions = append(expansions, fmt.Sprintf(`
      # Expand array parameter %s for MySQL/SQLite
      if %s.empty?
        raise ArgumentError.new("slice parameter '%s' cannot be empty")
      end
      placeholders_%s = %s.size.times.map { "?" }.join(", ")
      sql = sql.gsub("/*SLICE:%s*/?", placeholders_%s)`,
					p.Name, p.Name, p.Name, p.Name, p.Name,
					p.Name, p.Name))
				break
			}
		}
	}

	return strings.Join(expansions, "\n")
}

// isBooleanType checks if a Crystal type is Bool or Bool?
func isBooleanType(crystalType string) bool {
	return crystalType == "Bool" || crystalType == "Bool?"
}

// Connection Manager template
const databaseTemplate = `# Main entry point for generated database code
# Always require models and queries
require "./models"
require "./queries"

{{- if .GenerateRepositories }}
# Require all repository files
require "./repositories/*"
{{- end }}

{{- if .GenerateConnectionManager }}
module {{ .Package | crystalModule }}
  class Database
    @@instance : DB::Database?
    @@queries : Queries?

    def self.connection
      @@instance ||= DB.open(ENV["DATABASE_URL"])
    end

    def self.queries
      @@queries ||= Queries.new(connection)
    end

    def self.transaction(&block)
      connection.transaction do |tx|
        yield Queries.new(tx.connection)
      end
    end

    def self.close
      @@instance.try(&.close)
      @@instance = nil
      @@queries = nil
    end
  end

  # Transaction context for repositories
  class TransactionContext
    def initialize(@queries : Queries)
    end

    def queries : Queries
      @queries
    end
  end
end
{{- end }}
`

// Repository template for a single table
const repositoryTemplate = `module {{ .Package | crystalModule }}
  class {{ .TableName }}Repository
    def initialize(@queries : Queries? = nil)
    end

    private def queries
      @queries || Database.queries
    end
    {{- range .Methods }}
    {{- if .IsTableSpecific }}

    def {{ .MethodName }}({{ .Params | paramList }}){{ if .ReturnType }} : {{ .ReturnType }}{{ end }}
      queries.{{ .Name }}({{ .Params | paramNames }})
    end
    {{- end }}
    {{- end }}

    # Create a repository instance that uses the given transaction context
    def self.with_transaction(tx_queries : Queries)
      new(tx_queries)
    end

    # Execute a block within a transaction, automatically creating repository instances
    def self.transaction(&block : {{ .TableName }}Repository ->)
      Database.transaction do |tx_queries|
        repo = new(tx_queries)
        yield repo
      end
    end
  end
end
`
