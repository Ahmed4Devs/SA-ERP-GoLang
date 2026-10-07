package posting

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"erp/internal/accounting/domain"
	"erp/internal/accounting/ports"
	auditports "erp/internal/audit/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	workflowservices "erp/internal/workflow/application/services"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// SourceService composes source-document mutations with the common posting
// engine. Its transaction acquires the audit lock before any business row lock.
type SourceService struct {
	db     *database.DB
	engine ports.PostingEngine
	audit  ports.PostingAuditRecorder
	tx     transaction.Manager
}

// AccountForDate resolves an effective posting rule before the legacy account
// code. Explicit document mappings take precedence. Configured but inactive or
// not-yet-effective rules fail closed rather than silently using a fallback.
func (s *SourceService) AccountForDate(ctx context.Context, entity uuid.UUID, selected *uuid.UUID, module, role, fallbackCode string, date time.Time) (uuid.UUID, error) {
	if selected != nil && *selected != uuid.Nil {
		return s.Account(ctx, entity, selected, "")
	}
	var ruleID uuid.UUID
	var active bool
	err := s.db.DBTX(ctx).QueryRow(ctx, `SELECT id,is_active FROM accounting.posting_rules WHERE legal_entity_id=$1 AND source_module=$2 AND code=$3 FOR SHARE`, entity, module, role).Scan(&ruleID, &active)
	if err == pgx.ErrNoRows {
		return s.Account(ctx, entity, nil, fallbackCode)
	}
	if err != nil {
		return uuid.Nil, err
	}
	if !active || date.IsZero() {
		return uuid.Nil, fmt.Errorf("posting rule %s/%s is inactive or accounting date is missing", module, role)
	}
	var account uuid.UUID
	err = s.db.DBTX(ctx).QueryRow(ctx, `SELECT account_id FROM accounting.posting_rule_versions WHERE posting_rule_id=$1 AND effective_from <= $2 ORDER BY effective_from DESC,version_number DESC LIMIT 1`, ruleID, date).Scan(&account)
	if err != nil {
		return uuid.Nil, fmt.Errorf("posting rule %s/%s has no effective version: %w", module, role, err)
	}
	return s.Account(ctx, entity, &account, "")
}

// CheckApprovalLimit uses the persisted document amount and currency. A foreign
// amount cannot be compared to a base-currency authority without its saved FX rate.
func (s *SourceService) CheckApprovalLimit(ctx context.Context, entity, actor uuid.UUID, resourceType, currency string, rate, amount decimal.Decimal) error {
	var base string
	if err := s.db.DBTX(ctx).QueryRow(ctx, "SELECT base_currency FROM organization.legal_entities WHERE id=$1", entity).Scan(&base); err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(base), strings.TrimSpace(currency)) {
		rate = decimal.NewFromInt(1)
	}
	if !rate.IsPositive() {
		return fmt.Errorf("approval requires a saved foreign exchange rate")
	}
	return workflowservices.CheckApprovalLimit(ctx, s.db, entity, actor, resourceType, amount.Mul(rate).Round(6))
}

// RecordSource records non-journal source transitions in the same transaction,
// including inventory counts whose net financial variance is zero.
func (s *SourceService) RecordSource(ctx context.Context, entity, actor, id uuid.UUID, action string, snapshot any) error {
	if _, ok := transaction.TxFromContext(ctx); !ok || actor == uuid.Nil {
		return fmt.Errorf("source audit requires a transaction and actor")
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	_, err = s.audit.RecordEvent(ctx, auditports.RecordEventCommand{UserID: &actor, LegalEntityID: &entity, Action: action, ResourceType: action, ResourceID: id.String(), Payload: payload})
	return err
}

func NewSourceService(db *database.DB, engine ports.PostingEngine, audit ports.PostingAuditRecorder) *SourceService {
	return &SourceService{db: db, engine: engine, audit: audit, tx: transaction.NewManager(db.Pool)}
}

func (s *SourceService) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.audit.LockChain(ctx); err != nil {
			return err
		}
		return fn(ctx)
	})
}

func (s *SourceService) Account(ctx context.Context, entity uuid.UUID, selected *uuid.UUID, fallbackCode string) (uuid.UUID, error) {
	var id uuid.UUID
	if selected != nil && *selected != uuid.Nil {
		err := s.db.DBTX(ctx).QueryRow(ctx, `SELECT id FROM accounting.accounts WHERE id=$1 AND legal_entity_id=$2 AND is_active AND allow_posting FOR SHARE`, *selected, entity).Scan(&id)
		if err != nil {
			return id, fmt.Errorf("configured posting account is unavailable: %w", err)
		}
		return id, nil
	}
	err := s.db.DBTX(ctx).QueryRow(ctx, `SELECT id FROM accounting.accounts WHERE legal_entity_id=$1 AND code=$2 AND is_active AND allow_posting FOR SHARE`, entity, fallbackCode).Scan(&id)
	if err != nil {
		return id, fmt.Errorf("posting account %s must be configured: %w", fallbackCode, err)
	}
	return id, nil
}

func (s *SourceService) Post(ctx context.Context, journalType string, request domain.PostingRequest) (uuid.UUID, error) {
	if _, ok := transaction.TxFromContext(ctx); !ok {
		return uuid.Nil, fmt.Errorf("source posting requires an enclosing transaction")
	}
	if request.ActorID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("posting actor is required")
	}
	var baseCurrency string
	if err := s.db.DBTX(ctx).QueryRow(ctx, `SELECT base_currency FROM organization.legal_entities WHERE id=$1`, request.LegalEntityID).Scan(&baseCurrency); err != nil {
		return uuid.Nil, err
	}
	request.CurrencyCode = strings.ToUpper(strings.TrimSpace(request.CurrencyCode))
	if request.CurrencyCode == "" {
		request.CurrencyCode = strings.TrimSpace(baseCurrency)
	}
	if request.ExchangeRate.IsZero() {
		if request.CurrencyCode != strings.TrimSpace(baseCurrency) {
			return uuid.Nil, fmt.Errorf("foreign-currency posting requires an explicit exchange rate")
		}
		request.ExchangeRate = decimal.NewFromInt(1)
	}
	if !request.ExchangeRate.IsPositive() {
		return uuid.Nil, fmt.Errorf("exchange rate must be positive")
	}
	if request.JournalID == uuid.Nil {
		if err := s.db.DBTX(ctx).QueryRow(ctx, `SELECT id FROM accounting.journals WHERE legal_entity_id=$1 AND journal_type=$2 AND is_active ORDER BY code,id LIMIT 1 FOR SHARE`, request.LegalEntityID, journalType).Scan(&request.JournalID); err != nil {
			return uuid.Nil, fmt.Errorf("active %s journal must be configured: %w", journalType, err)
		}
	}
	// Allocate rounding by cumulative totals on each side, preserving balanced
	// base totals for multi-line foreign-currency documents.
	var debit, credit, previousDebit, previousCredit decimal.Decimal
	for i := range request.Lines {
		line := &request.Lines[i]
		debit = debit.Add(line.TransactionDebit)
		credit = credit.Add(line.TransactionCredit)
		nextDebit, nextCredit := debit.Mul(request.ExchangeRate).Round(6), credit.Mul(request.ExchangeRate).Round(6)
		line.BaseDebit, line.BaseCredit = nextDebit.Sub(previousDebit), nextCredit.Sub(previousCredit)
		previousDebit, previousCredit = nextDebit, nextCredit
	}
	result, err := s.engine.Post(ctx, request)
	return result.JournalEntryID, err
}
