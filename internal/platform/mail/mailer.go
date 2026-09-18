package mail

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
	"sync"
)

var (
	ErrRecipientRequired = errors.New("at least one recipient email address is required")
	ErrSubjectRequired   = errors.New("email subject cannot be empty")
	ErrMailSendFailed    = errors.New("failed to send email")
)

// EmailAttachment represents a file attached to an email message.
type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// EmailMessage represents an outbound email message.
type EmailMessage struct {
	From        string
	To          []string
	Subject     string
	BodyText    string
	BodyHTML    string
	Attachments []EmailAttachment
}

// Mailer defines the canonical interface for email dispatching (Section 56: email.send).
type Mailer interface {
	Send(ctx context.Context, msg EmailMessage) error
}

// MockMailer is a thread-safe in-memory mailer for testing and offline environments.
type MockMailer struct {
	mu           sync.RWMutex
	sentMessages []EmailMessage
}

var _ Mailer = (*MockMailer)(nil)

// NewMockMailer creates a new MockMailer instance.
func NewMockMailer() *MockMailer {
	return &MockMailer{
		sentMessages: make([]EmailMessage, 0),
	}
}

// Send records the email in memory.
func (m *MockMailer) Send(ctx context.Context, msg EmailMessage) error {
	if len(msg.To) == 0 {
		return ErrRecipientRequired
	}
	if strings.TrimSpace(msg.Subject) == "" {
		return ErrSubjectRequired
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = append(m.sentMessages, msg)
	return nil
}

// SentMessages returns a copy of all sent email messages.
func (m *MockMailer) SentMessages() []EmailMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copied := make([]EmailMessage, len(m.sentMessages))
	copy(copied, m.sentMessages)
	return copied
}

// LastMessage returns the most recently sent message, or nil if none.
func (m *MockMailer) LastMessage() *EmailMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.sentMessages) == 0 {
		return nil
	}
	last := m.sentMessages[len(m.sentMessages)-1]
	return &last
}

// Reset clears the recorded messages.
func (m *MockMailer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = make([]EmailMessage, 0)
}

// SMTPConfig holds configuration for production SMTP email dispatch.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// SMTPMailer implements Mailer using standard net/smtp.
type SMTPMailer struct {
	cfg SMTPConfig
}

var _ Mailer = (*SMTPMailer)(nil)

// NewSMTPMailer creates a new SMTPMailer instance.
func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

// Send sends an email message via SMTP.
func (s *SMTPMailer) Send(ctx context.Context, msg EmailMessage) error {
	if len(msg.To) == 0 {
		return ErrRecipientRequired
	}
	if strings.TrimSpace(msg.Subject) == "" {
		return ErrSubjectRequired
	}

	from := msg.From
	if from == "" {
		from = s.cfg.From
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	// Format RFC 822 / MIME message
	header := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		from, strings.Join(msg.To, ", "), msg.Subject)
	body := msg.BodyHTML
	if body == "" {
		body = msg.BodyText
	}

	rawMsg := []byte(header + body)
	err := smtp.SendMail(addr, auth, from, msg.To, rawMsg)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMailSendFailed, err)
	}
	return nil
}
