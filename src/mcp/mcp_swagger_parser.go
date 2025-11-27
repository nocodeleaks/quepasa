package mcp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// SwaggerEndpoint represents an API endpoint parsed from Swagger comments
type SwaggerEndpoint struct {
	FuncName    string         // Controller function name (e.g., ContactsController)
	Name        string         // Tool name from @Summary or @MCPTool
	Description string         // Tool description from @Description
	Method      string         // HTTP method (GET, POST, PUT, DELETE)
	Path        string         // API path (e.g., /contacts)
	Params      []SwaggerParam // Parameters from @Param
	Security    bool           // Whether @Security is present
	MCPHidden   bool           //	@MCPHidden	- exclude from MCP
	MCPCategory string         //	@MCPCategory	- tool category
	MCPAuth     string         //	@MCPAuth	- authentication mode
}

// SwaggerParam represents a parameter from @Param annotation
type SwaggerParam struct {
	Name        string // Parameter name
	In          string // Location: path, query, body, header
	Type        string // Data type: string, int, bool, object
	Required    bool   // Whether parameter is required
	Description string // Parameter description
}

// ParseSwaggerAnnotations scans API controller files for Swagger comments
func ParseSwaggerAnnotations(apiDir string) ([]SwaggerEndpoint, error) {
	log.Infof("MCP: Parsing Swagger annotations from %s", apiDir)

	fset := token.NewFileSet()
	pattern := filepath.Join(apiDir, "*.go")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob API files: %w", err)
	}

	log.Infof("MCP: Found %d Go files to parse", len(matches))

	var endpoints []SwaggerEndpoint
	for _, file := range matches {
		log.Debugf("MCP: Parsing file %s", file)

		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			log.Warnf("MCP: Failed to parse file %s: %v", file, err)
			continue
		}

		// Extract endpoints from this file
		fileEndpoints := extractEndpointsFromFile(f)
		log.Debugf("MCP: File %s yielded %d endpoints", filepath.Base(file), len(fileEndpoints))
		endpoints = append(endpoints, fileEndpoints...)
	}

	log.Infof("MCP: Parsing complete. Total endpoints found: %d", len(endpoints))
	return endpoints, nil
}

// extractEndpointsFromFile extracts endpoints from a parsed Go file
func extractEndpointsFromFile(f *ast.File) []SwaggerEndpoint {
	var endpoints []SwaggerEndpoint
	controllersFound := 0

	// Iterate through all declarations
	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Doc == nil {
			continue
		}

		// Check if this is a controller function (has http.ResponseWriter, *http.Request params)
		if !isControllerFunc(funcDecl) {
			continue
		}

		controllersFound++

		// Parse Swagger comments
		endpoint := parseSwaggerComments(funcDecl.Name.Name, funcDecl.Doc.Text())
		if endpoint.Path != "" {
			endpoints = append(endpoints, endpoint)
		}
	}

	if controllersFound > 0 {
		log.Debugf("MCP: File had %d controllers, %d with valid routes", controllersFound, len(endpoints))
	}

	return endpoints
}

// isControllerFunc checks if function signature matches a controller
func isControllerFunc(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) != 2 {
		return false
	}

	// Check for (w http.ResponseWriter, r *http.Request) signature
	params := funcDecl.Type.Params.List

	// First param should be http.ResponseWriter
	firstType := getTypeName(params[0].Type)
	if firstType != "ResponseWriter" && firstType != "http.ResponseWriter" {
		return false
	}

	// Second param should be *http.Request
	secondType := getTypeName(params[1].Type)
	if secondType != "*Request" && secondType != "*http.Request" {
		return false
	}

	return true
}

// getTypeName extracts type name from ast.Expr
func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeName(t.X)
	case *ast.SelectorExpr:
		return getTypeName(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

// parseSwaggerComments parses Swagger annotations from function comments
func parseSwaggerComments(funcName, docText string) SwaggerEndpoint {
	endpoint := SwaggerEndpoint{
		FuncName: funcName,
		Params:   []SwaggerParam{},
	}

	lines := strings.Split(docText, "\n")
	var descriptionLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse @Summary
		if strings.HasPrefix(line, "@Summary") {
			endpoint.Name = strings.TrimSpace(strings.TrimPrefix(line, "@Summary"))
			continue
		}

		// Parse @Description
		if strings.HasPrefix(line, "@Description") {
			desc := strings.TrimSpace(strings.TrimPrefix(line, "@Description"))
			descriptionLines = append(descriptionLines, desc)
			continue
		}

		// Parse @Router /path [method]
		if strings.HasPrefix(line, "@Router") {
			parseRouter(line, &endpoint)
			continue
		}

		// Parse @Param
		if strings.HasPrefix(line, "@Param") {
			param := parseParam(line)
			if param.Name != "" {
				endpoint.Params = append(endpoint.Params, param)
			}
			continue
		}

		// Parse @Security
		if strings.HasPrefix(line, "@Security") {
			endpoint.Security = true
			continue
		}

		// Parse @MCPTool (overrides @Summary)
		if strings.HasPrefix(line, "@MCPTool") {
			endpoint.Name = strings.TrimSpace(strings.TrimPrefix(line, "@MCPTool"))
			continue
		}

		// Parse @MCPHidden
		if strings.HasPrefix(line, "@MCPHidden") {
			endpoint.MCPHidden = true
			continue
		}

		// Parse @MCPCategory
		if strings.HasPrefix(line, "@MCPCategory") {
			endpoint.MCPCategory = strings.TrimSpace(strings.TrimPrefix(line, "@MCPCategory"))
			continue
		}

		// Parse @MCPAuth
		if strings.HasPrefix(line, "@MCPAuth") {
			endpoint.MCPAuth = strings.TrimSpace(strings.TrimPrefix(line, "@MCPAuth"))
			continue
		}
	}

	// Join description lines
	endpoint.Description = strings.Join(descriptionLines, " ")

	return endpoint
}

// parseRouter extracts path and method from @Router annotation
// Format: @Router /path [method]
func parseRouter(line string, endpoint *SwaggerEndpoint) {
	// Remove @Router prefix
	rest := strings.TrimSpace(strings.TrimPrefix(line, "@Router"))

	// Extract path and method using regex
	re := regexp.MustCompile(`^(/[^\s\[]*)\s*\[([^\]]+)\]`)
	matches := re.FindStringSubmatch(rest)

	if len(matches) == 3 {
		endpoint.Path = matches[1]
		endpoint.Method = strings.ToUpper(matches[2])
	}
}

// parseParam extracts parameter info from @Param annotation
// Format: @Param name location type required "description"
func parseParam(line string) SwaggerParam {
	param := SwaggerParam{}

	// Remove @Param prefix
	rest := strings.TrimSpace(strings.TrimPrefix(line, "@Param"))

	// Split by whitespace, but preserve quoted strings
	parts := splitRespectingQuotes(rest)

	if len(parts) >= 4 {
		param.Name = parts[0]
		param.In = parts[1]
		param.Type = parts[2]
		param.Required = parts[3] == "true"

		// Description is the rest, may be quoted
		if len(parts) > 4 {
			param.Description = strings.Join(parts[4:], " ")
			param.Description = strings.Trim(param.Description, "\"")
		}
	}

	return param
}

// splitRespectingQuotes splits string by whitespace but preserves quoted parts
func splitRespectingQuotes(s string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false

	for _, r := range s {
		switch r {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		case ' ', '\t':
			if inQuotes {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// GenerateToolName converts endpoint info to a snake_case tool name
func GenerateToolName(endpoint SwaggerEndpoint) string {
	// If @MCPTool specified, use it
	if endpoint.Name != "" && strings.Contains(endpoint.Name, "_") {
		return endpoint.Name
	}

	// Generate from Summary
	name := endpoint.Name
	if name == "" {
		name = endpoint.FuncName
	}

	// Convert to snake_case
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	// Remove special characters
	re := regexp.MustCompile(`[^a-z0-9_]`)
	name = re.ReplaceAllString(name, "")

	return name
}

// GenerateInputSchema creates JSON schema from Swagger params
func GenerateInputSchema(endpoint SwaggerEndpoint) map[string]interface{} {
	properties := make(map[string]interface{})
	required := []string{} // Initialize with empty array, NOT nil

	// Add token parameter for all endpoints (optional, needed for master key)
	properties["token"] = map[string]interface{}{
		"type":        "string",
		"description": "Bot token (required when using master key authentication)",
	}

	for _, param := range endpoint.Params {
		// Skip internal parameters
		if param.Name == "token" || param.Name == "Authorization" {
			continue
		}

		// Map Swagger types to JSON schema types
		jsonType := mapSwaggerTypeToJSON(param.Type)

		properties[param.Name] = map[string]interface{}{
			"type":        jsonType,
			"description": param.Description,
		}

		if param.Required {
			required = append(required, param.Name)
		}
	}

	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required, // Always include required, even if empty
	}
}

// mapSwaggerTypeToJSON maps Swagger data types to JSON schema types
func mapSwaggerTypeToJSON(swaggerType string) string {
	switch strings.ToLower(swaggerType) {
	case "integer", "int", "int32", "int64":
		return "number"
	case "number", "float", "double":
		return "number"
	case "boolean", "bool":
		return "boolean"
	case "array":
		return "array"
	case "object":
		return "object"
	default:
		return "string"
	}
}
