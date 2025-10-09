package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

// sdkCmd represents the sdk command
var sdkCmd = &cobra.Command{
	Use:   "sdk",
	Short: "Generate mobile SDKs for various platforms",
	Long: `Generate mobile SDKs (Software Development Kits) for different platforms like TypeScript, Swift, Kotlin, and Dart.

This command helps you create client libraries that can be easily integrated into your mobile applications.

Examples:
  mobile-backend-cli sdk --lang typescript --output ./sdks --package my-app-sdk --base-url http://localhost:8080
  mobile-backend-cli sdk --lang swift --output ./sdks --package MyAppSDK
  mobile-backend-cli sdk --lang kotlin --output ./sdks --package com.myapp.sdk`,
	Run: func(cmd *cobra.Command, args []string) {
		lang, _ := cmd.Flags().GetString("lang")
		outputDir, _ := cmd.Flags().GetString("output")
		packageName, _ := cmd.Flags().GetString("package")
		baseURL, _ := cmd.Flags().GetString("base-url")

		if packageName == "" {
			fmt.Println("Error: --package flag is required.")
			return
		}
		if baseURL == "" {
			fmt.Println("Error: --base-url flag is required.")
			return
		}

		switch strings.ToLower(lang) {
		case "typescript", "ts":
			generateTypeScriptSDK(outputDir, packageName, baseURL)
		case "swift":
			generateSwiftSDK(outputDir, packageName, baseURL)
		case "kotlin":
			generateKotlinSDK(outputDir, packageName, baseURL)
		case "dart":
			generateDartSDK(outputDir, packageName, baseURL)
		default:
			fmt.Printf("Error: Unsupported language '%s'. Supported languages are: typescript, swift, kotlin, dart\n", lang)
		}
	},
}

func init() {
	sdkCmd.Flags().StringP("lang", "l", "typescript", "Target language for the SDK (typescript, swift, kotlin, dart)")
	sdkCmd.Flags().StringP("output", "o", "./sdks", "Output directory for the generated SDK")
	sdkCmd.Flags().StringP("package", "p", "mobile-backend-sdk", "Package name for the generated SDK")
	sdkCmd.Flags().StringP("base-url", "u", "http://localhost:8080", "Base URL for the API endpoints")
}

// SDKConfig holds configuration for SDK generation
type SDKConfig struct {
	PackageName string           `json:"package_name"`
	BaseURL     string           `json:"base_url"`
	Version     string           `json:"version"`
	Language    string           `json:"language"`
	GeneratedAt string           `json:"generated_at"`
	Endpoints   []SDKAPIEndpoint `json:"endpoints"`
	Models      []Model          `json:"models"`
}

// SDK API Endpoint definition
type SDKAPIEndpoint struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Description string            `json:"description"`
	Parameters  []SDKAPIParameter `json:"parameters"`
	Response    interface{}       `json:"response"`
	Auth        bool              `json:"auth"`
}

type SDKAPIParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// Model definition
type Model struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Fields      []Field `json:"fields"`
}

// Field definition for a model
type Field struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	IsID        bool   `json:"is_id"`
	IsTimestamp bool   `json:"is_timestamp"`
}

// Relationship definition for a model
type Relationship struct {
	Type    string `json:"type"`
	Model   string `json:"model"`
	Field   string `json:"field"`
	Foreign string `json:"foreign"`
}

// Generate TypeScript SDK
func generateTypeScriptSDK(outputDir, packageName, baseURL string) {
	fmt.Printf("🚀 Generating TypeScript SDK...\n")

	config := SDKConfig{
		PackageName: packageName,
		BaseURL:     baseURL,
		Version:     "1.0.0",
		Language:    "typescript",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Endpoints:   getAPIEndpoints(),
		Models:      getModels(),
	}

	// Create output directory structure
	tsDir := filepath.Join(outputDir, "typescript")
	srcDir := filepath.Join(tsDir, "src")
	servicesDir := filepath.Join(srcDir, "services")
	typesDir := filepath.Join(srcDir, "types")

	os.MkdirAll(servicesDir, 0755)
	os.MkdirAll(typesDir, 0755)

	// Generate package.json
	generateTypeScriptPackageJSON(tsDir, config)

	// Generate TypeScript configuration
	generateTypeScriptConfig(tsDir, config)

	// Generate TypeScript files
	generateTypeScriptClient(srcDir, config)
	generateTypeScriptServices(servicesDir, config)
	generateTypeScriptTypes(typesDir, config)
	generateTypeScriptIndex(srcDir, config)
	generateTypeScriptReadme(tsDir, config)

	fmt.Printf("✅ TypeScript SDK generated in: %s\n", tsDir)
}

// Generate Swift SDK
func generateSwiftSDK(outputDir, packageName, baseURL string) {
	fmt.Printf("🚀 Generating Swift SDK...\n")

	config := SDKConfig{
		PackageName: packageName,
		BaseURL:     baseURL,
		Version:     "1.0.0",
		Language:    "swift",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Endpoints:   getAPIEndpoints(),
		Models:      getModels(),
	}

	swiftDir := filepath.Join(outputDir, "swift")
	sourcesDir := filepath.Join(swiftDir, "Sources", packageName)
	os.MkdirAll(sourcesDir, 0755)

	generateSwiftPackageSwift(swiftDir, config)
	generateSwiftClient(sourcesDir, config)
	generateSwiftModels(sourcesDir, config)
	generateSwiftServices(sourcesDir, config)
	generateSwiftReadme(swiftDir, config)

	fmt.Printf("✅ Swift SDK generated in: %s\n", swiftDir)
}

// Generate Kotlin SDK
func generateKotlinSDK(outputDir, packageName, baseURL string) {
	fmt.Printf("🚀 Generating Kotlin SDK...\n")

	config := SDKConfig{
		PackageName: packageName,
		BaseURL:     baseURL,
		Version:     "1.0.0",
		Language:    "kotlin",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Endpoints:   getAPIEndpoints(),
		Models:      getModels(),
	}

	kotlinDir := filepath.Join(outputDir, "kotlin")
	srcMainKotlinDir := filepath.Join(kotlinDir, "src", "main", "kotlin", strings.ReplaceAll(packageName, ".", "/"))
	os.MkdirAll(srcMainKotlinDir, 0755)

	generateKotlinBuildGradle(kotlinDir, config)
	generateKotlinClient(srcMainKotlinDir, config)
	generateKotlinModels(srcMainKotlinDir, config)
	generateKotlinServices(srcMainKotlinDir, config)
	generateKotlinReadme(kotlinDir, config)

	fmt.Printf("✅ Kotlin SDK generated in: %s\n", kotlinDir)
}

// Generate Dart SDK
func generateDartSDK(outputDir, packageName, baseURL string) {
	fmt.Printf("🚀 Generating Dart SDK...\n")

	config := SDKConfig{
		PackageName: packageName,
		BaseURL:     baseURL,
		Version:     "1.0.0",
		Language:    "dart",
		GeneratedAt: time.Now().Format(time.RFC3339),
		Endpoints:   getAPIEndpoints(),
		Models:      getModels(),
	}

	dartDir := filepath.Join(outputDir, "dart")
	libDir := filepath.Join(dartDir, "lib")
	modelsDir := filepath.Join(libDir, "models")
	servicesDir := filepath.Join(libDir, "services")

	os.MkdirAll(modelsDir, 0755)
	os.MkdirAll(servicesDir, 0755)

	generateDartPubspec(dartDir, config)
	generateDartClient(libDir, config)
	generateDartModels(modelsDir, config)
	generateDartServices(servicesDir, config)
	generateDartReadme(dartDir, config)

	fmt.Printf("✅ Dart SDK generated in: %s\n", dartDir)
}

// Helper function to write template content to a file
func writeTemplateToFile(filePath, templateContent string, data interface{}) error {
	tmpl, err := template.New(filepath.Base(filePath)).Parse(templateContent)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", filePath, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

// Helper function to get API endpoints (mock data for now)
func getAPIEndpoints() []SDKAPIEndpoint {
	return []SDKAPIEndpoint{
		{
			Method:      "POST",
			Path:        "/api/v1/auth/register",
			Description: "Register a new user",
			Parameters: []SDKAPIParameter{
				{Name: "email", Type: "string", Required: true, Description: "User email"},
				{Name: "password", Type: "string", Required: true, Description: "User password"},
				{Name: "name", Type: "string", Required: false, Description: "User name"},
			},
			Auth: false,
		},
		{
			Method:      "POST",
			Path:        "/api/v1/auth/login",
			Description: "Login user",
			Parameters: []SDKAPIParameter{
				{Name: "email", Type: "string", Required: true, Description: "User email"},
				{Name: "password", Type: "string", Required: true, Description: "User password"},
			},
			Auth: false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/users",
			Description: "Get all users",
			Parameters:  []SDKAPIParameter{},
			Auth:        true,
		},
		{
			Method:      "POST",
			Path:        "/api/v1/users",
			Description: "Create a new user",
			Parameters: []SDKAPIParameter{
				{Name: "name", Type: "string", Required: true, Description: "User name"},
				{Name: "email", Type: "string", Required: true, Description: "User email"},
			},
			Auth: true,
		},
	}
}

// Helper function to get models (mock data for now)
func getModels() []Model {
	return []Model{
		{
			Name:        "User",
			Description: "User model",
			Fields: []Field{
				{Name: "id", Type: "string", Required: true, Description: "User ID", IsID: true},
				{Name: "email", Type: "string", Required: true, Description: "User email"},
				{Name: "name", Type: "string", Required: false, Description: "User name"},
				{Name: "createdAt", Type: "string", Required: true, Description: "Creation date", IsTimestamp: true},
			},
		},
		{
			Name:        "Product",
			Description: "Product model",
			Fields: []Field{
				{Name: "id", Type: "string", Required: true, Description: "Product ID", IsID: true},
				{Name: "name", Type: "string", Required: true, Description: "Product name"},
				{Name: "price", Type: "number", Required: true, Description: "Product price"},
				{Name: "description", Type: "string", Required: false, Description: "Product description"},
			},
		},
	}
}

// TypeScript SDK generation functions
func generateTypeScriptPackageJSON(outputDir string, config SDKConfig) {
	templateContent := `{
  "name": "{{.PackageName}}",
  "version": "{{.Version}}",
  "description": "TypeScript SDK for Mobile Backend API",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "dev": "tsc --watch",
    "test": "jest",
    "lint": "eslint src/**/*.ts",
    "prepublishOnly": "npm run build"
  },
  "keywords": [
    "mobile",
    "backend",
    "api",
    "sdk",
    "typescript"
  ],
  "author": "Mobile Backend Team",
  "license": "MIT",
  "dependencies": {
    "axios": "^1.6.0",
    "ws": "^8.14.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "@types/ws": "^8.5.0",
    "@typescript-eslint/eslint-plugin": "^6.0.0",
    "@typescript-eslint/parser": "^6.0.0",
    "eslint": "^8.0.0",
    "jest": "^29.0.0",
    "typescript": "^5.0.0"
  },
  "files": [
    "dist/**/*",
    "README.md"
  ],
  "repository": {
    "type": "git",
    "url": "https://github.com/your-org/mobile-backend-sdk.git"
  },
  "bugs": {
    "url": "https://github.com/your-org/mobile-backend-sdk/issues"
  },
  "homepage": "https://github.com/your-org/mobile-backend-sdk#readme"
}`

	writeTemplateToFile(filepath.Join(outputDir, "package.json"), templateContent, config)
}

func generateTypeScriptConfig(outputDir string, config SDKConfig) {
	tsconfigContent := `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "lib": ["ES2020", "DOM"],
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "removeComments": false,
    "noImplicitAny": true,
    "noImplicitReturns": true,
    "noImplicitThis": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "exactOptionalPropertyTypes": true,
    "noImplicitOverride": true,
    "noPropertyAccessFromIndexSignature": true,
    "noUncheckedIndexedAccess": true,
    "allowUnusedLabels": false,
    "allowUnreachableCode": false,
    "resolveJsonModule": true,
    "moduleResolution": "node",
    "baseUrl": "./src",
    "paths": {
      "@/*": ["*"]
    }
  },
  "include": [
    "src/**/*"
  ],
  "exclude": [
    "node_modules",
    "dist",
    "**/*.test.ts",
    "**/*.spec.ts"
  ]
}`

	writeTemplateToFile(filepath.Join(outputDir, "tsconfig.json"), tsconfigContent, config)
}

func generateTypeScriptClient(outputDir string, config SDKConfig) {
	clientContent := `import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import WebSocket from 'ws';
import { AuthService } from './services/AuthService';
import { UserService } from './services/UserService';
import { ProductService } from './services/ProductService';
import { OrderService } from './services/OrderService';

export interface MobileBackendConfig {
  apiKey?: string;
  baseURL: string;
  timeout?: number;
  retries?: number;
  debug?: boolean;
}

export interface ApiResponse<T = any> {
  data: T;
  message?: string;
  success: boolean;
  timestamp: string;
}

export interface ApiError {
  message: string;
  code: string;
  details?: any;
  timestamp: string;
}

export class MobileBackendError extends Error {
  public code: string;
  public details?: any;
  public timestamp: string;

  constructor(error: ApiError) {
    super(error.message);
    this.name = 'MobileBackendError';
    this.code = error.code;
    this.details = error.details;
    this.timestamp = error.timestamp;
  }
}

export class MobileBackendClient {
  private axios: AxiosInstance;
  private ws?: WebSocket;
  private config: MobileBackendConfig;

  // Services
  public auth: AuthService;
  public users: UserService;
  public products: ProductService;
  public orders: OrderService;

  constructor(config: MobileBackendConfig) {
    this.config = {
      timeout: 30000,
      retries: 3,
      debug: false,
      ...config,
    };

    // Initialize Axios instance
    this.axios = axios.create({
      baseURL: this.config.baseURL,
      timeout: this.config.timeout,
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': 'MobileBackend-SDK-TypeScript/{{.Version}}',
      },
    });

    // Add request interceptor for authentication
    this.axios.interceptors.request.use(
      (config) => {
        const token = this.auth?.getToken();
        if (token) {
          config.headers.Authorization = 'Bearer ' + token;
        }
        if (this.config.apiKey) {
          config.headers['X-API-Key'] = this.config.apiKey;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Add response interceptor for error handling
    this.axios.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.data) {
          throw new MobileBackendError(error.response.data);
        }
        throw error;
      }
    );

    // Initialize services
    this.auth = new AuthService(this.axios);
    this.users = new UserService(this.axios);
    this.products = new ProductService(this.axios);
    this.orders = new OrderService(this.axios);
  }

  /**
   * Make a raw API request
   */
  async request<T = any>(config: AxiosRequestConfig): Promise<ApiResponse<T>> {
    try {
      const response: AxiosResponse<ApiResponse<T>> = await this.axios.request(config);
      return response.data;
    } catch (error) {
      if (error instanceof MobileBackendError) {
        throw error;
      }
      throw new MobileBackendError({
        message: error.message || 'Unknown error occurred',
        code: 'UNKNOWN_ERROR',
        timestamp: new Date().toISOString(),
      });
    }
  }

  /**
   * Connect to WebSocket for real-time updates
   */
  connectWebSocket(): Promise<void> {
    return new Promise((resolve, reject) => {
      const wsUrl = this.config.baseURL.replace('http', 'ws') + '/ws';
      this.ws = new WebSocket(wsUrl);

      this.ws.on('open', () => {
        if (this.config.debug) {
          console.log('WebSocket connected');
        }
        resolve();
      });

      this.ws.on('error', (error) => {
        if (this.config.debug) {
          console.error('WebSocket error:', error);
        }
        reject(error);
      });

      this.ws.on('close', () => {
        if (this.config.debug) {
          console.log('WebSocket disconnected');
        }
      });
    });
  }

  /**
   * Disconnect from WebSocket
   */
  disconnectWebSocket(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = undefined;
    }
  }

  /**
   * Send message via WebSocket
   */
  sendWebSocketMessage(message: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      throw new Error('WebSocket is not connected');
    }
  }

  /**
   * Listen to WebSocket messages
   */
  onWebSocketMessage(callback: (message: any) => void): void {
    if (this.ws) {
      this.ws.on('message', (data) => {
        try {
          const message = JSON.parse(data.toString());
          callback(message);
        } catch (error) {
          if (this.config.debug) {
            console.error('Failed to parse WebSocket message:', error);
          }
        }
      });
    }
  }

  /**
   * Get current configuration
   */
  getConfig(): MobileBackendConfig {
    return { ...this.config };
  }

  /**
   * Update configuration
   */
  updateConfig(newConfig: Partial<MobileBackendConfig>): void {
    this.config = { ...this.config, ...newConfig };

    // Update Axios instance
    this.axios.defaults.baseURL = this.config.baseURL;
    this.axios.defaults.timeout = this.config.timeout;
  }

  /**
   * Health check
   */
  async healthCheck(): Promise<boolean> {
    try {
      await this.request({ method: 'GET', url: '/health' });
      return true;
    } catch {
      return false;
    }
  }
}`

	writeTemplateToFile(filepath.Join(outputDir, "client.ts"), clientContent, config)
}

func generateTypeScriptServices(outputDir string, config SDKConfig) {
	// Generate AuthService
	authServiceContent := `import { AxiosInstance } from 'axios';
import { ApiResponse } from '../client';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name?: string;
}

export interface AuthResponse {
  token: string;
  user: {
    id: string;
    email: string;
    name?: string;
    createdAt: string;
  };
  expiresAt: string;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export class AuthService {
  private token: string | null = null;
  private refreshToken: string | null = null;

  constructor(private axios: AxiosInstance) {
    // Try to load token from localStorage in browser environment
    if (typeof window !== 'undefined' && window.localStorage) {
      this.token = localStorage.getItem('mobile_backend_token');
      this.refreshToken = localStorage.getItem('mobile_backend_refresh_token');
    }
  }

  /**
   * Register a new user
   */
  async register(data: RegisterRequest): Promise<ApiResponse<AuthResponse>> {
    const response = await this.axios.post<ApiResponse<AuthResponse>>('/api/v1/auth/register', data);
    this.setTokens(response.data.data);
    return response.data;
  }

  /**
   * Login user
   */
  async login(data: LoginRequest): Promise<ApiResponse<AuthResponse>> {
    const response = await this.axios.post<ApiResponse<AuthResponse>>('/api/v1/auth/login', data);
    this.setTokens(response.data.data);
    return response.data;
  }

  /**
   * Logout user
   */
  async logout(): Promise<ApiResponse<void>> {
    try {
      const response = await this.axios.post<ApiResponse<void>>('/api/v1/auth/logout');
      this.clearTokens();
      return response.data;
    } catch (error) {
      // Clear tokens even if logout fails
      this.clearTokens();
      throw error;
    }
  }

  /**
   * Get current user profile
   */
  async getProfile(): Promise<ApiResponse<AuthResponse['user']>> {
    const response = await this.axios.get<ApiResponse<AuthResponse['user']>>('/api/v1/profile');
    return response.data;
  }

  /**
   * Get current authentication token
   */
  getToken(): string | null {
    return this.token;
  }

  /**
   * Check if user is authenticated
   */
  isAuthenticated(): boolean {
    return this.token !== null;
  }

  /**
   * Set authentication tokens
   */
  private setTokens(authData: AuthResponse): void {
    this.token = authData.token;
    
    // Save to localStorage in browser environment
    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.setItem('mobile_backend_token', authData.token);
    }
  }

  /**
   * Clear authentication tokens
   */
  private clearTokens(): void {
    this.token = null;
    this.refreshToken = null;
    
    // Remove from localStorage in browser environment
    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.removeItem('mobile_backend_token');
      localStorage.removeItem('mobile_backend_refresh_token');
    }
  }
}`

	writeTemplateToFile(filepath.Join(outputDir, "AuthService.ts"), authServiceContent, config)

	// Generate other services (simplified versions)
	userServiceContent := `import { AxiosInstance } from 'axios';
import { ApiResponse } from '../client';

export interface User {
  id: string;
  email: string;
  name?: string;
  avatar?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateUserRequest {
  email: string;
  name?: string;
  password: string;
}

export interface UpdateUserRequest {
  name?: string;
  avatar?: string;
  isActive?: boolean;
}

export interface UserListParams {
  page?: number;
  limit?: number;
  search?: string;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface UserListResponse {
  users: User[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export class UserService {
  constructor(private axios: AxiosInstance) {}

  /**
   * Get all users with pagination and filtering
   */
  async list(params: UserListParams = {}): Promise<ApiResponse<UserListResponse>> {
    const response = await this.axios.get<ApiResponse<UserListResponse>>('/api/v1/users', { params });
    return response.data;
  }

  /**
   * Get user by ID
   */
  async getById(id: string): Promise<ApiResponse<User>> {
    const response = await this.axios.get<ApiResponse<User>>('/api/v1/users/' + id);
    return response.data;
  }

  /**
   * Create a new user
   */
  async create(data: CreateUserRequest): Promise<ApiResponse<User>> {
    const response = await this.axios.post<ApiResponse<User>>('/api/v1/users', data);
    return response.data;
  }

  /**
   * Update user by ID
   */
  async update(id: string, data: UpdateUserRequest): Promise<ApiResponse<User>> {
    const response = await this.axios.put<ApiResponse<User>>('/api/v1/users/' + id, data);
    return response.data;
  }

  /**
   * Delete user by ID
   */
  async delete(id: string): Promise<ApiResponse<void>> {
    const response = await this.axios.delete<ApiResponse<void>>('/api/v1/users/' + id);
    return response.data;
  }
}`

	writeTemplateToFile(filepath.Join(outputDir, "UserService.ts"), userServiceContent, config)

	// Generate ProductService and OrderService (simplified)
	productServiceContent := `import { AxiosInstance } from 'axios';
import { ApiResponse } from '../client';

export interface Product {
  id: string;
  name: string;
  description?: string;
  price: number;
  category: string;
  images: string[];
  isActive: boolean;
  stock: number;
  createdAt: string;
  updatedAt: string;
}

export class ProductService {
  constructor(private axios: AxiosInstance) {}

  async list(): Promise<ApiResponse<Product[]>> {
    const response = await this.axios.get<ApiResponse<Product[]>>('/api/v1/products');
    return response.data;
  }

  async getById(id: string): Promise<ApiResponse<Product>> {
    const response = await this.axios.get<ApiResponse<Product>>('/api/v1/products/' + id);
    return response.data;
  }
}`

	writeTemplateToFile(filepath.Join(outputDir, "ProductService.ts"), productServiceContent, config)

	orderServiceContent := `import { AxiosInstance } from 'axios';
import { ApiResponse } from '../client';

export interface Order {
  id: string;
  userId: string;
  total: number;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export class OrderService {
  constructor(private axios: AxiosInstance) {}

  async list(): Promise<ApiResponse<Order[]>> {
    const response = await this.axios.get<ApiResponse<Order[]>>('/api/v1/orders');
    return response.data;
  }

  async getById(id: string): Promise<ApiResponse<Order>> {
    const response = await this.axios.get<ApiResponse<Order>>('/api/v1/orders/' + id);
    return response.data;
  }
}`

	writeTemplateToFile(filepath.Join(outputDir, "OrderService.ts"), orderServiceContent, config)
}

func generateTypeScriptTypes(outputDir string, config SDKConfig) {
	typesContent := `// Common types and interfaces for Mobile Backend SDK

export interface PaginationParams {
  page?: number;
  limit?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface PaginationResponse {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface SearchParams extends PaginationParams {
  search?: string;
}

export interface ApiError {
  message: string;
  code: string;
  details?: any;
  timestamp: string;
}

export interface ValidationError {
  field: string;
  message: string;
  code: string;
}

export interface ApiResponse<T = any> {
  data: T;
  message?: string;
  success: boolean;
  timestamp: string;
  errors?: ValidationError[];
}

export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: string;
}

export interface SDKConfig {
  apiKey?: string;
  baseURL: string;
  timeout?: number;
  retries?: number;
  debug?: boolean;
}`

	writeTemplateToFile(filepath.Join(outputDir, "index.ts"), typesContent, config)
}

func generateTypeScriptIndex(outputDir string, config SDKConfig) {
	indexContent := `// Mobile Backend SDK for TypeScript/JavaScript
// Generated on {{.GeneratedAt}}

export { MobileBackendClient, MobileBackendError } from './client';
export type { 
  MobileBackendConfig, 
  ApiResponse, 
  ApiError 
} from './client';

// Services
export { AuthService } from './services/AuthService';
export type { 
  LoginRequest, 
  RegisterRequest, 
  AuthResponse, 
  RefreshTokenRequest 
} from './services/AuthService';

export { UserService } from './services/UserService';
export type { 
  User, 
  CreateUserRequest, 
  UpdateUserRequest, 
  UserListParams, 
  UserListResponse 
} from './services/UserService';

export { ProductService } from './services/ProductService';
export type { Product } from './services/ProductService';

export { OrderService } from './services/OrderService';
export type { Order } from './services/OrderService';

// Common types
export type {
  PaginationParams,
  PaginationResponse,
  SearchParams,
  ValidationError,
  WebSocketMessage
} from './types';

// Default export
export default MobileBackendClient;`

	writeTemplateToFile(filepath.Join(outputDir, "index.ts"), indexContent, config)
}

func generateTypeScriptReadme(outputDir string, config SDKConfig) {
	readmeContent := `# {{.PackageName}}

TypeScript/JavaScript SDK for Mobile Backend API

## Installation

` + "```" + `bash
npm install {{.PackageName}}
# or
yarn add {{.PackageName}}
# or
pnpm add {{.PackageName}}
` + "```" + `

## Quick Start

` + "```" + `typescript
import { MobileBackendClient } from '{{.PackageName}}';

// Initialize the client
const client = new MobileBackendClient({
  baseURL: '{{.BaseURL}}',
  apiKey: 'your-api-key', // optional
  debug: true // optional
});

// Authentication
await client.auth.login({
  email: 'user@example.com',
  password: 'password123'
});

// Use the services
const users = await client.users.list();
const products = await client.products.list();
const orders = await client.orders.list();
` + "```" + `

## Features

- ✅ **Type Safety**: Full TypeScript definitions
- ✅ **Auto-completion**: IDE support for all methods
- ✅ **Error Handling**: Consistent error handling
- ✅ **Authentication**: Built-in auth management
- ✅ **Real-time**: WebSocket integration
- ✅ **Browser & Node.js**: Works in both environments

## Services

### Authentication Service

` + "```" + `typescript
// Register a new user
await client.auth.register({
  email: 'user@example.com',
  password: 'password123',
  name: 'John Doe'
});

// Login
await client.auth.login({
  email: 'user@example.com',
  password: 'password123'
});

// Get current user
const profile = await client.auth.getProfile();

// Logout
await client.auth.logout();
` + "```" + `

### User Service

` + "```" + `typescript
// List users with pagination
const users = await client.users.list({
  page: 1,
  limit: 10,
  search: 'john'
});

// Get user by ID
const user = await client.users.getById('user-id');

// Create user
const newUser = await client.users.create({
  email: 'new@example.com',
  name: 'New User',
  password: 'password123'
});
` + "```" + `

## Real-time Features

` + "```" + `typescript
// Connect to WebSocket
await client.connectWebSocket();

// Listen to messages
client.onWebSocketMessage((message) => {
  console.log('Received:', message);
});

// Send message
client.sendWebSocketMessage({
  type: 'ping',
  data: { timestamp: Date.now() }
});
` + "```" + `

## Error Handling

` + "```" + `typescript
import { MobileBackendError } from '{{.PackageName}}';

try {
  await client.users.getById('invalid-id');
} catch (error) {
  if (error instanceof MobileBackendError) {
    console.error('API Error:', error.message);
    console.error('Error Code:', error.code);
  } else {
    console.error('Unknown error:', error);
  }
}
` + "```" + `

## Development

` + "```" + `bash
# Install dependencies
npm install

# Build the SDK
npm run build

# Run tests
npm test

# Lint code
npm run lint
` + "```" + `

## License

MIT

## Support

- GitHub Issues: [Report a bug](https://github.com/your-org/mobile-backend-sdk/issues)
- Documentation: [Read the docs]({{.BaseURL}}/docs)
- Email: support@example.com`

	writeTemplateToFile(filepath.Join(outputDir, "README.md"), readmeContent, config)
}

// Swift SDK generation functions
func generateSwiftPackageSwift(outputDir string, config SDKConfig) {
	packageSwiftContent := `// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "{{.PackageName}}",
    platforms: [
        .iOS(.v13),
        .macOS(.v10_15),
        .watchOS(.v6),
        .tvOS(.v13)
    ],
    products: [
        .library(
            name: "{{.PackageName}}",
            targets: ["{{.PackageName}}"]
        ),
    ],
    dependencies: [
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.8.0"),
        .package(url: "https://github.com/SwiftyJSON/SwiftyJSON.git", from: "5.0.0")
    ],
    targets: [
        .target(
            name: "{{.PackageName}}",
            dependencies: [
                "Alamofire",
                "SwiftyJSON"
            ]
        ),
        .testTarget(
            name: "{{.PackageName}}Tests",
            dependencies: ["{{.PackageName}}"]
        ),
    ]
)`

	writeTemplateToFile(filepath.Join(outputDir, "Package.swift"), packageSwiftContent, config)
}

func generateSwiftClient(outputDir string, config SDKConfig) {
	clientContent := `import Foundation
import Alamofire
import SwiftyJSON

public class MobileBackendClient {
    private let baseURL: String
    private let apiKey: String?
    private let session: Session
    
    public init(baseURL: String, apiKey: String? = nil) {
        self.baseURL = baseURL
        self.apiKey = apiKey
        
        let configuration = URLSessionConfiguration.default
        configuration.timeoutIntervalForRequest = 30
        self.session = Session(configuration: configuration)
    }
    
    public func request<T: Codable>(
        endpoint: String,
        method: HTTPMethod = .get,
        parameters: Parameters? = nil,
        responseType: T.Type,
        completion: @escaping (Result<T, Error>) -> Void
    ) {
        var headers: HTTPHeaders = [
            "Content-Type": "application/json"
        ]
        
        if let apiKey = apiKey {
            headers["X-API-Key"] = apiKey
        }
        
        session.request(
            baseURL + endpoint,
            method: method,
            parameters: parameters,
            encoding: JSONEncoding.default,
            headers: headers
        )
        .validate()
        .responseData { response in
            switch response.result {
            case .success(let data):
                do {
                    let decodedResponse = try JSONDecoder().decode(T.self, from: data)
                    completion(.success(decodedResponse))
                } catch {
                    completion(.failure(error))
                }
            case .failure(let error):
                completion(.failure(error))
            }
        }
    }
}`

	writeTemplateToFile(filepath.Join(outputDir, "Sources/{{.PackageName}}/Client.swift"), clientContent, config)
}

func generateSwiftModels(outputDir string, config SDKConfig) {
	modelsContent := `import Foundation

public struct User: Codable {
    public let id: String
    public let email: String
    public let name: String?
    public let createdAt: String
    public let updatedAt: String
    
    public init(id: String, email: String, name: String?, createdAt: String, updatedAt: String) {
        self.id = id
        self.email = email
        self.name = name
        self.createdAt = createdAt
        self.updatedAt = updatedAt
    }
}

public struct Product: Codable {
    public let id: String
    public let name: String
    public let description: String?
    public let price: Double
    public let category: String
    public let createdAt: String
    public let updatedAt: String
    
    public init(id: String, name: String, description: String?, price: Double, category: String, createdAt: String, updatedAt: String) {
        self.id = id
        self.name = name
        self.description = description
        self.price = price
        self.category = category
        self.createdAt = createdAt
        self.updatedAt = updatedAt
    }
}`

	writeTemplateToFile(filepath.Join(outputDir, "Sources/{{.PackageName}}/Models.swift"), modelsContent, config)
}

func generateSwiftServices(outputDir string, config SDKConfig) {
	servicesContent := `import Foundation

public class AuthService {
    private let client: MobileBackendClient
    
    public init(client: MobileBackendClient) {
        self.client = client
    }
    
    public func login(email: String, password: String, completion: @escaping (Result<AuthResponse, Error>) -> Void) {
        let parameters: [String: Any] = [
            "email": email,
            "password": password
        ]
        
        client.request(
            endpoint: "/api/v1/auth/login",
            method: .post,
            parameters: parameters,
            responseType: AuthResponse.self,
            completion: completion
        )
    }
}

public struct AuthResponse: Codable {
    public let token: String
    public let user: User
    public let expiresAt: String
}`

	writeTemplateToFile(filepath.Join(outputDir, "Sources/{{.PackageName}}/Services.swift"), servicesContent, config)
}

func generateSwiftReadme(outputDir string, config SDKConfig) {
	readmeContent := `# {{.PackageName}}

Swift SDK for Mobile Backend API

## Installation

### Swift Package Manager

Add the following to your Package.swift file:

` + "```" + `swift
dependencies: [
    .package(url: "https://github.com/your-org/{{.PackageName}}.git", from: "1.0.0")
]
` + "```" + `

## Usage

` + "```" + `swift
import {{.PackageName}}

let client = MobileBackendClient(baseURL: "{{.BaseURL}}", apiKey: "your-api-key")

// Authentication
let authService = AuthService(client: client)
authService.login(email: "user@example.com", password: "password123") { result in
    switch result {
    case .success(let response):
        print("Login successful: \\(response.user.email)")
    case .failure(let error):
        print("Login failed: \\(error)")
    }
}
` + "```" + `

## License

MIT`

	writeTemplateToFile(filepath.Join(outputDir, "README.md"), readmeContent, config)
}

// Kotlin SDK generation functions
func generateKotlinBuildGradle(outputDir string, config SDKConfig) {
	buildGradleContent := `plugins {
    id 'org.jetbrains.kotlin.jvm' version '1.9.0'
    id 'maven-publish'
}

group = 'com.mobilebackend'
version = '{{.Version}}'

repositories {
    mavenCentral()
}

dependencies {
    implementation 'com.squareup.retrofit2:retrofit:2.9.0'
    implementation 'com.squareup.retrofit2:converter-gson:2.9.0'
    implementation 'com.squareup.okhttp3:logging-interceptor:4.11.0'
    implementation 'org.jetbrains.kotlinx:kotlinx-coroutines-core:1.7.3'
    
    testImplementation 'junit:junit:4.13.2'
    testImplementation 'org.mockito:mockito-core:5.5.0'
}

kotlin {
    jvmToolchain(11)
}

publishing {
    publications {
        maven(MavenPublication) {
            from components.java
        }
    }
}`

	writeTemplateToFile(filepath.Join(outputDir, "build.gradle.kts"), buildGradleContent, config)
}

func generateKotlinClient(outputDir string, config SDKConfig) {
	clientContent := `package com.mobilebackend.sdk

import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import java.util.concurrent.TimeUnit

class MobileBackendClient(
    private val baseURL: String,
    private val apiKey: String? = null
) {
    private val retrofit: Retrofit
    
    init {
        val client = OkHttpClient.Builder()
            .addInterceptor { chain ->
                val request = chain.request().newBuilder()
                apiKey?.let { request.addHeader("X-API-Key", it) }
                chain.proceed(request.build())
            }
            .addInterceptor(HttpLoggingInterceptor().apply {
                level = HttpLoggingInterceptor.Level.BODY
            })
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .build()
            
        retrofit = Retrofit.Builder()
            .baseUrl(baseURL)
            .client(client)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }
    
    fun getAuthService(): AuthService = retrofit.create(AuthService::class.java)
    fun getUserService(): UserService = retrofit.create(UserService::class.java)
}`

	writeTemplateToFile(filepath.Join(outputDir, "src/main/kotlin/com/mobilebackend/sdk/Client.kt"), clientContent, config)
}

func generateKotlinModels(outputDir string, config SDKConfig) {
	modelsContent := `package com.mobilebackend.sdk

import com.google.gson.annotations.SerializedName

data class User(
    val id: String,
    val email: String,
    val name: String?,
    @SerializedName("created_at") val createdAt: String,
    @SerializedName("updated_at") val updatedAt: String
)

data class Product(
    val id: String,
    val name: String,
    val description: String?,
    val price: Double,
    val category: String,
    @SerializedName("created_at") val createdAt: String,
    @SerializedName("updated_at") val updatedAt: String
)

data class ApiResponse<T>(
    val data: T,
    val message: String?,
    val success: Boolean,
    val timestamp: String
)`

	writeTemplateToFile(filepath.Join(outputDir, "src/main/kotlin/com/mobilebackend/sdk/Models.kt"), modelsContent, config)
}

func generateKotlinServices(outputDir string, config SDKConfig) {
	servicesContent := `package com.mobilebackend.sdk

import retrofit2.http.*

interface AuthService {
    @POST("api/v1/auth/login")
    suspend fun login(@Body request: LoginRequest): ApiResponse<AuthResponse>
    
    @POST("api/v1/auth/register")
    suspend fun register(@Body request: RegisterRequest): ApiResponse<AuthResponse>
}

interface UserService {
    @GET("api/v1/users")
    suspend fun getUsers(): ApiResponse<List<User>>
    
    @GET("api/v1/users/{id}")
    suspend fun getUser(@Path("id") id: String): ApiResponse<User>
}

data class LoginRequest(
    val email: String,
    val password: String
)

data class RegisterRequest(
    val email: String,
    val password: String,
    val name: String?
)

data class AuthResponse(
    val token: String,
    val user: User,
    @SerializedName("expires_at") val expiresAt: String
)`

	writeTemplateToFile(filepath.Join(outputDir, "src/main/kotlin/com/mobilebackend/sdk/Services.kt"), servicesContent, config)
}

func generateKotlinReadme(outputDir string, config SDKConfig) {
	readmeContent := `# {{.PackageName}}

Kotlin SDK for Mobile Backend API

## Installation

### Gradle

Add to your ` + "`" + `build.gradle.kts` + "`" + `:

` + "```" + `kotlin
dependencies {
    implementation("com.mobilebackend:{{.PackageName}}:{{.Version}}")
}
` + "```" + `

## Usage

` + "```" + `kotlin
import com.mobilebackend.sdk.*

val client = MobileBackendClient("{{.BaseURL}}", "your-api-key")

// Authentication
val authService = client.getAuthService()
try {
    val response = authService.login(LoginRequest("user@example.com", "password123"))
    println("Login successful: \${response.data.user.email}")
} catch (e: Exception) {
    println("Login failed: \${e.message}")
}
` + "```" + `

## License

MIT`

	writeTemplateToFile(filepath.Join(outputDir, "README.md"), readmeContent, config)
}

// Dart SDK generation functions
func generateDartPubspec(outputDir string, config SDKConfig) {
	pubspecContent := `name: {{.PackageName}}
description: Dart SDK for Mobile Backend API
version: {{.Version}}
homepage: https://github.com/your-org/{{.PackageName}}

environment:
  sdk: '>=3.0.0 <4.0.0'

dependencies:
  http: ^1.1.0
  dio: ^5.3.0
  json_annotation: ^4.8.1

dev_dependencies:
  test: ^1.24.0
  json_serializable: ^6.7.1
  build_runner: ^2.4.7

dependency_overrides:`

	writeTemplateToFile(filepath.Join(outputDir, "pubspec.yaml"), pubspecContent, config)
}

func generateDartClient(outputDir string, config SDKConfig) {
	clientContent := `import 'package:dio/dio.dart';
import 'models/models.dart';
import 'services/services.dart';

class MobileBackendClient {
  late final Dio _dio;
  late final AuthService _authService;
  late final UserService _userService;
  
  MobileBackendClient({
    required String baseUrl,
    String? apiKey,
    Duration? timeout,
  }) {
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      connectTimeout: timeout ?? const Duration(seconds: 30),
      receiveTimeout: timeout ?? const Duration(seconds: 30),
      headers: {
        'Content-Type': 'application/json',
        if (apiKey != null) 'X-API-Key': apiKey,
      },
    ));
    
    _authService = AuthService(_dio);
    _userService = UserService(_dio);
  }
  
  AuthService get auth => _authService;
  UserService get users => _userService;
}`

	writeTemplateToFile(filepath.Join(outputDir, "client.dart"), clientContent, config)
}

func generateDartModels(outputDir string, config SDKConfig) {
	modelsContent := `import 'package:json_annotation/json_annotation.dart';

part 'models.g.dart';

@JsonSerializable()
class User {
  final String id;
  final String email;
  final String? name;
  @JsonKey(name: 'created_at')
  final String createdAt;
  @JsonKey(name: 'updated_at')
  final String updatedAt;
  
  User({
    required this.id,
    required this.email,
    this.name,
    required this.createdAt,
    required this.updatedAt,
  });
  
  factory User.fromJson(Map<String, dynamic> json) => _$UserFromJson(json);
  Map<String, dynamic> toJson() => _$UserToJson(this);
}

@JsonSerializable()
class ApiResponse<T> {
  final T data;
  final String? message;
  final bool success;
  final String timestamp;
  
  ApiResponse({
    required this.data,
    this.message,
    required this.success,
    required this.timestamp,
  });
  
  factory ApiResponse.fromJson(Map<String, dynamic> json) => _$ApiResponseFromJson(json);
  Map<String, dynamic> toJson() => _$ApiResponseToJson(this);
}`

	writeTemplateToFile(filepath.Join(outputDir, "models.dart"), modelsContent, config)
}

func generateDartServices(outputDir string, config SDKConfig) {
	servicesContent := `import 'package:dio/dio.dart';
import '../models/models.dart';

class AuthService {
  final Dio _dio;
  
  AuthService(this._dio);
  
  Future<ApiResponse<AuthResponse>> login(String email, String password) async {
    final response = await _dio.post('/api/v1/auth/login', data: {
      'email': email,
      'password': password,
    });
    return ApiResponse<AuthResponse>.fromJson(response.data);
  }
  
  Future<ApiResponse<AuthResponse>> register(String email, String password, {String? name}) async {
    final response = await _dio.post('/api/v1/auth/register', data: {
      'email': email,
      'password': password,
      if (name != null) 'name': name,
    });
    return ApiResponse<AuthResponse>.fromJson(response.data);
  }
}

class UserService {
  final Dio _dio;
  
  UserService(this._dio);
  
  Future<ApiResponse<List<User>>> getUsers() async {
    final response = await _dio.get('/api/v1/users');
    return ApiResponse<List<User>>.fromJson(response.data);
  }
  
  Future<ApiResponse<User>> getUser(String id) async {
    final response = await _dio.get('/api/v1/users/$id');
    return ApiResponse<User>.fromJson(response.data);
  }
}

@JsonSerializable()
class AuthResponse {
  final String token;
  final User user;
  @JsonKey(name: 'expires_at')
  final String expiresAt;
  
  AuthResponse({
    required this.token,
    required this.user,
    required this.expiresAt,
  });
  
  factory AuthResponse.fromJson(Map<String, dynamic> json) => _$AuthResponseFromJson(json);
  Map<String, dynamic> toJson() => _$AuthResponseToJson(this);
}`

	writeTemplateToFile(filepath.Join(outputDir, "services.dart"), servicesContent, config)
}

func generateDartReadme(outputDir string, config SDKConfig) {
	readmeContent := `# {{.PackageName}}

Dart SDK for Mobile Backend API

## Installation

Add to your ` + "`" + `pubspec.yaml` + "`" + `:

` + "```" + `yaml
dependencies:
  {{.PackageName}}: ^{{.Version}}
` + "```" + `

## Usage

` + "```" + `dart
import 'package:{{.PackageName}}/{{.PackageName}}.dart';

void main() async {
  final client = MobileBackendClient(
    baseUrl: '{{.BaseURL}}',
    apiKey: 'your-api-key',
  );
  
  // Authentication
  try {
    final response = await client.auth.login('user@example.com', 'password123');
    print('Login successful: \${response.data.user.email}');
  } catch (e) {
    print('Login failed: \$e');
  }
}
` + "```" + `

## License

MIT`

	writeTemplateToFile(filepath.Join(outputDir, "README.md"), readmeContent, config)
}
