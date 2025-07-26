package codegen

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sqlc-dev/plugin-sdk-go/codegen"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/watzon/sqlc-gen-crystal/internal/crystal"
)

// Options represents the plugin-specific options
type Options struct {
	Module                    string `json:"module"`
	EmitJSONTags              bool   `json:"emit_json_tags"`
	EmitDBTags                bool   `json:"emit_db_tags"`
	EmitResultStructPointers  bool   `json:"emit_result_struct_pointers"`
	GenerateConnectionManager bool   `json:"generate_connection_manager"`
	GenerateRepositories      bool   `json:"generate_repositories"`
	EmitBooleanQuestionGetters bool   `json:"emit_boolean_question_getters"`
}

// Run is the main entry point for the plugin
func Run() {
	codegen.Run(generate)
}

func generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	// Parse plugin options
	var options Options
	if len(req.PluginOptions) > 0 {
		if err := json.Unmarshal(req.PluginOptions, &options); err != nil {
			return nil, fmt.Errorf("failed to parse plugin options: %w", err)
		}
	}
	
	// Set default module name if not provided
	moduleName := options.Module
	if moduleName == "" {
		moduleName = "Db"
	}
	
	if !options.EmitJSONTags && !options.EmitDBTags {
		options.EmitDBTags = true // Default to emitting DB tags
	}
	
	// Create the Crystal code generator
	gen := crystal.NewGenerator(req, moduleName, crystal.GeneratorOptions{
		EmitJSONTags:              options.EmitJSONTags,
		EmitDBTags:                options.EmitDBTags,
		EmitResultStructPointers:  options.EmitResultStructPointers,
		GenerateConnectionManager: options.GenerateConnectionManager,
		GenerateRepositories:      options.GenerateRepositories,
		EmitBooleanQuestionGetters: options.EmitBooleanQuestionGetters,
	})
	
	// Generate the code
	resp, err := gen.Generate(ctx)
	if err != nil {
		return nil, fmt.Errorf("code generation failed: %w", err)
	}
	
	return resp, nil
}