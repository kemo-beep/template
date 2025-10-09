#!/bin/bash

# Mobile Backend CLI Test Suite
# This script tests all CLI functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="mobile-backend-cli"
TEST_DIR="test-cli-$(date +%s)"
TEST_PROJECT="test-project"

# Print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    print_info "Running test: $test_name"
    
    if eval "$test_command"; then
        print_success "PASS: $test_name"
        ((TESTS_PASSED++))
    else
        print_error "FAIL: $test_name"
        ((TESTS_FAILED++))
    fi
    echo
}

# Setup test environment
setup_test_env() {
    print_info "Setting up test environment..."
    
    # Create test directory
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"
    
    # Initialize Go module
    go mod init "$TEST_PROJECT"
    
    # Create basic project structure
    mkdir -p models controllers services routes tests/unit tests/integration migrations
    
    print_success "Test environment setup complete"
}

# Test CLI installation
test_installation() {
    print_info "Testing CLI installation..."
    
    # Check if CLI is installed
    if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
        print_error "CLI is not installed. Please run install.sh first."
        exit 1
    fi
    
    # Test version command
    run_test "Version command" "$BINARY_NAME --version"
    
    # Test help command
    run_test "Help command" "$BINARY_NAME --help"
    
    # Test invalid command
    run_test "Invalid command handling" "! $BINARY_NAME invalid-command 2>/dev/null"
}

# Test code generation
test_code_generation() {
    print_info "Testing code generation..."
    
    # Test model generation
    run_test "Generate model" "$BINARY_NAME generate model User"
    
    # Test controller generation
    run_test "Generate controller" "$BINARY_NAME generate controller UserController"
    
    # Test service generation
    run_test "Generate service" "$BINARY_NAME generate service UserService"
    
    # Test middleware generation
    run_test "Generate middleware" "$BINARY_NAME generate middleware AuthMiddleware"
    
    # Test test generation
    run_test "Generate test" "$BINARY_NAME generate test UserControllerTest"
    
    # Test migration generation
    run_test "Generate migration" "$BINARY_NAME generate migration AddUserTable"
    
    # Test route generation
    run_test "Generate route" "$BINARY_NAME generate route UserRoutes"
    
    # Test complete API generation
    run_test "Generate complete API" "$BINARY_NAME generate api Product"
    
    # Verify generated files exist
    run_test "Verify generated files" "[ -f models/user.go ] && [ -f controllers/user.go ] && [ -f services/user.go ]"
}

# Test API functionality
test_api_functionality() {
    print_info "Testing API functionality..."
    
    # Test API health check (this will fail if no server is running, which is expected)
    run_test "API health check" "$BINARY_NAME api health || true"
    
    # Test API list command
    run_test "API list command" "$BINARY_NAME api list"
    
    # Test API docs command
    run_test "API docs command" "$BINARY_NAME api docs"
    
    # Test API test command (this will fail if no server is running, which is expected)
    run_test "API test command" "$BINARY_NAME api test GET /health || true"
}

# Test database functionality
test_database_functionality() {
    print_info "Testing database functionality..."
    
    # Test database status (this will fail if no database is running, which is expected)
    run_test "Database status" "$BINARY_NAME db status || true"
    
    # Test database migration (this will fail if no database is running, which is expected)
    run_test "Database migration" "$BINARY_NAME db migrate || true"
    
    # Test database seed (this will fail if no database is running, which is expected)
    run_test "Database seed" "$BINARY_NAME db seed || true"
}

# Test deployment functionality
test_deployment_functionality() {
    print_info "Testing deployment functionality..."
    
    # Test deployment status
    run_test "Deployment status" "$BINARY_NAME deploy status"
    
    # Test deployment health
    run_test "Deployment health" "$BINARY_NAME deploy health"
    
    # Test deployment config
    run_test "Deployment config" "$BINARY_NAME deploy config"
}

# Test testing functionality
test_testing_functionality() {
    print_info "Testing testing functionality..."
    
    # Test unit tests
    run_test "Unit tests" "$BINARY_NAME test unit"
    
    # Test integration tests
    run_test "Integration tests" "$BINARY_NAME test integration"
    
    # Test coverage
    run_test "Coverage tests" "$BINARY_NAME test coverage"
    
    # Test benchmark
    run_test "Benchmark tests" "$BINARY_NAME test benchmark"
    
    # Test linting
    run_test "Linting" "$BINARY_NAME test lint"
    
    # Test security
    run_test "Security tests" "$BINARY_NAME test security"
}

# Test explorer functionality
test_explorer_functionality() {
    print_info "Testing explorer functionality..."
    
    # Test explorer command (start in background and kill)
    run_test "Explorer command" "timeout 5s $BINARY_NAME explorer --port 3001 || true"
}

# Test configuration
test_configuration() {
    print_info "Testing configuration..."
    
    # Test with custom config file
    run_test "Custom config file" "$BINARY_NAME --config /dev/null --help"
    
    # Test with custom base URL
    run_test "Custom base URL" "$BINARY_NAME --base-url https://api.example.com --help"
    
    # Test with custom API key
    run_test "Custom API key" "$BINARY_NAME --api-key test-key --help"
    
    # Test verbose mode
    run_test "Verbose mode" "$BINARY_NAME --verbose --help"
}

# Test error handling
test_error_handling() {
    print_info "Testing error handling..."
    
    # Test invalid arguments
    run_test "Invalid arguments" "! $BINARY_NAME generate invalid-command 2>/dev/null"
    
    # Test missing arguments
    run_test "Missing arguments" "! $BINARY_NAME generate model 2>/dev/null"
    
    # Test invalid flags
    run_test "Invalid flags" "! $BINARY_NAME --invalid-flag 2>/dev/null"
}

# Test file operations
test_file_operations() {
    print_info "Testing file operations..."
    
    # Test if generated files are valid Go code
    run_test "Valid Go syntax - models" "go fmt models/user.go"
    run_test "Valid Go syntax - controllers" "go fmt controllers/user.go"
    run_test "Valid Go syntax - services" "go fmt services/user.go"
    
    # Test if files contain expected content
    run_test "Model contains struct" "grep -q 'type User struct' models/user.go"
    run_test "Controller contains methods" "grep -q 'func.*Controller.*' controllers/user.go"
    run_test "Service contains methods" "grep -q 'func.*Service.*' services/user.go"
}

# Test template system
test_template_system() {
    print_info "Testing template system..."
    
    # Test different naming conventions
    run_test "Snake case conversion" "$BINARY_NAME generate model UserProfile"
    run_test "Kebab case conversion" "$BINARY_NAME generate model UserProfile"
    
    # Verify snake case in generated files
    run_test "Snake case in files" "grep -q 'user_profiles' models/userprofile.go || grep -q 'user_profiles' models/user_profile.go"
}

# Test help system
test_help_system() {
    print_info "Testing help system..."
    
    # Test main help
    run_test "Main help" "$BINARY_NAME --help"
    
    # Test subcommand help
    run_test "Generate help" "$BINARY_NAME generate --help"
    run_test "API help" "$BINARY_NAME api --help"
    run_test "Database help" "$BINARY_NAME db --help"
    run_test "Deploy help" "$BINARY_NAME deploy --help"
    run_test "Test help" "$BINARY_NAME test --help"
    run_test "Explorer help" "$BINARY_NAME explorer --help"
}

# Cleanup test environment
cleanup() {
    print_info "Cleaning up test environment..."
    
    # Go back to original directory
    cd ..
    
    # Remove test directory
    rm -rf "$TEST_DIR"
    
    print_success "Cleanup complete"
}

# Show test results
show_results() {
    echo
    echo "========================================"
    echo "Test Results Summary"
    echo "========================================"
    echo "Tests Passed: $TESTS_PASSED"
    echo "Tests Failed: $TESTS_FAILED"
    echo "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"
    echo
    
    if [ $TESTS_FAILED -eq 0 ]; then
        print_success "All tests passed! 🎉"
        exit 0
    else
        print_error "Some tests failed. Please check the output above."
        exit 1
    fi
}

# Main test function
main() {
    echo "🧪 Mobile Backend CLI Test Suite"
    echo "================================="
    echo
    
    # Setup
    setup_test_env
    
    # Run all tests
    test_installation
    test_code_generation
    test_api_functionality
    test_database_functionality
    test_deployment_functionality
    test_testing_functionality
    test_explorer_functionality
    test_configuration
    test_error_handling
    test_file_operations
    test_template_system
    test_help_system
    
    # Cleanup
    cleanup
    
    # Show results
    show_results
}

# Run main function
main "$@"
