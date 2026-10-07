package accounting_test

import (
	"errors"
	"testing"
	"time"

	"erp/internal/accounting/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

func TestPostingEngine_RejectsCrossEntityDimensions(t *testing.T) {
	mod, _, _, db, entityID, actorID, ctx := setupAccountingTest(t)
	defer db.Close()
	journals, err := mod.AccountingService.ListJournals(ctx, entityID)
	if err != nil || len(journals) == 0 {
		t.Fatalf("journals: %v", err)
	}
	cash, err := mod.AccountingService.GetAccountByCode(ctx, entityID, "111002")
	if err != nil {
		t.Fatal(err)
	}
	revenue, err := mod.AccountingService.GetAccountByCode(ctx, entityID, "410000")
	if err != nil {
		t.Fatal(err)
	}
	otherEntityID := uuid.New()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO organization.legal_entities(id, organization_id, code, name_ar, name_en)
		SELECT $1, organization_id, $2, 'منشأة أخرى', 'Other entity' FROM organization.legal_entities WHERE id=$3`,
		otherEntityID, "DIM-"+otherEntityID.String(), entityID); err != nil {
		t.Fatal(err)
	}
	for _, dimension := range []struct {
		table  string
		column string
		assign func(*domain.PostingLine, *uuid.UUID)
	}{
		{"organization.branches", "branch_id", func(line *domain.PostingLine, id *uuid.UUID) { line.BranchID = id }},
		{"organization.departments", "department_id", func(line *domain.PostingLine, id *uuid.UUID) { line.DepartmentID = id }},
		{"organization.cost_centers", "cost_center_id", func(line *domain.PostingLine, id *uuid.UUID) { line.CostCenterID = id }},
	} {
		t.Run(dimension.table, func(t *testing.T) {
			for _, ownerID := range []uuid.UUID{otherEntityID, entityID} {
				dimensionID := uuid.New()
				if _, err := db.Pool.Exec(ctx, "INSERT INTO "+dimension.table+"(id, legal_entity_id, code, name_ar, name_en) VALUES($1,$2,$3,'بعد','Dimension')",
					dimensionID, ownerID, dimensionID.String()); err != nil {
					t.Fatal(err)
				}
				request := domain.PostingRequest{
					LegalEntityID: entityID, SourceModule: "TEST", SourceType: "DIMENSION", SourceID: uuid.New(),
					JournalID: journals[0].ID, AccountingDate: time.Now().UTC(), ActorID: actorID,
					Lines: []domain.PostingLine{
						{AccountID: cash.ID, TransactionDebit: decimal.NewFromInt(83)},
						{AccountID: revenue.ID, TransactionCredit: decimal.NewFromInt(83)},
					},
				}
				dimension.assign(&request.Lines[0], &dimensionID)
				_, err := mod.PostingEngine.Post(ctx, request)
				if ownerID == entityID {
					if err != nil {
						t.Fatalf("own-entity dimension must be accepted: %v", err)
					}
					_, err := db.Pool.Exec(ctx, "UPDATE "+dimension.table+" SET legal_entity_id=$1 WHERE id=$2", otherEntityID, dimensionID)
					var pgErr *pgconn.PgError
					if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
						t.Fatalf("database must prevent moving a posted dimension to another entity: %v", err)
					}
					continue
				}
				if !errors.Is(err, domain.ErrInvalidPostingDimension) {
					t.Fatalf("expected cross-entity dimension rejection: %v", err)
				}
				var count int
				if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM accounting.journal_entries WHERE source_id=$1", request.SourceID).Scan(&count); err != nil {
					t.Fatal(err)
				}
				if count != 0 {
					t.Fatal("cross-entity dimension left a journal header")
				}
				invalidLines := []domain.JournalLine{
					{AccountID: cash.ID, TransactionDebit: decimal.NewFromInt(83), BranchID: request.Lines[0].BranchID, DepartmentID: request.Lines[0].DepartmentID, CostCenterID: request.Lines[0].CostCenterID},
					{AccountID: revenue.ID, TransactionCredit: decimal.NewFromInt(83)},
				}
				_, err = mod.LedgerService.CreateJournalEntry(ctx, domain.JournalEntry{
					LegalEntityID: entityID, JournalID: journals[0].ID, AccountingDate: request.AccountingDate,
					CreatedBy: actorID, Description: "Manual dimension validation",
				}, invalidLines)
				if !errors.Is(err, domain.ErrInvalidPostingDimension) {
					t.Fatalf("manual draft must reject foreign dimension: %v", err)
				}
				draft, err := mod.LedgerService.CreateJournalEntry(ctx, domain.JournalEntry{
					LegalEntityID: entityID, JournalID: journals[0].ID, AccountingDate: request.AccountingDate,
					CreatedBy: actorID, Description: "Direct SQL boundary test",
				}, []domain.JournalLine{
					{AccountID: cash.ID, TransactionDebit: decimal.NewFromInt(83)},
					{AccountID: revenue.ID, TransactionCredit: decimal.NewFromInt(83)},
				})
				if err != nil {
					t.Fatal(err)
				}
				_, err = db.Pool.Exec(ctx, "UPDATE accounting.journal_lines SET "+dimension.column+"=$1 WHERE journal_entry_id=$2 AND line_number=1", dimensionID, draft.ID)
				var pgErr *pgconn.PgError
				if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
					t.Fatalf("database must reject cross-entity dimension SQL bypass: %v", err)
				}
			}
		})
	}
}
