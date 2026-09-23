package notifications_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"erp/internal/iam"
	"erp/internal/notifications"
	notificationapp "erp/internal/notifications/application"
	"erp/internal/notifications/domain"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/mail"
	"erp/internal/platform/outbox"
	"erp/internal/platform/transaction"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type failingOutbox struct{}

func (failingOutbox) Enqueue(context.Context, string, string, *uuid.UUID, any) (*outbox.Event, error) {
	return nil, fmt.Errorf("forced outbox failure")
}

func setupNotificationsIntegrationTest(t *testing.T) (*notifications.Module, *database.DB, *mail.MockMailer, uuid.UUID, uuid.UUID, context.Context) {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("integration database setup failed: %v", err)
	}
	t.Cleanup(db.Close)

	txManager := transaction.NewManager(db.Pool)
	if err := database.MigrateUp(ctx, cfg.Database.URL, "../../../database/migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)

	// 1. Seed Organization
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// 2. Admin User
	admin, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var adminID uuid.UUID
	if err == nil && admin != nil {
		adminID = admin.ID
	} else {
		newAdmin, err := iamMod.AuthService.SeedAdmin(ctx, "notif_admin", "notif_admin@erp.sa", "مدير الإشعارات", "Admin@ERP123!")
		if err == nil && newAdmin != nil {
			adminID = newAdmin.ID
		}
	}

	// 3. Mock Mailer
	mockMailer := mail.NewMockMailer()
	mod := notifications.NewModule(cfg, db, mockMailer)

	return mod, db, mockMailer, le.ID, adminID, ctx
}

func TestNotifications_InAppDispatchAndUnreadCount(t *testing.T) {
	mod, _, _, leID, userID, ctx := setupNotificationsIntegrationTest(t)

	initialUnread, _ := mod.Service.GetUnreadCount(ctx, leID, &userID)

	// 1. Dispatch in-app notification
	params := domain.SendNotificationParams{
		LegalEntityID:   leID,
		RecipientUserID: &userID,
		Channel:         domain.ChannelInApp,
		TitleAr:         "طلب اعتماد فاتورة مشتريات",
		TitleEn:         "Purchase Invoice Approval Request",
		BodyAr:          "تتطلب فاتورة المشتريات رقم PO-2026-004 اعتمادكم المالي.",
		BodyEn:          "Purchase Invoice #PO-2026-004 requires your financial approval.",
		Category:        domain.CategoryWorkflow,
		Priority:        domain.PriorityHigh,
		ActionURL:       "/purchasing/invoices/PO-2026-004",
	}

	notif, err := mod.Service.Send(ctx, params)
	if err != nil {
		t.Fatalf("failed to send in-app notification: %v", err)
	}

	if notif.ID == uuid.Nil {
		t.Fatal("expected non-nil notification ID")
	}
	if notif.Status != domain.StatusSent {
		t.Errorf("expected status SENT, got %s", notif.Status)
	}

	// 2. Verify unread count increased
	newUnread, err := mod.Service.GetUnreadCount(ctx, leID, &userID)
	if err != nil {
		t.Fatalf("failed to get unread count: %v", err)
	}
	if newUnread != initialUnread+1 {
		t.Errorf("expected unread count %d, got %d", initialUnread+1, newUnread)
	}

	// 3. Get notification details
	fetched, err := mod.Service.Get(ctx, leID, notif.ID)
	if err != nil {
		t.Fatalf("failed to get notification: %v", err)
	}
	if fetched.TitleAr != params.TitleAr {
		t.Errorf("expected title_ar %q, got %q", params.TitleAr, fetched.TitleAr)
	}
	if fetched.Priority != domain.PriorityHigh {
		t.Errorf("expected priority HIGH, got %s", fetched.Priority)
	}

	// 4. Mark as read
	if err := mod.Service.MarkAsRead(ctx, leID, notif.ID); err != nil {
		t.Fatalf("failed to mark notification as read: %v", err)
	}

	afterReadCount, _ := mod.Service.GetUnreadCount(ctx, leID, &userID)
	if afterReadCount != initialUnread {
		t.Errorf("expected unread count to return to %d, got %d", initialUnread, afterReadCount)
	}
}

func TestNotifications_EmailDispatchViaMailer(t *testing.T) {
	mod, _, mockMailer, leID, userID, ctx := setupNotificationsIntegrationTest(t)

	recipientEmail := "cfo.finance@saudi-enterprise.sa"
	params := domain.SendNotificationParams{
		LegalEntityID:   leID,
		RecipientUserID: &userID,
		RecipientEmail:  recipientEmail,
		Channel:         domain.ChannelEmail,
		TitleAr:         "تنبيه تسوية ضريبة القيمة المضافة ZATCA",
		TitleEn:         "ZATCA VAT Return Settlement Notice",
		BodyAr:          "تم إغلاق إقرار الربع الأول بنجاح وسداد الرصيد المستحق.",
		BodyEn:          "Q1 VAT Return has been reconciled and settled with ZATCA portal.",
		Category:        domain.CategoryZATCA,
		Priority:        domain.PriorityUrgent,
	}

	notif, err := mod.Service.Send(ctx, params)
	if err != nil {
		t.Fatalf("failed to send email notification: %v", err)
	}

	if notif.Status != domain.StatusPending {
		t.Errorf("expected status PENDING before worker delivery, got %s", notif.Status)
	}
	if len(mockMailer.SentMessages()) != 0 {
		t.Fatal("email must not be sent before the outbox worker processes it")
	}

	dispatcher := outbox.NewDispatcher(mod.Outbox, nil, outbox.DefaultDispatcherConfig())
	if err := mod.RegisterOutboxHandlers(dispatcher); err != nil {
		t.Fatalf("failed to register notification outbox handler: %v", err)
	}
	processed, err := dispatcher.RunOnce(ctx)
	if err != nil {
		t.Fatalf("outbox delivery failed: %v", err)
	}
	if processed < 1 {
		t.Fatalf("expected at least one delivered outbox event, got %d", processed)
	}

	delivered, err := mod.Service.Get(ctx, leID, notif.ID)
	if err != nil {
		t.Fatalf("failed to reload delivered notification: %v", err)
	}
	if delivered.Status != domain.StatusSent || delivered.SentAt == nil {
		t.Fatalf("expected delivered notification status SENT, got %s", delivered.Status)
	}

	// Verify email delivered to MockMailer
	sentEmails := mockMailer.SentMessages()
	if len(sentEmails) == 0 {
		t.Fatal("expected mock mailer to have received at least 1 email")
	}

	last := mockMailer.LastMessage()
	if last == nil {
		t.Fatal("expected non-nil last email")
	}
	if len(last.To) == 0 || last.To[0] != recipientEmail {
		t.Errorf("expected recipient %s, got %+v", recipientEmail, last.To)
	}
	if last.Subject != params.TitleEn {
		t.Errorf("expected subject %q, got %q", params.TitleEn, last.Subject)
	}
}

func TestNotifications_EmailAndOutboxAreAtomic(t *testing.T) {
	mod, db, _, leID, userID, ctx := setupNotificationsIntegrationTest(t)
	service := notificationapp.NewNotificationService(mod.Repo, failingOutbox{}, transaction.NewManager(db.Pool))

	_, before, err := mod.Service.List(ctx, domain.NotificationFilter{LegalEntityID: leID, RecipientUserID: &userID})
	if err != nil {
		t.Fatalf("failed to count notifications before rollback test: %v", err)
	}
	_, err = service.Send(ctx, domain.SendNotificationParams{
		LegalEntityID:   leID,
		RecipientUserID: &userID,
		RecipientEmail:  "atomic@example.sa",
		Channel:         domain.ChannelEmail,
		TitleAr:         "اختبار الذرية",
		TitleEn:         "Atomic delivery test",
		BodyAr:          "يجب التراجع عن السجل",
		BodyEn:          "The record must roll back",
	})
	if err == nil {
		t.Fatal("expected forced outbox failure")
	}
	_, after, err := mod.Service.List(ctx, domain.NotificationFilter{LegalEntityID: leID, RecipientUserID: &userID})
	if err != nil {
		t.Fatalf("failed to count notifications after rollback test: %v", err)
	}
	if after != before {
		t.Fatalf("notification insert was not rolled back: before=%d after=%d", before, after)
	}
}

func TestNotifications_MarkAllAsRead(t *testing.T) {
	mod, _, _, leID, userID, ctx := setupNotificationsIntegrationTest(t)

	// Send 3 notifications
	for i := 1; i <= 3; i++ {
		_, err := mod.Service.Send(ctx, domain.SendNotificationParams{
			LegalEntityID:   leID,
			RecipientUserID: &userID,
			Channel:         domain.ChannelInApp,
			TitleAr:         fmt.Sprintf("إشعار مالي رقم %d", i),
			TitleEn:         fmt.Sprintf("Financial Alert #%d", i),
			BodyAr:          "تفاصيل العملية المالية",
			BodyEn:          "Financial transaction details",
			Category:        domain.CategoryFinance,
		})
		if err != nil {
			t.Fatalf("failed to send notification %d: %v", i, err)
		}
	}

	unreadBefore, err := mod.Service.GetUnreadCount(ctx, leID, &userID)
	if err != nil || unreadBefore < 3 {
		t.Fatalf("expected unread count >= 3, got %d (err: %v)", unreadBefore, err)
	}

	// Mark all as read
	if err := mod.Service.MarkAllAsRead(ctx, leID, &userID); err != nil {
		t.Fatalf("failed to mark all as read: %v", err)
	}

	unreadAfter, err := mod.Service.GetUnreadCount(ctx, leID, &userID)
	if err != nil || unreadAfter != 0 {
		t.Errorf("expected unread count 0 after mark-all-read, got %d", unreadAfter)
	}
}

func TestNotifications_FilteringAndSummary(t *testing.T) {
	mod, _, _, leID, userID, ctx := setupNotificationsIntegrationTest(t)

	// Send inventory alert
	_, err := mod.Service.Send(ctx, domain.SendNotificationParams{
		LegalEntityID:   leID,
		RecipientUserID: &userID,
		Channel:         domain.ChannelInApp,
		TitleAr:         "تحذير نفاد مخزون مادة أولية",
		TitleEn:         "Low Stock Inventory Warning",
		BodyAr:          "مستوى مخزون الصنف وصل إلى حد إعادة الطلب.",
		BodyEn:          "Stock balance has dropped below reorder threshold.",
		Category:        domain.CategoryInventory,
	})
	if err != nil {
		t.Fatalf("failed to send inventory notification: %v", err)
	}

	// Filter by Category
	catInv := domain.CategoryInventory
	list, total, err := mod.Service.List(ctx, domain.NotificationFilter{
		LegalEntityID:   leID,
		RecipientUserID: &userID,
		Category:        &catInv,
	})
	if err != nil {
		t.Fatalf("failed to list by category: %v", err)
	}
	if total < 1 || len(list) < 1 {
		t.Fatalf("expected at least 1 inventory notification, got %d", total)
	}

	// Check Summary analytics
	summary, err := mod.Service.GetSummary(ctx, leID, &userID)
	if err != nil {
		t.Fatalf("failed to get summary: %v", err)
	}
	if summary.TotalCount < 1 {
		t.Errorf("expected total count >= 1, got %d", summary.TotalCount)
	}
	if summary.CategoriesCount["INVENTORY"] < 1 {
		t.Errorf("expected INVENTORY category count >= 1 in summary")
	}
}

func TestNotifications_SeedDefaultData(t *testing.T) {
	mod, db, _, _, _, ctx := setupNotificationsIntegrationTest(t)

	// Fresh legal entity
	var freshLEID uuid.UUID
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO organization.legal_entities (
			id, organization_id, code, name_en, name_ar, cr_number, 
			base_currency, is_active
		) VALUES (
			uuidv7(), (SELECT organization_id FROM organization.legal_entities LIMIT 1),
			'NOTIF-SEED-' || substr(uuidv7()::text, 1, 8), 
			'Seed Notification Corp', 'شركة التنبيهات التجريبية', 
			'1010' || substr((random()*1000000)::int::text, 1, 6),
			'SAR', true
		) RETURNING id
	`).Scan(&freshLEID)
	if err != nil {
		t.Fatalf("failed to create fresh legal entity for seed test: %v", err)
	}

	// Run Seed procedure
	if err := mod.SeedDefaultNotificationsData(ctx, freshLEID); err != nil {
		t.Fatalf("failed to seed default notifications: %v", err)
	}

	// Verify standard template exists
	tpl, err := mod.Repo.GetTemplateByCode(ctx, "ZATCA_INVOICE_CLEARED")
	if err != nil {
		t.Fatalf("failed to get seeded template: %v", err)
	}
	if tpl.Category != domain.CategoryZATCA {
		t.Errorf("expected template category ZATCA, got %s", tpl.Category)
	}

	// Verify initial welcome alert seeded for fresh legal entity
	list, total, err := mod.Service.List(ctx, domain.NotificationFilter{
		LegalEntityID: freshLEID,
	})
	if err != nil {
		t.Fatalf("failed to list seeded notifications: %v", err)
	}
	if total < 1 || len(list) < 1 {
		t.Fatalf("expected at least 1 welcome notification seeded, got %d", total)
	}
}

func TestNotifications_HTTPTransport(t *testing.T) {
	mod, _, _, leID, userID, _ := setupNotificationsIntegrationTest(t)

	r := chi.NewRouter()
	mod.RegisterRoutes(r, nil)

	// 1. POST /api/v1/notifications/send
	sendPayload := domain.SendNotificationParams{
		LegalEntityID:   leID,
		RecipientUserID: &userID,
		Channel:         domain.ChannelInApp,
		TitleAr:         "تنبيه اختبار المنظومة",
		TitleEn:         "HTTP Transport Alert",
		BodyAr:          "رسالة تجريبية عبر واجهة برمجة التطبيقات",
		BodyEn:          "Test notification message via REST API",
		Category:        domain.CategorySystem,
		Priority:        domain.PriorityNormal,
	}
	payloadBytes, _ := json.Marshal(sendPayload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/send", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected HTTP 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var created domain.Notification
	_ = json.NewDecoder(w.Body).Decode(&created)
	if created.ID == uuid.Nil {
		t.Fatal("expected valid notification ID from HTTP response")
	}

	// 2. GET /api/v1/notifications/unread-count
	reqCount := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/notifications/unread-count?legal_entity_id=%s&recipient_user_id=%s", leID, userID), nil)
	wCount := httptest.NewRecorder()
	r.ServeHTTP(wCount, reqCount)

	if wCount.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK on unread-count, got %d: %s", wCount.Code, wCount.Body.String())
	}

	// 3. PATCH /api/v1/notifications/{id}/read
	reqRead := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/notifications/%s/read?legal_entity_id=%s", created.ID, leID), nil)
	wRead := httptest.NewRecorder()
	r.ServeHTTP(wRead, reqRead)

	if wRead.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK on mark-as-read, got %d: %s", wRead.Code, wRead.Body.String())
	}

	// 4. POST /api/v1/notifications/mark-all-read
	reqMarkAll := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/notifications/mark-all-read?legal_entity_id=%s&recipient_user_id=%s", leID, userID), nil)
	wMarkAll := httptest.NewRecorder()
	r.ServeHTTP(wMarkAll, reqMarkAll)

	if wMarkAll.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK on mark-all-read, got %d: %s", wMarkAll.Code, wMarkAll.Body.String())
	}
}
