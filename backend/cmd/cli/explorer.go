package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// explorerCmd represents the explorer command
var explorerCmd = &cobra.Command{
	Use:   "explorer",
	Short: "Start interactive API explorer",
	Long: `Start an interactive web-based API explorer for testing and exploring your API:

Examples:
  mobile-backend-cli explorer
  mobile-backend-cli explorer --port 3000
  mobile-backend-cli explorer --open`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		port, _ := cmd.Flags().GetString("port")
		host, _ := cmd.Flags().GetString("host")
		theme, _ := cmd.Flags().GetString("theme")
		cors, _ := cmd.Flags().GetBool("cors")
		open, _ := cmd.Flags().GetBool("open")

		// Get configuration
		baseURL := viper.GetString("base_url")
		apiKey := viper.GetString("api_key")

		config := ExplorerConfig{
			BaseURL:    baseURL,
			APIKey:     apiKey,
			Endpoints:  loadAPIEndpoints(),
			Theme:      theme,
			EnableCORS: cors,
		}

		// Start the explorer server
		startExplorerServer(host, port, config)

		// Open browser if requested
		if open {
			openBrowser("http://" + host + ":" + port)
		}
	},
}

func init() {

	explorerCmd.Flags().StringP("port", "p", "3000", "Port for the API explorer")
	explorerCmd.Flags().StringP("host", "H", "localhost", "Host for the API explorer")
	explorerCmd.Flags().BoolP("open", "o", false, "Open explorer in browser")
	explorerCmd.Flags().StringP("theme", "t", "light", "Theme (light, dark)")
	explorerCmd.Flags().BoolP("cors", "c", true, "Enable CORS for API calls")
}

// API Explorer data structures
type APIEndpoint struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Description string            `json:"description"`
	Parameters  []APIParameter    `json:"parameters"`
	Headers     map[string]string `json:"headers"`
	Body        interface{}       `json:"body"`
	Responses   []APIResponse     `json:"responses"`
}

type APIParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

type APIResponse struct {
	StatusCode  int         `json:"status_code"`
	Description string      `json:"description"`
	Body        interface{} `json:"body"`
}

type ExplorerConfig struct {
	BaseURL    string        `json:"base_url"`
	APIKey     string        `json:"api_key"`
	Endpoints  []APIEndpoint `json:"endpoints"`
	Theme      string        `json:"theme"`
	EnableCORS bool          `json:"enable_cors"`
}

func openBrowser(url string) {
	var err error
	switch {
	case strings.Contains(strings.ToLower(os.Getenv("OS")), "windows"):
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case strings.Contains(strings.ToLower(os.Getenv("SHELL")), "zsh") || strings.Contains(strings.ToLower(os.Getenv("SHELL")), "bash"):
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		fmt.Printf("⚠️  Could not open browser: %v\n", err)
		fmt.Printf("🌐 Please open your browser and navigate to: %s\n", url)
	}
}

func startAPIExplorer() {
	port, _ := explorerCmd.Flags().GetString("port")
	host, _ := explorerCmd.Flags().GetString("host")
	theme, _ := explorerCmd.Flags().GetString("theme")
	enableCORS, _ := explorerCmd.Flags().GetBool("cors")
	open, _ := explorerCmd.Flags().GetBool("open")

	baseURL := viper.GetString("base_url")
	apiKey := viper.GetString("api_key")

	fmt.Printf("🔍 Starting API Explorer...\n")
	fmt.Printf("🌐 Host: %s\n", host)
	fmt.Printf("🔌 Port: %s\n", port)
	fmt.Printf("🎨 Theme: %s\n", theme)
	fmt.Printf("🌍 Base URL: %s\n", baseURL)
	fmt.Printf("🔑 API Key: %s\n", maskAPIKey(apiKey))
	fmt.Printf("🔀 CORS: %t\n\n", enableCORS)

	// Load API endpoints
	endpoints := loadAPIEndpoints()

	// Create explorer configuration
	config := ExplorerConfig{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		Endpoints:  endpoints,
		Theme:      theme,
		EnableCORS: enableCORS,
	}

	// Start web server
	startExplorerServer(host, port, config)

	if open {
		fmt.Printf("🌐 Opening API Explorer in browser...\n")
		// This would open the browser
		fmt.Printf("   URL: http://%s:%s\n", host, port)
	}
}

func loadAPIEndpoints() []APIEndpoint {
	// Load endpoints from Swagger documentation or define them
	endpoints := []APIEndpoint{
		{
			Method:      "GET",
			Path:        "/health",
			Description: "Health check endpoint",
			Parameters:  []APIParameter{},
			Headers:     map[string]string{},
			Responses: []APIResponse{
				{
					StatusCode:  200,
					Description: "Service is healthy",
					Body: map[string]interface{}{
						"status":    "healthy",
						"timestamp": time.Now().Format(time.RFC3339),
					},
				},
			},
		},
		{
			Method:      "GET",
			Path:        "/users",
			Description: "List all users",
			Parameters: []APIParameter{
				{
					Name:        "page",
					Type:        "integer",
					Required:    false,
					Description: "Page number",
					Example:     "1",
				},
				{
					Name:        "limit",
					Type:        "integer",
					Required:    false,
					Description: "Items per page",
					Example:     "10",
				},
			},
			Headers: map[string]string{
				"Authorization": "Bearer {{api_key}}",
			},
			Responses: []APIResponse{
				{
					StatusCode:  200,
					Description: "List of users",
					Body: map[string]interface{}{
						"data": []map[string]interface{}{
							{
								"id":    1,
								"name":  "John Doe",
								"email": "john@example.com",
							},
						},
						"total": 1,
						"page":  1,
						"limit": 10,
					},
				},
			},
		},
		{
			Method:      "POST",
			Path:        "/users",
			Description: "Create a new user",
			Parameters:  []APIParameter{},
			Headers: map[string]string{
				"Authorization": "Bearer {{api_key}}",
				"Content-Type":  "application/json",
			},
			Body: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			Responses: []APIResponse{
				{
					StatusCode:  201,
					Description: "User created successfully",
					Body: map[string]interface{}{
						"id":    1,
						"name":  "John Doe",
						"email": "john@example.com",
					},
				},
				{
					StatusCode:  400,
					Description: "Validation error",
					Body: map[string]interface{}{
						"error": "Invalid request body",
					},
				},
			},
		},
		{
			Method:      "GET",
			Path:        "/users/{id}",
			Description: "Get user by ID",
			Parameters: []APIParameter{
				{
					Name:        "id",
					Type:        "integer",
					Required:    true,
					Description: "User ID",
					Example:     "1",
				},
			},
			Headers: map[string]string{
				"Authorization": "Bearer {{api_key}}",
			},
			Responses: []APIResponse{
				{
					StatusCode:  200,
					Description: "User details",
					Body: map[string]interface{}{
						"id":    1,
						"name":  "John Doe",
						"email": "john@example.com",
					},
				},
				{
					StatusCode:  404,
					Description: "User not found",
					Body: map[string]interface{}{
						"error": "User not found",
					},
				},
			},
		},
		{
			Method:      "PUT",
			Path:        "/users/{id}",
			Description: "Update user by ID",
			Parameters: []APIParameter{
				{
					Name:        "id",
					Type:        "integer",
					Required:    true,
					Description: "User ID",
					Example:     "1",
				},
			},
			Headers: map[string]string{
				"Authorization": "Bearer {{api_key}}",
				"Content-Type":  "application/json",
			},
			Body: map[string]interface{}{
				"name":  "John Doe Updated",
				"email": "john.updated@example.com",
			},
			Responses: []APIResponse{
				{
					StatusCode:  200,
					Description: "User updated successfully",
					Body: map[string]interface{}{
						"id":    1,
						"name":  "John Doe Updated",
						"email": "john.updated@example.com",
					},
				},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/users/{id}",
			Description: "Delete user by ID",
			Parameters: []APIParameter{
				{
					Name:        "id",
					Type:        "integer",
					Required:    true,
					Description: "User ID",
					Example:     "1",
				},
			},
			Headers: map[string]string{
				"Authorization": "Bearer {{api_key}}",
			},
			Responses: []APIResponse{
				{
					StatusCode:  200,
					Description: "User deleted successfully",
					Body: map[string]interface{}{
						"message": "User deleted successfully",
					},
				},
			},
		},
	}

	return endpoints
}

func startExplorerServer(host, port string, config ExplorerConfig) {
	// Create HTTP server
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("cmd/cli/static/"))))

	// API endpoints
	mux.HandleFunc("/api/endpoints", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config.Endpoints)
	})

	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	})

	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		handleAPITest(w, r, config)
	})

	// Main explorer page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveExplorerPage(w, r, config)
	})

	// Start server
	addr := host + ":" + port
	fmt.Printf("🚀 API Explorer server starting on http://%s\n", addr)
	fmt.Printf("🛑 Press Ctrl+C to stop\n\n")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("❌ Failed to start server: %v\n", err)
	}
}

func handleAPITest(w http.ResponseWriter, r *http.Request, config ExplorerConfig) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var testRequest struct {
		Method  string            `json:"method"`
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&testRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Make API call
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(testRequest.Method, testRequest.URL, strings.NewReader(testRequest.Body))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Add headers
	for key, value := range testRequest.Headers {
		req.Header.Set(key, value)
	}

	// Add API key if not present
	if config.APIKey != "" && req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+config.APIKey)
	}

	// Make request
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		http.Error(w, "Request failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        string(body),
		"duration":    duration.Milliseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func serveExplorerPage(w http.ResponseWriter, r *http.Request, config ExplorerConfig) {
	// Check if static files exist, if not create them
	if err := createStaticFiles(); err != nil {
		http.Error(w, "Failed to create static files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Serve the main HTML page
	tmpl := template.Must(template.New("explorer").Parse(explorerHTML))
	tmpl.Execute(w, config)
}

func createStaticFiles() error {
	// Create static directory
	staticDir := "cmd/cli/static"
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return err
	}

	// Create CSS file
	cssFile := filepath.Join(staticDir, "style.css")
	if _, err := os.Stat(cssFile); os.IsNotExist(err) {
		if err := ioutil.WriteFile(cssFile, []byte(explorerCSS), 0644); err != nil {
			return err
		}
	}

	// Create JavaScript file
	jsFile := filepath.Join(staticDir, "script.js")
	if _, err := os.Stat(jsFile); os.IsNotExist(err) {
		if err := ioutil.WriteFile(jsFile, []byte(explorerJS), 0644); err != nil {
			return err
		}
	}

	return nil
}

func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "Not set"
	}
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "***" + apiKey[len(apiKey)-4:]
}

// HTML template for the API explorer
const explorerHTML = `<!DOCTYPE html>
<html>
<head>
    <title>API Explorer</title>
</head>
<body>
    <h1>API Explorer</h1>
    <p>Interactive API testing tool</p>
</body>
</html>`

// CSS for the API explorer
const explorerCSS = `body { font-family: Arial, sans-serif; }`

// JavaScript for the API explorer
const explorerJS = `console.log('API Explorer loaded');`
