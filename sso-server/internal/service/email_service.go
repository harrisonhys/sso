package service

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"path/filepath"
)

// EmailConfig holds email service configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// EmailService handles sending emails
type EmailService struct {
	config    *EmailConfig
	templates map[string]*template.Template
}

// NewEmailService creates a new email service
func NewEmailService(config *EmailConfig) (*EmailService, error) {
	service := &EmailService{
		config:    config,
		templates: make(map[string]*template.Template),
	}

	// Load email templates
	if err := service.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	return service, nil
}

// loadTemplates loads all email templates
func (s *EmailService) loadTemplates() error {
	templateDir := "templates/emails"

	templateFiles := map[string]string{
		"password_reset": filepath.Join(templateDir, "password_reset.html"),
		"welcome":        filepath.Join(templateDir, "welcome.html"),
	}

	for name, path := range templateFiles {
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		s.templates[name] = tmpl
	}

	return nil
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(to, name, resetToken string) error {
	// Build reset URL
	resetURL := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", resetToken)

	data := struct {
		Name     string
		ResetURL string
	}{
		Name:     name,
		ResetURL: resetURL,
	}

	body, err := s.renderTemplate("password_reset", data)
	if err != nil {
		return fmt.Errorf("failed to render password reset template: %w", err)
	}

	subject := "Password Reset Request - SSO Server"
	return s.sendEmail(to, subject, body)
}

// SendWelcomeEmail sends a welcome email to new users
func (s *EmailService) SendWelcomeEmail(to, name string) error {
	data := struct {
		Name string
	}{
		Name: name,
	}

	body, err := s.renderTemplate("welcome", data)
	if err != nil {
		return fmt.Errorf("failed to render welcome template: %w", err)
	}

	subject := "Welcome to SSO Server!"
	return s.sendEmail(to, subject, body)
}

// renderTemplate renders an email template with data
func (s *EmailService) renderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, exists := s.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// sendEmail sends an email via SMTP
func (s *EmailService) sendEmail(to, subject, htmlBody string) error {
	// Build email message
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	// SMTP authentication
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)
	err := smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendEmail sends a generic email (for testing or custom emails)
func (s *EmailService) SendEmail(to, subject, body string) error {
	return s.sendEmail(to, subject, body)
}
