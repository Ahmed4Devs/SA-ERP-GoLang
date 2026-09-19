package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"erp/internal/notifications/domain"
	"erp/internal/notifications/ports"
	"erp/internal/platform/mail"
	"erp/internal/platform/outbox"
)

// EmailDelivery turns a durable notification outbox event into an SMTP call.
type EmailDelivery struct {
	repo   ports.Repository
	mailer mail.Mailer
}

func NewEmailDelivery(repo ports.Repository, mailer mail.Mailer) *EmailDelivery {
	return &EmailDelivery{repo: repo, mailer: mailer}
}

func (d *EmailDelivery) Handle(ctx context.Context, event outbox.Event) error {
	var payload EmailRequestedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode email notification event: %w", err)
	}
	if payload.LegalEntityID == [16]byte{} || payload.NotificationID == [16]byte{} {
		return errors.New("email notification event is missing identifiers")
	}

	notification, err := d.repo.GetByID(ctx, payload.LegalEntityID, payload.NotificationID)
	if err != nil {
		return err
	}
	if notification.Status == domain.StatusSent || notification.Status == domain.StatusRead {
		return nil
	}
	if notification.Channel != domain.ChannelEmail {
		return fmt.Errorf("notification %s is not an email", notification.ID)
	}
	if strings.TrimSpace(notification.RecipientEmail) == "" {
		return fmt.Errorf("notification %s has no recipient email", notification.ID)
	}

	message := mail.EmailMessage{
		To:       []string{notification.RecipientEmail},
		Subject:  notification.TitleEn,
		BodyText: fmt.Sprintf("%s\n\n%s", notification.BodyAr, notification.BodyEn),
		BodyHTML: fmt.Sprintf(`<div dir="rtl" style="font-family:sans-serif;padding:20px;line-height:1.6"><h2 style="color:#1e3a8a">%s</h2><p style="font-size:16px">%s</p><hr style="margin:20px 0;border:0;border-top:1px solid #e5e7eb"><div dir="ltr" style="text-align:left"><h3 style="color:#4b5563">%s</h3><p style="font-size:14px;color:#4b5563">%s</p></div></div>`,
			html.EscapeString(notification.TitleAr),
			html.EscapeString(notification.BodyAr),
			html.EscapeString(notification.TitleEn),
			html.EscapeString(notification.BodyEn),
		),
	}

	if err := d.mailer.Send(ctx, message); err != nil {
		errorMessage := err.Error()
		statusErr := d.repo.UpdateStatus(ctx, notification.LegalEntityID, notification.ID, domain.StatusFailed, nil, &errorMessage)
		return errors.Join(err, statusErr)
	}

	now := time.Now().UTC()
	if err := d.repo.UpdateStatus(ctx, notification.LegalEntityID, notification.ID, domain.StatusSent, &now, nil); err != nil {
		return fmt.Errorf("mark email notification sent: %w", err)
	}
	return nil
}
