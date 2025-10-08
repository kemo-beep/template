package services

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
}

func NewEmailService() *EmailService {
	return &EmailService{
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    os.Getenv("FROM_EMAIL"),
	}
}

func (e *EmailService) SendEmail(to, subject, body string) error {
	// Create message
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)

	// Set up authentication
	auth := smtp.PlainAuth("", e.smtpUsername, e.smtpPassword, e.smtpHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)
	err := smtp.SendMail(addr, auth, e.fromEmail, []string{to}, []byte(message))

	return err
}

func (e *EmailService) SendWelcomeEmail(to, name string) error {
	subject := "Welcome to Mobile Backend!"
	body := fmt.Sprintf(`
Hello %s,

Welcome to our mobile backend service! We're excited to have you on board.

Best regards,
The Mobile Backend Team
`, name)

	return e.SendEmail(to, subject, body)
}

func (e *EmailService) SendPasswordResetEmail(to, resetToken string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
You have requested to reset your password.

Please click the following link to reset your password:
%s/reset-password?token=%s

This link will expire in 1 hour.

If you did not request this password reset, please ignore this email.

Best regards,
The Mobile Backend Team
`, os.Getenv("FRONTEND_URL"), resetToken)

	return e.SendEmail(to, subject, body)
}
