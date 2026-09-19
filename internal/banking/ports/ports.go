package ports

import (
	"context"
	"erp/internal/banking/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BankingRepository interface {
	// Bank Accounts
	CreateBankAccount(ctx context.Context, account *domain.BankAccount) (*domain.BankAccount, error)
	GetBankAccountByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.BankAccount, error)
	GetBankAccountByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.BankAccount, error)
	GetBankAccountByIBAN(ctx context.Context, legalEntityID uuid.UUID, iban string) (*domain.BankAccount, error)
	ListBankAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.BankAccount, error)
	UpdateBankAccountBalance(ctx context.Context, legalEntityID, id uuid.UUID, delta decimal.Decimal) (*domain.BankAccount, error)
	UpdateBankAccountStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.BankAccountStatus) (*domain.BankAccount, error)

	// Statements
	CreateBankStatement(ctx context.Context, statement *domain.Statement) (*domain.Statement, error)
	GetBankStatementByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Statement, error)
	ListBankStatements(ctx context.Context, bankAccountID uuid.UUID) ([]*domain.Statement, error)
	UpdateBankStatementStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.StatementStatus) (*domain.Statement, error)

	// Statement Lines
	CreateStatementLine(ctx context.Context, line *domain.StatementLine) (*domain.StatementLine, error)
	GetStatementLineByID(ctx context.Context, id uuid.UUID) (*domain.StatementLine, error)
	LockStatementLineByID(ctx context.Context, id uuid.UUID) (*domain.StatementLine, error)
	ListStatementLines(ctx context.Context, statementID uuid.UUID) ([]*domain.StatementLine, error)
	UpdateStatementLineReconciledStatus(ctx context.Context, id uuid.UUID, status domain.ReconciledStatus) (*domain.StatementLine, error)

	// Reconciliations
	CreateBankReconciliation(ctx context.Context, recon *domain.Reconciliation) (*domain.Reconciliation, error)
	GetBankReconciliationByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Reconciliation, error)
	LockBankReconciliationByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Reconciliation, error)
	ListBankReconciliations(ctx context.Context, bankAccountID uuid.UUID) ([]*domain.Reconciliation, error)
	UpdateBankReconciliation(ctx context.Context, recon *domain.Reconciliation) (*domain.Reconciliation, error)

	// Matches
	CreateReconciliationMatch(ctx context.Context, match *domain.ReconciliationMatch) (*domain.ReconciliationMatch, error)
	ListReconciliationMatches(ctx context.Context, reconciliationID uuid.UUID) ([]*domain.ReconciliationMatch, error)
	DeleteReconciliationMatch(ctx context.Context, id uuid.UUID) error
}

type BankingService interface {
	// Bank Accounts
	CreateBankAccount(ctx context.Context, account *domain.BankAccount) (*domain.BankAccount, error)
	GetBankAccount(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.BankAccount, error)
	ListBankAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.BankAccount, error)

	// Statements
	ImportBankStatement(ctx context.Context, statement *domain.Statement, lines []*domain.StatementLine) (*domain.Statement, error)
	GetBankStatement(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Statement, error)
	ListBankStatements(ctx context.Context, legalEntityID, bankAccountID uuid.UUID) ([]*domain.Statement, error)
	ListStatementLines(ctx context.Context, legalEntityID, statementID uuid.UUID) ([]*domain.StatementLine, error)

	// Reconciliation
	CreateReconciliation(ctx context.Context, recon *domain.Reconciliation) (*domain.Reconciliation, error)
	GetReconciliation(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Reconciliation, error)
	ListReconciliations(ctx context.Context, legalEntityID, bankAccountID uuid.UUID) ([]*domain.Reconciliation, error)
	RunAutoMatch(ctx context.Context, legalEntityID, reconciliationID uuid.UUID, candidateEntries []domain.CandidateLedgerEntry) ([]domain.AutoMatchResult, error)
	AddManualMatch(ctx context.Context, legalEntityID uuid.UUID, match *domain.ReconciliationMatch) (*domain.ReconciliationMatch, error)
	LockReconciliation(ctx context.Context, legalEntityID, id uuid.UUID, lockedBy uuid.UUID) (*domain.Reconciliation, error)
	ListMatches(ctx context.Context, legalEntityID, reconciliationID uuid.UUID) ([]*domain.ReconciliationMatch, error)
}
