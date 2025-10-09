package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	AllowOrigins           []string
	AllowMethods           []string
	AllowHeaders           []string
	ExposeHeaders          []string
	AllowCredentials       bool
	MaxAge                 time.Duration
	AllowWildcard          bool
	AllowBrowserExtensions bool
}

func CORS() gin.HandlerFunc {
	config := getCORSConfig()
	return cors.New(config)
}

func getCORSConfig() cors.Config {
	// Get allowed origins from environment or use defaults
	origins := getAllowedOrigins()

	// Get allowed methods from environment or use defaults
	methods := getStringSliceFromEnv("CORS_ALLOWED_METHODS", []string{
		"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
	})

	// Get allowed headers from environment or use defaults
	headers := getStringSliceFromEnv("CORS_ALLOWED_HEADERS", []string{
		"Origin", "Content-Length", "Content-Type", "Authorization",
		"X-Requested-With", "Accept", "X-API-Key", "X-Request-ID",
		"X-Trace-ID", "X-Span-ID", "X-Correlation-ID",
	})

	// Get exposed headers from environment or use defaults
	exposeHeaders := getStringSliceFromEnv("CORS_EXPOSED_HEADERS", []string{
		"Content-Length", "X-Total-Count", "X-Page-Count",
		"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset",
	})

	// Get max age from environment or use default
	maxAge := getDurationFromEnv("CORS_MAX_AGE", 12*time.Hour)

	// Allow credentials by default unless explicitly disabled
	allowCredentials := getBoolFromEnv("CORS_ALLOW_CREDENTIALS", true)

	// Allow wildcard origins (for development)
	allowWildcard := getBoolFromEnv("CORS_ALLOW_WILDCARD", false)

	// Allow browser extensions
	allowBrowserExtensions := getBoolFromEnv("CORS_ALLOW_BROWSER_EXTENSIONS", true)

	config := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     methods,
		AllowHeaders:     headers,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: allowCredentials,
		MaxAge:           maxAge,
	}

	// Add wildcard support if enabled
	if allowWildcard {
		config.AllowOriginFunc = func(origin string) bool {
			// Allow localhost and 127.0.0.1 for development
			if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
				return true
			}

			// Allow browser extensions
			if allowBrowserExtensions && (strings.HasPrefix(origin, "chrome-extension://") ||
				strings.HasPrefix(origin, "moz-extension://") ||
				strings.HasPrefix(origin, "safari-extension://")) {
				return true
			}

			// Check against allowed origins
			for _, allowedOrigin := range origins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					return true
				}
				// Support wildcard subdomains
				if strings.Contains(allowedOrigin, "*") {
					pattern := strings.Replace(allowedOrigin, "*", ".*", 1)
					if matched, _ := matchPattern(pattern, origin); matched {
						return true
					}
				}
			}

			return false
		}
	}

	return config
}

func getAllowedOrigins() []string {
	origins := getStringSliceFromEnv("CORS_ALLOWED_ORIGINS", []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:8080",
		"http://localhost:8081",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:3001",
		"http://127.0.0.1:8080",
		"http://127.0.0.1:8081",
	})

	// Add production origins if specified
	if prodOrigins := os.Getenv("CORS_PRODUCTION_ORIGINS"); prodOrigins != "" {
		prodList := strings.Split(prodOrigins, ",")
		for _, origin := range prodList {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				origins = append(origins, origin)
			}
		}
	}

	return origins
}

func getStringSliceFromEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getDurationFromEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolFromEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

// Simple pattern matching for wildcard support
func matchPattern(pattern, str string) (bool, error) {
	// This is a simplified implementation
	// In production, you might want to use a more robust pattern matching library
	if pattern == str {
		return true, nil
	}

	if strings.Contains(pattern, "*") {
		// Convert wildcard pattern to regex-like matching
		pattern = strings.Replace(pattern, "*", ".*", 1)
		pattern = "^" + pattern + "$"

		// Simple wildcard matching (not regex)
		if strings.HasPrefix(pattern, ".*") && strings.HasSuffix(pattern, ".*") {
			// Contains pattern
			subPattern := pattern[2 : len(pattern)-2]
			return strings.Contains(str, subPattern), nil
		} else if strings.HasPrefix(pattern, ".*") {
			// Ends with pattern
			suffix := pattern[2:]
			return strings.HasSuffix(str, suffix), nil
		} else if strings.HasSuffix(pattern, ".*") {
			// Starts with pattern
			prefix := pattern[:len(pattern)-2]
			return strings.HasPrefix(str, prefix), nil
		}
	}

	return false, nil
}

// PreflightHandler handles preflight requests
func PreflightHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization, X-Requested-With, Accept, X-API-Key, X-Request-ID, X-Trace-ID, X-Span-ID, X-Correlation-ID")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400") // 24 hours
			c.Status(200)
			c.Abort()
			return
		}
		c.Next()
	}
}
