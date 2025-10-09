package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	// Custom validators
	validate *validator.Validate

	// Pre-compiled regex patterns
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex    = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)
	slugRegex     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// Initialize custom validators
func init() {
	validate = validator.New()

	// Register custom validators
	validate.RegisterValidation("email", validateEmail)
	validate.RegisterValidation("phone", validatePhone)
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("slug", validateSlug)
	validate.RegisterValidation("uuid", validateUUID)
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("date", validateDate)
	validate.RegisterValidation("time", validateTime)
	validate.RegisterValidation("json", validateJSON)
	validate.RegisterValidation("url", validateURL)
	validate.RegisterValidation("ip", validateIP)
	validate.RegisterValidation("mac", validateMAC)
	validate.RegisterValidation("creditcard", validateCreditCard)
	validate.RegisterValidation("ssn", validateSSN)
	validate.RegisterValidation("postalcode", validatePostalCode)
	validate.RegisterValidation("currency", validateCurrency)
	validate.RegisterValidation("latitude", validateLatitude)
	validate.RegisterValidation("longitude", validateLongitude)
	validate.RegisterValidation("hexcolor", validateHexColor)
	validate.RegisterValidation("rgbcolor", validateRGBColor)
	validate.RegisterValidation("isbn", validateISBN)
}

// Custom validation functions
func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return true // Let required handle empty values
	}

	// Use Go's built-in email validation
	_, err := mail.ParseAddress(email)
	return err == nil
}

func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true
	}
	return phoneRegex.MatchString(phone)
}

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if username == "" {
		return true
	}
	return usernameRegex.MatchString(username)
}

func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	if slug == "" {
		return true
	}
	return slugRegex.MatchString(slug)
}

func validateUUID(fl validator.FieldLevel) bool {
	uuid := fl.Field().String()
	if uuid == "" {
		return true
	}
	return uuidRegex.MatchString(uuid)
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if password == "" {
		return true
	}

	// Password requirements:
	// - At least 8 characters
	// - At least one uppercase letter
	// - At least one lowercase letter
	// - At least one number
	// - At least one special character
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func validateDate(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	if dateStr == "" {
		return true
	}

	// Try common date formats
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"02/01/2006",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, dateStr); err == nil {
			return true
		}
	}

	return false
}

func validateTime(fl validator.FieldLevel) bool {
	timeStr := fl.Field().String()
	if timeStr == "" {
		return true
	}

	// Try common time formats
	formats := []string{
		"15:04:05",
		"15:04",
		"3:04 PM",
		"3:04:05 PM",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, timeStr); err == nil {
			return true
		}
	}

	return false
}

func validateJSON(fl validator.FieldLevel) bool {
	jsonStr := fl.Field().String()
	if jsonStr == "" {
		return true
	}

	// Basic JSON validation
	var temp interface{}
	return json.Unmarshal([]byte(jsonStr), &temp) == nil
}

func validateURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	if url == "" {
		return true
	}

	// Basic URL validation
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "ftp://")
}

func validateIP(fl validator.FieldLevel) bool {
	ip := fl.Field().String()
	if ip == "" {
		return true
	}

	// Basic IP validation
	return net.ParseIP(ip) != nil
}

func validateMAC(fl validator.FieldLevel) bool {
	mac := fl.Field().String()
	if mac == "" {
		return true
	}

	// MAC address validation
	_, err := net.ParseMAC(mac)
	return err == nil
}

func validateCreditCard(fl validator.FieldLevel) bool {
	card := fl.Field().String()
	if card == "" {
		return true
	}

	// Luhn algorithm for credit card validation
	return luhnCheck(card)
}

func validateSSN(fl validator.FieldLevel) bool {
	ssn := fl.Field().String()
	if ssn == "" {
		return true
	}

	// SSN format: XXX-XX-XXXX
	ssnRegex := regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`)
	return ssnRegex.MatchString(ssn)
}

func validatePostalCode(fl validator.FieldLevel) bool {
	postal := fl.Field().String()
	if postal == "" {
		return true
	}

	// US ZIP code validation
	zipRegex := regexp.MustCompile(`^\d{5}(-\d{4})?$`)
	return zipRegex.MatchString(postal)
}

func validateCurrency(fl validator.FieldLevel) bool {
	currency := fl.Field().String()
	if currency == "" {
		return true
	}

	// Currency code validation (ISO 4217)
	currencyRegex := regexp.MustCompile(`^[A-Z]{3}$`)
	return currencyRegex.MatchString(currency)
}

func validateLatitude(fl validator.FieldLevel) bool {
	lat := fl.Field().Float()
	return lat >= -90 && lat <= 90
}

func validateLongitude(fl validator.FieldLevel) bool {
	lng := fl.Field().Float()
	return lng >= -180 && lng <= 180
}

func validateHexColor(fl validator.FieldLevel) bool {
	color := fl.Field().String()
	if color == "" {
		return true
	}

	hexRegex := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
	return hexRegex.MatchString(color)
}

func validateRGBColor(fl validator.FieldLevel) bool {
	color := fl.Field().String()
	if color == "" {
		return true
	}

	rgbRegex := regexp.MustCompile(`^rgb\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*\)$`)
	return rgbRegex.MatchString(color)
}

func validateISBN(fl validator.FieldLevel) bool {
	isbn := fl.Field().String()
	if isbn == "" {
		return true
	}

	// Remove hyphens and spaces
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")

	// Check length
	if len(isbn) != 10 && len(isbn) != 13 {
		return false
	}

	// Validate based on length
	if len(isbn) == 10 {
		return validateISBN10(isbn)
	}
	return validateISBN13(isbn)
}

func validateISBN10(isbn string) bool {
	if len(isbn) != 10 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		if isbn[i] < '0' || isbn[i] > '9' {
			return false
		}
		sum += int(isbn[i]-'0') * (10 - i)
	}

	checkDigit := isbn[9]
	if checkDigit == 'X' {
		sum += 10
	} else if checkDigit >= '0' && checkDigit <= '9' {
		sum += int(checkDigit - '0')
	} else {
		return false
	}

	return sum%11 == 0
}

func validateISBN13(isbn string) bool {
	if len(isbn) != 13 {
		return false
	}

	sum := 0
	for i := 0; i < 12; i++ {
		if isbn[i] < '0' || isbn[i] > '9' {
			return false
		}
		multiplier := 1
		if i%2 == 1 {
			multiplier = 3
		}
		sum += int(isbn[i]-'0') * multiplier
	}

	checkDigit := int(isbn[12] - '0')
	if checkDigit < 0 || checkDigit > 9 {
		return false
	}

	return (10-(sum%10))%10 == checkDigit
}

// Luhn algorithm for credit card validation
func luhnCheck(card string) bool {
	// Remove spaces and non-digits
	card = regexp.MustCompile(`\D`).ReplaceAllString(card, "")

	if len(card) < 13 || len(card) > 19 {
		return false
	}

	sum := 0
	alternate := false

	// Process digits from right to left
	for i := len(card) - 1; i >= 0; i-- {
		digit := int(card[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// Validation helper functions
func ValidateStruct(s interface{}) []ValidationError {
	var errors []ValidationError

	if err := validate.Struct(s); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Field:   err.Field(),
				Message: getValidationMessage(err),
				Value:   err.Value(),
			})
		}
	}

	return errors
}

func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", err.Field())
	case "username":
		return fmt.Sprintf("%s must be 3-20 characters and contain only letters, numbers, hyphens, and underscores", err.Field())
	case "slug":
		return fmt.Sprintf("%s must be a valid slug (lowercase letters, numbers, and hyphens only)", err.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", err.Field())
	case "password":
		return fmt.Sprintf("%s must be at least 8 characters with uppercase, lowercase, number, and special character", err.Field())
	case "date":
		return fmt.Sprintf("%s must be a valid date", err.Field())
	case "time":
		return fmt.Sprintf("%s must be a valid time", err.Field())
	case "json":
		return fmt.Sprintf("%s must be valid JSON", err.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", err.Field())
	case "ip":
		return fmt.Sprintf("%s must be a valid IP address", err.Field())
	case "mac":
		return fmt.Sprintf("%s must be a valid MAC address", err.Field())
	case "creditcard":
		return fmt.Sprintf("%s must be a valid credit card number", err.Field())
	case "ssn":
		return fmt.Sprintf("%s must be a valid SSN (XXX-XX-XXXX)", err.Field())
	case "postalcode":
		return fmt.Sprintf("%s must be a valid postal code", err.Field())
	case "currency":
		return fmt.Sprintf("%s must be a valid currency code", err.Field())
	case "latitude":
		return fmt.Sprintf("%s must be between -90 and 90", err.Field())
	case "longitude":
		return fmt.Sprintf("%s must be between -180 and 180", err.Field())
	case "hexcolor":
		return fmt.Sprintf("%s must be a valid hex color", err.Field())
	case "rgbcolor":
		return fmt.Sprintf("%s must be a valid RGB color", err.Field())
	case "isbn":
		return fmt.Sprintf("%s must be a valid ISBN", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", err.Field(), err.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", err.Field(), err.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", err.Field(), err.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", err.Field(), err.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", err.Field(), err.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", err.Field())
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", err.Field())
	case "numeric":
		return fmt.Sprintf("%s must contain only numbers", err.Field())
	case "alphaunicode":
		return fmt.Sprintf("%s must contain only unicode letters", err.Field())
	case "alphanumunicode":
		return fmt.Sprintf("%s must contain only unicode letters and numbers", err.Field())
	case "boolean":
		return fmt.Sprintf("%s must be a boolean value", err.Field())
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime", err.Field())
	case "file":
		return fmt.Sprintf("%s must be a valid file", err.Field())
	case "image":
		return fmt.Sprintf("%s must be a valid image", err.Field())
	case "mime":
		return fmt.Sprintf("%s must be a valid MIME type", err.Field())
	case "base64":
		return fmt.Sprintf("%s must be valid base64", err.Field())
	case "base64url":
		return fmt.Sprintf("%s must be valid base64url", err.Field())
	case "base64rawurl":
		return fmt.Sprintf("%s must be valid base64rawurl", err.Field())
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", err.Field())
	case "urn":
		return fmt.Sprintf("%s must be a valid URN", err.Field())
	case "hostname":
		return fmt.Sprintf("%s must be a valid hostname", err.Field())
	case "fqdn":
		return fmt.Sprintf("%s must be a valid FQDN", err.Field())
	case "tld":
		return fmt.Sprintf("%s must be a valid TLD", err.Field())
	case "datauri":
		return fmt.Sprintf("%s must be a valid data URI", err.Field())
	case "jwt":
		return fmt.Sprintf("%s must be a valid JWT", err.Field())
	case "mongodb":
		return fmt.Sprintf("%s must be a valid MongoDB ObjectID", err.Field())
	case "cron":
		return fmt.Sprintf("%s must be a valid cron expression", err.Field())
	case "timezone":
		return fmt.Sprintf("%s must be a valid timezone", err.Field())
	case "language":
		return fmt.Sprintf("%s must be a valid language code", err.Field())
	case "country":
		return fmt.Sprintf("%s must be a valid country code", err.Field())
	case "locale":
		return fmt.Sprintf("%s must be a valid locale", err.Field())
	case "bic":
		return fmt.Sprintf("%s must be a valid BIC", err.Field())
	case "iban":
		return fmt.Sprintf("%s must be a valid IBAN", err.Field())
	case "btc":
		return fmt.Sprintf("%s must be a valid Bitcoin address", err.Field())
	case "eth":
		return fmt.Sprintf("%s must be a valid Ethereum address", err.Field())
	case "btc_bech32":
		return fmt.Sprintf("%s must be a valid Bitcoin Bech32 address", err.Field())
	case "eth_checksum":
		return fmt.Sprintf("%s must be a valid Ethereum checksum address", err.Field())
	case "dive":
		return fmt.Sprintf("%s validation failed", err.Field())
	case "required_with":
		return fmt.Sprintf("%s is required when %s is present", err.Field(), err.Param())
	case "required_without":
		return fmt.Sprintf("%s is required when %s is not present", err.Field(), err.Param())
	case "required_with_all":
		return fmt.Sprintf("%s is required when all of %s are present", err.Field(), err.Param())
	case "required_without_all":
		return fmt.Sprintf("%s is required when none of %s are present", err.Field(), err.Param())
	case "excluded_with":
		return fmt.Sprintf("%s cannot be present when %s is present", err.Field(), err.Param())
	case "excluded_without":
		return fmt.Sprintf("%s cannot be present when %s is not present", err.Field(), err.Param())
	case "excluded_with_all":
		return fmt.Sprintf("%s cannot be present when all of %s are present", err.Field(), err.Param())
	case "excluded_without_all":
		return fmt.Sprintf("%s cannot be present when none of %s are present", err.Field(), err.Param())
	case "unique":
		return fmt.Sprintf("%s must be unique", err.Field())
	case "isdefault":
		return fmt.Sprintf("%s must be the default value", err.Field())
	case "eq":
		return fmt.Sprintf("%s must equal %s", err.Field(), err.Param())
	case "ne":
		return fmt.Sprintf("%s must not equal %s", err.Field(), err.Param())
	case "eqfield":
		return fmt.Sprintf("%s must equal %s", err.Field(), err.Param())
	case "nefield":
		return fmt.Sprintf("%s must not equal %s", err.Field(), err.Param())
	case "gtfield":
		return fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param())
	case "gtefield":
		return fmt.Sprintf("%s must be greater than or equal to %s", err.Field(), err.Param())
	case "ltfield":
		return fmt.Sprintf("%s must be less than %s", err.Field(), err.Param())
	case "ltefield":
		return fmt.Sprintf("%s must be less than or equal to %s", err.Field(), err.Param())
	case "eqcsfield":
		return fmt.Sprintf("%s must equal %s", err.Field(), err.Param())
	case "necsfield":
		return fmt.Sprintf("%s must not equal %s", err.Field(), err.Param())
	case "gtcsfield":
		return fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param())
	case "gtecsfield":
		return fmt.Sprintf("%s must be greater than or equal to %s", err.Field(), err.Param())
	case "ltcsfield":
		return fmt.Sprintf("%s must be less than %s", err.Field(), err.Param())
	case "ltecsfield":
		return fmt.Sprintf("%s must be less than or equal to %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("%s is not valid", err.Field())
	}
}
