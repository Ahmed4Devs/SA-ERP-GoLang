package database_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func setupTestDB(t *testing.T) (*database.DB, *config.Config) {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL != "" {
		cfg.Database.URL = testDBURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	migrationsDir := "../../../database/migrations"
	if err := database.MigrateUp(ctx, cfg.Database.URL, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db, cfg
}

func TestDatabaseReadinessChecksSchemaVersion(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	if err := db.Check(); err != nil {
		t.Fatalf("database readiness check failed: %v", err)
	}
}

func textVal(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func timestamptzVal(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func uuidVal(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

func int4Val(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

func TestDatabaseConnectivityAndHealth(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	if err := db.Check(); err != nil {
		t.Fatalf("expected database check to pass, got: %v", err)
	}
}

func TestCurrencyConstraintsAndQueries(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	queries := sqlc.New(db.Pool)

	// 1. Positive Test: Insert valid currency (SAR)
	err := queries.UpsertCurrency(ctx, sqlc.UpsertCurrencyParams{
		Code:          "SAR",
		NameEn:        "Saudi Riyal",
		NameAr:        textVal("ريال سعودي"),
		DecimalPlaces: 2,
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("failed to upsert SAR currency: %v", err)
	}

	sar, err := queries.GetCurrency(ctx, "SAR")
	if err != nil {
		t.Fatalf("failed to get SAR currency: %v", err)
	}
	if strings.TrimSpace(sar.Code) != "SAR" || sar.DecimalPlaces != 2 {
		t.Errorf("unexpected currency values: %+v", sar)
	}

	// 2. Negative Test: ck_currency_code_upper (lowercase code rejected)
	err = queries.UpsertCurrency(ctx, sqlc.UpsertCurrencyParams{
		Code:          "usd", // lowercase should violate ck_currency_code_upper
		NameEn:        "US Dollar",
		DecimalPlaces: 2,
		IsActive:      true,
	})
	if err == nil {
		t.Fatal("expected error violating ck_currency_code_upper, got nil")
	}

	// 3. Negative Test: ck_currency_decimal_places (decimal_places > 6 rejected)
	err = queries.UpsertCurrency(ctx, sqlc.UpsertCurrencyParams{
		Code:          "BTC",
		NameEn:        "Bitcoin",
		DecimalPlaces: 8, // exceeds 6
		IsActive:      true,
	})
	if err == nil {
		t.Fatal("expected error violating ck_currency_decimal_places, got nil")
	}
}

func TestCountryConstraintsAndQueries(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	queries := sqlc.New(db.Pool)

	// 1. Positive Test: Insert valid country (SA)
	err := queries.UpsertCountry(ctx, sqlc.UpsertCountryParams{
		Code:        "SA",
		NameEn:      "Saudi Arabia",
		NameAr:      textVal("المملكة العربية السعودية"),
		PhonePrefix: textVal("+966"),
		IsActive:    true,
	})
	if err != nil {
		t.Fatalf("failed to upsert SA country: %v", err)
	}

	sa, err := queries.GetCountry(ctx, "SA")
	if err != nil {
		t.Fatalf("failed to get SA country: %v", err)
	}
	if strings.TrimSpace(sa.Code) != "SA" {
		t.Errorf("unexpected country code: %s", sa.Code)
	}

	// 2. Negative Test: ck_country_code_upper (lowercase code rejected)
	err = queries.UpsertCountry(ctx, sqlc.UpsertCountryParams{
		Code:     "ae", // lowercase
		NameEn:   "UAE",
		IsActive: true,
	})
	if err == nil {
		t.Fatal("expected error violating ck_country_code_upper, got nil")
	}
}

func TestOutboxEvents(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	// Pending events intentionally survive normal application restarts. Clear
	// this integration-test queue so LIMIT ordering is deterministic across
	// repeated local and CI runs against the same dedicated test database.
	if _, err := db.Pool.Exec(ctx, "TRUNCATE platform.outbox_events"); err != nil {
		t.Fatalf("failed to reset test outbox: %v", err)
	}
	queries := sqlc.New(db.Pool)

	payload := map[string]any{
		"event":  "OrganizationCreated",
		"org_id": uuid.NewString(),
	}
	payloadBytes, _ := json.Marshal(payload)

	aggID := uuid.New()
	now := time.Now().UTC()

	event, err := queries.InsertOutboxEvent(ctx, sqlc.InsertOutboxEventParams{
		EventType:     "org.created",
		AggregateType: "Organization",
		AggregateID:   uuidVal(aggID),
		Payload:       payloadBytes,
		OccurredAt:    timestamptzVal(now),
		AvailableAt:   timestamptzVal(now),
	})
	if err != nil {
		t.Fatalf("failed to insert outbox event: %v", err)
	}

	if event.AttemptCount != 0 {
		t.Errorf("expected default attempt_count 0, got %d", event.AttemptCount)
	}

	// Fetch pending
	pending, err := queries.FetchPendingOutboxEvents(ctx, 10)
	if err != nil {
		t.Fatalf("failed to fetch pending outbox events: %v", err)
	}

	found := false
	for _, p := range pending {
		if p.ID == event.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("inserted outbox event %v not found in pending list", event.ID)
	}

	// Mark published
	if err := queries.MarkOutboxEventPublished(ctx, event.ID); err != nil {
		t.Fatalf("failed to mark outbox event as published: %v", err)
	}

	// Verify no longer in pending list
	pendingAfter, err := queries.FetchPendingOutboxEvents(ctx, 10)
	if err != nil {
		t.Fatalf("failed to fetch pending events after publish: %v", err)
	}
	for _, p := range pendingAfter {
		if p.ID == event.ID {
			t.Errorf("published outbox event %v should not be returned in pending", event.ID)
		}
	}
}

func TestIdempotencyKeys(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	queries := sqlc.New(db.Pool)

	scope := "sales.invoices"
	key := "inv-" + uuid.NewString()
	hash := "sha256-test-hash"
	expiresAt := time.Now().Add(24 * time.Hour)

	// 1. Insert PROCESSING key
	record, err := queries.InsertIdempotencyKey(ctx, sqlc.InsertIdempotencyKeyParams{
		Scope:          scope,
		IdempotencyKey: key,
		RequestHash:    hash,
		ExpiresAt:      timestamptzVal(expiresAt),
	})
	if err != nil {
		t.Fatalf("failed to insert idempotency key: %v", err)
	}

	if record.State != "PROCESSING" {
		t.Errorf("expected state PROCESSING, got %s", record.State)
	}

	// 2. Negative Test: Duplicate (scope, idempotency_key) must violate unique constraint
	_, err = queries.InsertIdempotencyKey(ctx, sqlc.InsertIdempotencyKeyParams{
		Scope:          scope,
		IdempotencyKey: key,
		RequestHash:    hash,
		ExpiresAt:      timestamptzVal(expiresAt),
	})
	if err == nil {
		t.Fatal("expected duplicate idempotency key to fail, got nil")
	}

	// 3. Complete idempotency key
	respBody, _ := json.Marshal(map[string]string{"invoice_id": uuid.NewString()})
	err = queries.CompleteIdempotencyKey(ctx, sqlc.CompleteIdempotencyKeyParams{
		Scope:          scope,
		IdempotencyKey: key,
		ResponseStatus: int4Val(201),
		ResponseBody:   respBody,
	})
	if err != nil {
		t.Fatalf("failed to complete idempotency key: %v", err)
	}

	updated, err := queries.GetIdempotencyKey(ctx, sqlc.GetIdempotencyKeyParams{
		Scope:          scope,
		IdempotencyKey: key,
	})
	if err != nil {
		t.Fatalf("failed to fetch updated idempotency key: %v", err)
	}

	if updated.State != "COMPLETED" || updated.ResponseStatus.Int32 != 201 {
		t.Errorf("unexpected completed idempotency key state: %+v", updated)
	}
}

func TestTransactionManagerAtomicity(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	txManager := transaction.NewManager(db.Pool)

	testCurrencyCode := "TX1"
	_, _ = db.Pool.Exec(ctx, "DELETE FROM reference.currencies WHERE code IN ('TX1', 'TX2')")

	// 1. Test Rollback on Error
	expectedErr := errors.New("simulated domain error")
	err := txManager.WithinTx(ctx, func(txCtx context.Context) error {
		tx, ok := transaction.TxFromContext(txCtx)
		if !ok {
			return errors.New("expected tx in context")
		}

		queries := sqlc.New(tx)
		if err := queries.UpsertCurrency(txCtx, sqlc.UpsertCurrencyParams{
			Code:          testCurrencyCode,
			NameEn:        "Tx Test Currency",
			DecimalPlaces: 2,
			IsActive:      true,
		}); err != nil {
			return err
		}

		// Return error to trigger rollback
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	// Verify record was rolled back
	_, err = sqlc.New(db.Pool).GetCurrency(ctx, testCurrencyCode)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected ErrNoRows for rolled-back currency, got %v", err)
	}

	// 2. Test Commit on Success
	testCurrencyCode2 := "TX2"
	err = txManager.WithinTx(ctx, func(txCtx context.Context) error {
		queries := sqlc.New(db.DBTX(txCtx))
		return queries.UpsertCurrency(txCtx, sqlc.UpsertCurrencyParams{
			Code:          testCurrencyCode2,
			NameEn:        "Tx Commit Currency",
			DecimalPlaces: 2,
			IsActive:      true,
		})
	})
	if err != nil {
		t.Fatalf("expected commit to succeed, got %v", err)
	}

	// Verify record was committed
	cur, err := sqlc.New(db.Pool).GetCurrency(ctx, testCurrencyCode2)
	if err != nil {
		t.Fatalf("failed to fetch committed currency: %v", err)
	}
	if strings.TrimSpace(cur.Code) != testCurrencyCode2 {
		t.Errorf("expected currency code %s, got %s", testCurrencyCode2, cur.Code)
	}
}
