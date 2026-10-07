package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"erp/internal/accounting/application/posting"
	"erp/internal/accounting/domain"
	"erp/internal/platform/database"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ReferenceService manages audited, company-scoped accounting configuration.
// Rates and posting mappings are append-only effective versions.
type ReferenceService struct {
	db     *database.DB
	ledger *posting.SourceService
}

func NewReferenceService(db *database.DB, ledger *posting.SourceService) *ReferenceService {
	return &ReferenceService{db: db, ledger: ledger}
}

type DimensionCommand struct {
	DimensionID uuid.UUID `json:"dimension_id"`
	Code        string    `json:"code"`
	NameEn      string    `json:"name_en"`
	NameAr      string    `json:"name_ar"`
}

type ExchangeRateCommand struct {
	FromCurrency  string          `json:"from_currency"`
	ToCurrency    string          `json:"to_currency"`
	EffectiveDate string          `json:"effective_date"`
	Rate          decimal.Decimal `json:"rate"`
	Source        string          `json:"source"`
}

type PostingRuleCommand struct {
	SourceModule  string     `json:"source_module"`
	Code          string     `json:"code"`
	NameEn        string     `json:"name_en"`
	NameAr        string     `json:"name_ar"`
	EffectiveFrom string     `json:"effective_from"`
	AccountID     uuid.UUID  `json:"account_id"`
	JournalID     *uuid.UUID `json:"journal_id"`
}

type OpeningBalanceCommand struct {
	FiscalPeriodID uuid.UUID       `json:"fiscal_period_id"`
	AccountID      uuid.UUID       `json:"account_id"`
	CurrencyCode   string          `json:"currency_code"`
	Debit          decimal.Decimal `json:"debit"`
	Credit         decimal.Decimal `json:"credit"`
}

type ReconciliationCommand struct {
	AccountID          uuid.UUID   `json:"account_id"`
	Reference          string      `json:"reference"`
	ReconciliationDate string      `json:"reconciliation_date"`
	CurrencyCode       string      `json:"currency_code"`
	JournalLineIDs     []uuid.UUID `json:"journal_line_ids"`
}

// List uses only fixed SQL definitions; user input never becomes an identifier.
func (s *ReferenceService) List(ctx context.Context, entity uuid.UUID, resource string, limit, offset int) (json.RawMessage, int64, error) {
	queries := map[string]string{
		"dimensions":       `SELECT * FROM accounting.dimensions WHERE legal_entity_id=$1`,
		"dimension-values": `SELECT * FROM accounting.dimension_values WHERE legal_entity_id=$1`,
		"exchange-rates":   `SELECT id,legal_entity_id,from_currency,to_currency,effective_date,rate::text,version_number,source,created_by,created_at FROM accounting.exchange_rates WHERE legal_entity_id=$1`,
		"posting-rules":    `SELECT p.*,v.id AS version_id,v.version_number,v.effective_from,v.account_id,v.journal_id FROM accounting.posting_rules p LEFT JOIN accounting.posting_rule_versions v ON v.posting_rule_id=p.id WHERE p.legal_entity_id=$1`,
		"opening-balances": `SELECT id,legal_entity_id,fiscal_period_id,account_id,currency_code,debit::text,credit::text,journal_entry_id,status,created_by,created_at FROM accounting.opening_balances WHERE legal_entity_id=$1`,
		"reconciliations":  `SELECT r.*,(SELECT coalesce(jsonb_agg(l.journal_line_id),'[]'::jsonb) FROM accounting.reconciliation_lines l WHERE l.reconciliation_id=r.id) AS journal_line_ids FROM accounting.reconciliations r WHERE legal_entity_id=$1`,
	}
	query, ok := queries[resource]
	if !ok || entity == uuid.Nil {
		return nil, 0, errors.New("invalid accounting resource or scope")
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var rows json.RawMessage
	var total int64
	// The count and page use one snapshot, including when the page is empty.
	err := s.db.DBTX(ctx).QueryRow(ctx, `WITH records AS (`+query+`) SELECT (SELECT count(*) FROM records),coalesce((SELECT jsonb_agg(to_jsonb(p)) FROM (SELECT * FROM records ORDER BY created_at DESC,id DESC LIMIT $2 OFFSET $3) p),'[]'::jsonb)`, entity, limit, offset).Scan(&total, &rows)
	return rows, total, err
}

func (s *ReferenceService) mutate(ctx context.Context, entity, actor uuid.UUID, action string, fn func(context.Context) (uuid.UUID, any, error)) (uuid.UUID, error) {
	if entity == uuid.Nil || actor == uuid.Nil {
		return uuid.Nil, errors.New("legal entity and actor are required")
	}
	var id uuid.UUID
	err := s.ledger.WithinTx(ctx, func(ctx context.Context) error {
		var data any
		var err error
		id, data, err = fn(ctx)
		if err != nil {
			return err
		}
		return s.ledger.RecordSource(ctx, entity, actor, id, action, data)
	})
	return id, err
}

func (s *ReferenceService) CreateDimension(ctx context.Context, entity, actor uuid.UUID, cmd DimensionCommand, value bool) (uuid.UUID, error) {
	cmd.Code = strings.ToUpper(strings.TrimSpace(cmd.Code))
	cmd.NameEn = strings.TrimSpace(cmd.NameEn)
	if cmd.Code == "" || cmd.NameEn == "" {
		return uuid.Nil, errors.New("code and English name are required")
	}
	return s.mutate(ctx, entity, actor, "ACCOUNTING_DIMENSION_CREATED", func(ctx context.Context) (uuid.UUID, any, error) {
		var id uuid.UUID
		var err error
		if value {
			err = s.db.DBTX(ctx).QueryRow(ctx, `INSERT INTO accounting.dimension_values(legal_entity_id,dimension_id,code,name_en,name_ar) VALUES($1,$2,$3,$4,$5) RETURNING id`, entity, cmd.DimensionID, cmd.Code, cmd.NameEn, cmd.NameAr).Scan(&id)
		} else {
			err = s.db.DBTX(ctx).QueryRow(ctx, `INSERT INTO accounting.dimensions(legal_entity_id,code,name_en,name_ar) VALUES($1,$2,$3,$4) RETURNING id`, entity, cmd.Code, cmd.NameEn, cmd.NameAr).Scan(&id)
		}
		return id, cmd, err
	})
}

func (s *ReferenceService) AssignDimension(ctx context.Context, entity, actor, lineID, dimensionID, valueID uuid.UUID) error {
	_, err := s.mutate(ctx, entity, actor, "ACCOUNTING_LINE_DIMENSION_ASSIGNED", func(ctx context.Context) (uuid.UUID, any, error) {
		_, err := s.db.DBTX(ctx).Exec(ctx, `INSERT INTO accounting.journal_line_dimensions(journal_line_id,legal_entity_id,dimension_id,dimension_value_id) VALUES($1,$2,$3,$4) ON CONFLICT(journal_line_id,dimension_id) DO UPDATE SET dimension_value_id=EXCLUDED.dimension_value_id`, lineID, entity, dimensionID, valueID)
		return lineID, map[string]any{"dimension_id": dimensionID, "dimension_value_id": valueID}, err
	})
	return err
}

func (s *ReferenceService) CreateExchangeRate(ctx context.Context, entity, actor uuid.UUID, cmd ExchangeRateCommand) (uuid.UUID, error) {
	date, err := time.Parse(time.DateOnly, cmd.EffectiveDate)
	if err != nil {
		return uuid.Nil, errors.New("invalid effective date")
	}
	cmd.FromCurrency = strings.ToUpper(strings.TrimSpace(cmd.FromCurrency))
	cmd.ToCurrency = strings.ToUpper(strings.TrimSpace(cmd.ToCurrency))
	cmd.Source = strings.TrimSpace(cmd.Source)
	if !cmd.Rate.IsPositive() || !cmd.Rate.Equal(cmd.Rate.Round(12)) || cmd.Source == "" {
		return uuid.Nil, errors.New("source and positive exchange rate with at most 12 decimals are required")
	}
	return s.mutate(ctx, entity, actor, "ACCOUNTING_EXCHANGE_RATE_CREATED", func(ctx context.Context) (uuid.UUID, any, error) {
		var id uuid.UUID
		err := s.db.DBTX(ctx).QueryRow(ctx, `INSERT INTO accounting.exchange_rates(legal_entity_id,from_currency,to_currency,effective_date,rate,version_number,source,created_by) SELECT $1,$2,$3,$4,$5,coalesce(max(version_number),0)+1,$6,$7 FROM accounting.exchange_rates WHERE legal_entity_id=$1 AND from_currency=$2 AND to_currency=$3 AND effective_date=$4 RETURNING id`, entity, cmd.FromCurrency, cmd.ToCurrency, date, cmd.Rate.String(), cmd.Source, actor).Scan(&id)
		return id, cmd, err
	})
}

func (s *ReferenceService) CreatePostingRule(ctx context.Context, entity, actor uuid.UUID, cmd PostingRuleCommand) (uuid.UUID, error) {
	date, err := time.Parse(time.DateOnly, cmd.EffectiveFrom)
	if err != nil {
		return uuid.Nil, errors.New("invalid effective date")
	}
	cmd.SourceModule = strings.ToUpper(strings.TrimSpace(cmd.SourceModule))
	cmd.Code = strings.ToUpper(strings.TrimSpace(cmd.Code))
	cmd.NameEn = strings.TrimSpace(cmd.NameEn)
	if cmd.SourceModule == "" || cmd.Code == "" || cmd.NameEn == "" {
		return uuid.Nil, errors.New("module, code and name are required")
	}
	return s.mutate(ctx, entity, actor, "ACCOUNTING_POSTING_RULE_VERSION_CREATED", func(ctx context.Context) (uuid.UUID, any, error) {
		if _, err := s.ledger.Account(ctx, entity, &cmd.AccountID, ""); err != nil {
			return uuid.Nil, nil, err
		}
		var ruleID, id uuid.UUID
		err := s.db.DBTX(ctx).QueryRow(ctx, `INSERT INTO accounting.posting_rules(legal_entity_id,source_module,code,name_en,name_ar) VALUES($1,$2,$3,$4,$5) ON CONFLICT(legal_entity_id,source_module,code) DO UPDATE SET name_en=EXCLUDED.name_en,name_ar=EXCLUDED.name_ar RETURNING id`, entity, cmd.SourceModule, cmd.Code, cmd.NameEn, cmd.NameAr).Scan(&ruleID)
		if err != nil {
			return uuid.Nil, nil, err
		}
		err = s.db.DBTX(ctx).QueryRow(ctx, `INSERT INTO accounting.posting_rule_versions(legal_entity_id,posting_rule_id,version_number,effective_from,account_id,journal_id,created_by) SELECT $1,$2,coalesce(max(version_number),0)+1,$3,$4,$5,$6 FROM accounting.posting_rule_versions WHERE posting_rule_id=$2 RETURNING id`, entity, ruleID, date, cmd.AccountID, cmd.JournalID, actor).Scan(&id)
		return id, cmd, err
	})
}

func (s *ReferenceService) CreateOpeningBalance(ctx context.Context, entity, actor uuid.UUID, cmd OpeningBalanceCommand) (uuid.UUID, error) {
	if !cmd.Debit.Equal(cmd.Debit.Round(6)) || !cmd.Credit.Equal(cmd.Credit.Round(6)) {
		return uuid.Nil, errors.New("opening amounts accept at most six decimal places")
	}
	return s.mutate(ctx, entity, actor, "ACCOUNTING_OPENING_BALANCE_CREATED", func(ctx context.Context) (uuid.UUID, any, error) {
		if _, err := s.ledger.Account(ctx, entity, &cmd.AccountID, ""); err != nil {
			return uuid.Nil, nil, err
		}
		var id uuid.UUID
		err := s.db.DBTX(ctx).QueryRow(ctx, `INSERT INTO accounting.opening_balances(legal_entity_id,fiscal_period_id,account_id,currency_code,debit,credit,created_by) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`, entity, cmd.FiscalPeriodID, cmd.AccountID, cmd.CurrencyCode, cmd.Debit.String(), cmd.Credit.String(), actor).Scan(&id)
		return id, cmd, err
	})
}

func (s *ReferenceService) PostOpeningBalance(ctx context.Context, entity, actor, id, offsetAccount uuid.UUID, rate decimal.Decimal) (uuid.UUID, error) {
	var journal uuid.UUID
	_, err := s.mutate(ctx, entity, actor, "ACCOUNTING_OPENING_BALANCE_POSTED", func(ctx context.Context) (uuid.UUID, any, error) {
		var account, creator uuid.UUID
		var currency, status string
		var debit, credit decimal.Decimal
		var date time.Time
		err := s.db.DBTX(ctx).QueryRow(ctx, `SELECT o.account_id,o.created_by,o.currency_code,o.status,o.debit::text,o.credit::text,p.start_date FROM accounting.opening_balances o JOIN accounting.fiscal_periods p ON p.id=o.fiscal_period_id WHERE o.id=$1 AND o.legal_entity_id=$2 FOR UPDATE OF o`, id, entity).Scan(&account, &creator, &currency, &status, &debit, &credit, &date)
		if err != nil {
			return id, nil, err
		}
		if status != "DRAFT" || actor == creator {
			return id, nil, errors.New("opening balance must be a draft posted by an independent actor")
		}
		if account == offsetAccount {
			return id, nil, errors.New("opening balance offset must be a different account")
		}
		if _, err = s.ledger.Account(ctx, entity, &offsetAccount, ""); err != nil {
			return id, nil, err
		}
		journal, err = s.ledger.Post(ctx, "GENERAL", domain.PostingRequest{LegalEntityID: entity, SourceModule: "ACCOUNTING", SourceType: "OPENING_BALANCE", SourceID: id, AccountingDate: date, CurrencyCode: currency, ExchangeRate: rate, ActorID: actor, Description: "Opening balance", Lines: []domain.PostingLine{{AccountID: account, TransactionDebit: debit, TransactionCredit: credit}, {AccountID: offsetAccount, TransactionDebit: credit, TransactionCredit: debit}}})
		if err != nil {
			return id, nil, err
		}
		_, err = s.db.DBTX(ctx).Exec(ctx, `UPDATE accounting.opening_balances SET journal_entry_id=$1,status='POSTED' WHERE id=$2`, journal, id)
		return id, map[string]any{"journal_entry_id": journal}, err
	})
	return journal, err
}

func (s *ReferenceService) CreateReconciliation(ctx context.Context, entity, actor uuid.UUID, cmd ReconciliationCommand) (uuid.UUID, error) {
	date, err := time.Parse(time.DateOnly, cmd.ReconciliationDate)
	if err != nil {
		return uuid.Nil, errors.New("invalid reconciliation date")
	}
	if len(cmd.JournalLineIDs) < 2 || len(cmd.JournalLineIDs) > 500 {
		return uuid.Nil, errors.New("select between two and 500 journal lines")
	}
	return s.mutate(ctx, entity, actor, "ACCOUNTING_RECONCILIATION_CREATED", func(ctx context.Context) (uuid.UUID, any, error) {
		var id uuid.UUID
		err := s.db.DBTX(ctx).QueryRow(ctx, `INSERT INTO accounting.reconciliations(legal_entity_id,account_id,reference,reconciliation_date,currency_code,created_by) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, entity, cmd.AccountID, cmd.Reference, date, cmd.CurrencyCode, actor).Scan(&id)
		if err != nil {
			return id, nil, err
		}
		for _, line := range cmd.JournalLineIDs {
			var valid bool
			err = s.db.DBTX(ctx).QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM accounting.journal_lines jl JOIN accounting.journal_entries je ON je.id=jl.journal_entry_id WHERE jl.id=$1 AND jl.legal_entity_id=$2 AND jl.account_id=$3 AND jl.currency_code=$4 AND je.status IN ('POSTED','REVERSED') AND je.accounting_date<=$5)`, line, entity, cmd.AccountID, cmd.CurrencyCode, date).Scan(&valid)
			if err != nil {
				return id, nil, err
			}
			if !valid {
				return id, nil, errors.New("reconciliation lines must be posted in the selected account, currency and date")
			}
			_, err = s.db.DBTX(ctx).Exec(ctx, `INSERT INTO accounting.reconciliation_lines(legal_entity_id,reconciliation_id,journal_line_id) VALUES($1,$2,$3)`, entity, id, line)
			if err != nil {
				return id, nil, err
			}
		}
		return id, cmd, nil
	})
}

func (s *ReferenceService) CompleteReconciliation(ctx context.Context, entity, actor, id uuid.UUID) error {
	_, err := s.mutate(ctx, entity, actor, "ACCOUNTING_RECONCILIATION_COMPLETED", func(ctx context.Context) (uuid.UUID, any, error) {
		result, err := s.db.DBTX(ctx).Exec(ctx, `UPDATE accounting.reconciliations SET status='COMPLETED',completed_by=$1,completed_at=now() WHERE id=$2 AND legal_entity_id=$3 AND status='DRAFT' AND created_by<>$1`, actor, id, entity)
		if err == nil && result.RowsAffected() != 1 {
			err = fmt.Errorf("reconciliation must be a draft completed by an independent actor")
		}
		return id, map[string]any{"status": "COMPLETED"}, err
	})
	return err
}
