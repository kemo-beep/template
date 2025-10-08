package generators

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// AutoRegister handles automatic registration of generated APIs
type AutoRegister struct {
	mainGoPath string
}

// NewAutoRegister creates a new AutoRegister instance
func NewAutoRegister(mainGoPath string) *AutoRegister {
	return &AutoRegister{mainGoPath: mainGoPath}
}

// RegisterGeneratedRoutes automatically registers generated routes in main.go
func (ar *AutoRegister) RegisterGeneratedRoutes(modelName string) error {
	// Read the current main.go file
	content, err := os.ReadFile(ar.mainGoPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	// Check if the route is already registered
	if strings.Contains(string(content), fmt.Sprintf("Setup%sRoutes", modelName)) {
		fmt.Printf("Routes for %s are already registered\n", modelName)
		return nil
	}

	// Generate the route registration code
	routeCode := ar.generateRouteRegistrationCode(modelName)

	// Find the insertion point (after generator routes)
	insertionPoint := "routes.SetupGeneratorRoutes(r, generatorController)"
	if !strings.Contains(string(content), insertionPoint) {
		return fmt.Errorf("insertion point not found in main.go")
	}

	// Insert the new route registration
	newContent := strings.Replace(string(content),
		insertionPoint,
		fmt.Sprintf("%s\n\n\t// Setup generated %s routes\n\t%s", insertionPoint, strings.ToLower(modelName), routeCode),
		1)

	// Write the updated main.go
	if err := os.WriteFile(ar.mainGoPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	fmt.Printf("✅ Routes for %s registered in main.go\n", modelName)
	return nil
}

// generateRouteRegistrationCode generates the code to register routes
func (ar *AutoRegister) generateRouteRegistrationCode(modelName string) string {
	tmpl := `{{.ModelNameLower}}Controller := controllers.New{{.ModelName}}Controller(config.GetDB())
	routes.Setup{{.ModelName}}Routes(r, {{.ModelNameLower}}Controller)`

	t, _ := template.New("route").Parse(tmpl)

	var buf strings.Builder
	t.Execute(&buf, map[string]string{
		"ModelName":      modelName,
		"ModelNameLower": strings.ToLower(modelName),
	})

	return buf.String()
}

// RegenerateSwagger regenerates the Swagger documentation
func (ar *AutoRegister) RegenerateSwagger() error {
	// Change to the backend directory
	backendDir := filepath.Dir(ar.mainGoPath)

	// Run swag init
	cmd := exec.Command("swag", "init", "-g", "main.go")
	cmd.Dir = backendDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to regenerate Swagger: %w\nOutput: %s", err, string(output))
	}

	fmt.Println("✅ Swagger documentation regenerated")
	return nil
}

// UnregisterGeneratedRoutes removes generated routes from main.go
func (ar *AutoRegister) UnregisterGeneratedRoutes(modelName string) error {
	// Read the current main.go file
	content, err := os.ReadFile(ar.mainGoPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	// Remove the route registration
	lines := strings.Split(string(content), "\n")
	var newLines []string
	inRouteBlock := false

	for _, line := range lines {
		// Check if we're entering the route block
		if strings.Contains(line, fmt.Sprintf("Setup%sRoutes", modelName)) {
			inRouteBlock = true
			continue
		}

		// Check if we're exiting the route block (next non-comment, non-empty line)
		if inRouteBlock && strings.TrimSpace(line) != "" && !strings.HasPrefix(strings.TrimSpace(line), "//") {
			inRouteBlock = false
		}

		// Skip lines in the route block
		if inRouteBlock {
			continue
		}

		newLines = append(newLines, line)
	}

	// Write the updated main.go
	if err := os.WriteFile(ar.mainGoPath, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	fmt.Printf("✅ Routes for %s unregistered from main.go\n", modelName)
	return nil
}
