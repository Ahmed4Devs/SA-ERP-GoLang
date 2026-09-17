package mail_test

import (
	"context"
	"testing"

	"erp/internal/platform/mail"
)

func TestMockMailer(t *testing.T) {
	mailer := mail.NewMockMailer()
	ctx := context.Background()

	// 1. Send valid email
	msg := mail.EmailMessage{
		To:       []string{"finance@erp.sa", "auditor@erp.sa"},
		Subject:  "ZATCA Phase 2 E-Invoice Cleared",
		BodyHTML: "<h1>Invoice Cleared</h1><p>Invoice #INV-2026-001 has been validated.</p>",
		BodyText: "Invoice #INV-2026-001 has been validated.",
	}

	err := mailer.Send(ctx, msg)
	if err != nil {
		t.Fatalf("failed to send mock email: %v", err)
	}

	if len(mailer.SentMessages()) != 1 {
		t.Fatalf("expected 1 sent message, got %d", len(mailer.SentMessages()))
	}

	last := mailer.LastMessage()
	if last == nil || last.Subject != msg.Subject {
		t.Errorf("last message mismatch")
	}

	// 2. Validation errors
	err = mailer.Send(ctx, mail.EmailMessage{
		To:      []string{},
		Subject: "Test",
	})
	if err != mail.ErrRecipientRequired {
		t.Errorf("expected ErrRecipientRequired, got: %v", err)
	}

	err = mailer.Send(ctx, mail.EmailMessage{
		To:      []string{"test@erp.sa"},
		Subject: "   ",
	})
	if err != mail.ErrSubjectRequired {
		t.Errorf("expected ErrSubjectRequired, got: %v", err)
	}

	// 3. Reset
	mailer.Reset()
	if len(mailer.SentMessages()) != 0 {
		t.Errorf("expected 0 messages after reset, got %d", len(mailer.SentMessages()))
	}
}
