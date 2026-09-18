package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/leave/domain"
	"erp/internal/leave/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type leaveService struct {
	repo ports.LeaveRepository
	db   *database.DB
	tx   transaction.Manager
}

func NewLeaveService(repo ports.LeaveRepository, db *database.DB) ports.LeaveService {
	return &leaveService{
		repo: repo,
		db:   db,
		tx:   transaction.NewManager(db.Pool),
	}
}

func (s *leaveService) lockBalance(ctx context.Context, employeeID, leaveTypeID uuid.UUID, fiscalYear int32) error {
	key := fmt.Sprintf("leave-balance:%s:%s:%d", employeeID, leaveTypeID, fiscalYear)
	if _, err := s.db.DBTX(ctx).Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", key); err != nil {
		return fmt.Errorf("failed to lock leave balance: %w", err)
	}
	return nil
}

// ============================================================================
// TYPES & POLICIES
// ============================================================================

func (s *leaveService) CreateLeaveType(ctx context.Context, lt *domain.LeaveType) (*domain.LeaveType, error) {
	if lt.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if lt.Code == "" {
		return nil, fmt.Errorf("leave type code is required")
	}
	if lt.NameEn == "" || lt.NameAr == "" {
		return nil, fmt.Errorf("bilingual leave type names are required")
	}
	return s.repo.CreateLeaveType(ctx, lt)
}

func (s *leaveService) ListLeaveTypes(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeaveType, error) {
	return s.repo.ListLeaveTypes(ctx, legalEntityID)
}

func (s *leaveService) DefinePolicy(ctx context.Context, p *domain.LeavePolicy) (*domain.LeavePolicy, error) {
	if p.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if p.LeaveTypeID == uuid.Nil {
		return nil, fmt.Errorf("leave_type_id is required")
	}
	if p.AccrualFrequency == "" {
		p.AccrualFrequency = domain.AccrualMonthly
	}
	if p.AccrualRatePerPeriod.IsZero() {
		p.AccrualRatePerPeriod = decimal.NewFromFloat(1.75) // 21 days annual
	}
	p.IsActive = true
	return s.repo.CreateLeavePolicy(ctx, p)
}

func (s *leaveService) ListPolicies(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeavePolicy, error) {
	return s.repo.ListLeavePolicies(ctx, legalEntityID)
}

// ============================================================================
// BALANCES
// ============================================================================

func (s *leaveService) GetEmployeeBalances(ctx context.Context, employeeID uuid.UUID, fiscalYear int32) ([]*domain.EmployeeLeaveBalance, error) {
	if fiscalYear <= 0 {
		fiscalYear = int32(time.Now().Year())
	}
	return s.repo.ListEmployeeLeaveBalances(ctx, employeeID, fiscalYear)
}

func (s *leaveService) AccrueLeave(
	ctx context.Context,
	employeeID, leaveTypeID uuid.UUID,
	fiscalYear int32,
	days decimal.Decimal,
	reason string,
) (*domain.EmployeeLeaveBalance, error) {
	if fiscalYear <= 0 {
		fiscalYear = int32(time.Now().Year())
	}

	var updated *domain.EmployeeLeaveBalance
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.lockBalance(txCtx, employeeID, leaveTypeID, fiscalYear); err != nil {
			return err
		}
		balance, err := s.repo.GetEmployeeLeaveBalance(txCtx, employeeID, leaveTypeID, fiscalYear)
		if err != nil {
			return err
		}
		if balance == nil {
			balance = &domain.EmployeeLeaveBalance{
				EmployeeID: employeeID, LeaveTypeID: leaveTypeID, FiscalYear: fiscalYear,
				AccruedDays: days, ClosingBalance: days,
			}
		} else {
			balance.AccruedDays = balance.AccruedDays.Add(days)
			balance.ClosingBalance = balance.ClosingBalance.Add(days)
		}
		updated, err = s.repo.UpsertEmployeeLeaveBalance(txCtx, balance)
		if err != nil {
			return err
		}
		if _, err := s.repo.CreateLeaveTransaction(txCtx, &domain.LeaveTransaction{
			EmployeeID: employeeID, LeaveTypeID: leaveTypeID, FiscalYear: fiscalYear,
			TransactionType: domain.TxAccrual, Days: days, Notes: &reason,
		}); err != nil {
			return fmt.Errorf("failed to record leave accrual transaction: %w", err)
		}
		return nil
	})
	return updated, err
}

// ============================================================================
// REQUESTS WORKFLOW
// ============================================================================

func (s *leaveService) SubmitLeaveRequest(ctx context.Context, req *domain.LeaveRequest) (*domain.LeaveRequest, error) {
	if req.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if req.EmployeeID == uuid.Nil {
		return nil, fmt.Errorf("employee_id is required")
	}
	if req.LeaveTypeID == uuid.Nil {
		return nil, fmt.Errorf("leave_type_id is required")
	}
	if req.EndDate.Before(req.StartDate) {
		return nil, domain.ErrInvalidLeaveDates
	}

	// 1. Calculate working days
	totalDays := domain.CalculateWorkingDays(req.StartDate, req.EndDate)
	if totalDays.IsZero() {
		totalDays = decimal.NewFromInt(1)
	}
	req.TotalDays = totalDays
	if req.ReturnDate.IsZero() {
		req.ReturnDate = req.EndDate.AddDate(0, 0, 1)
	}
	if req.RequestNumber == "" {
		req.RequestNumber = identifiers.NewDocumentNumber("LR", time.Now())
	}

	req.Status = domain.RequestPending
	var created *domain.LeaveRequest
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		fiscalYear := int32(req.StartDate.Year())
		lt, err := s.repo.GetLeaveTypeByID(txCtx, req.LegalEntityID, req.LeaveTypeID)
		if err != nil {
			return err
		}
		if lt.Code == "ANNUAL" {
			if err := s.lockBalance(txCtx, req.EmployeeID, req.LeaveTypeID, fiscalYear); err != nil {
				return err
			}
			balance, err := s.repo.GetEmployeeLeaveBalance(txCtx, req.EmployeeID, req.LeaveTypeID, fiscalYear)
			if err != nil {
				return err
			}
			if balance == nil || balance.ClosingBalance.LessThan(totalDays) {
				return domain.ErrInsufficientBalance
			}
			balance.PendingDays = balance.PendingDays.Add(totalDays)
			balance.ClosingBalance = balance.ClosingBalance.Sub(totalDays)
			if _, err := s.repo.UpsertEmployeeLeaveBalance(txCtx, balance); err != nil {
				return err
			}
		}
		created, err = s.repo.CreateLeaveRequest(txCtx, req)
		return err
	})
	return created, err
}

func (s *leaveService) ApproveLeaveRequest(ctx context.Context, legalEntityID, id uuid.UUID, approverUserID uuid.UUID, notes string) (*domain.LeaveRequest, error) {
	var approved *domain.LeaveRequest
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.db.DBTX(txCtx).QueryRow(txCtx, "SELECT id FROM leave.requests WHERE id = $1 AND legal_entity_id = $2 FOR UPDATE", id, legalEntityID).Scan(&id); err != nil {
			return domain.ErrLeaveRequestNotFound
		}
		req, err := s.repo.GetLeaveRequestByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if req.Status != domain.RequestPending {
			return domain.ErrAlreadyProcessed
		}
		fiscalYear := int32(req.StartDate.Year())
		if err := s.lockBalance(txCtx, req.EmployeeID, req.LeaveTypeID, fiscalYear); err != nil {
			return err
		}
		balance, err := s.repo.GetEmployeeLeaveBalance(txCtx, req.EmployeeID, req.LeaveTypeID, fiscalYear)
		if err != nil {
			return err
		}
		if balance != nil {
			balance.PendingDays = balance.PendingDays.Sub(req.TotalDays)
			if balance.PendingDays.IsNegative() {
				balance.PendingDays = decimal.Zero
			}
			balance.TakenDays = balance.TakenDays.Add(req.TotalDays)
			if _, err := s.repo.UpsertEmployeeLeaveBalance(txCtx, balance); err != nil {
				return err
			}
			desc := fmt.Sprintf("Approved Leave Request %s (%s to %s)", req.RequestNumber, req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02"))
			if _, err := s.repo.CreateLeaveTransaction(txCtx, &domain.LeaveTransaction{
				EmployeeID: req.EmployeeID, LeaveTypeID: req.LeaveTypeID, FiscalYear: fiscalYear,
				TransactionType: domain.TxUsage, Days: req.TotalDays, ReferenceID: &req.ID, Notes: &desc,
			}); err != nil {
				return fmt.Errorf("failed to record approved leave transaction: %w", err)
			}
		}
		approved, err = s.repo.UpdateLeaveRequestStatus(txCtx, legalEntityID, id, domain.RequestApproved, approverUserID, notes)
		return err
	})
	return approved, err
}

func (s *leaveService) RejectLeaveRequest(ctx context.Context, legalEntityID, id uuid.UUID, approverUserID uuid.UUID, notes string) (*domain.LeaveRequest, error) {
	var rejected *domain.LeaveRequest
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.db.DBTX(txCtx).QueryRow(txCtx, "SELECT id FROM leave.requests WHERE id = $1 AND legal_entity_id = $2 FOR UPDATE", id, legalEntityID).Scan(&id); err != nil {
			return domain.ErrLeaveRequestNotFound
		}
		req, err := s.repo.GetLeaveRequestByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if req.Status != domain.RequestPending {
			return domain.ErrAlreadyProcessed
		}
		fiscalYear := int32(req.StartDate.Year())
		if err := s.lockBalance(txCtx, req.EmployeeID, req.LeaveTypeID, fiscalYear); err != nil {
			return err
		}
		balance, err := s.repo.GetEmployeeLeaveBalance(txCtx, req.EmployeeID, req.LeaveTypeID, fiscalYear)
		if err != nil {
			return err
		}
		if balance != nil {
			balance.PendingDays = balance.PendingDays.Sub(req.TotalDays)
			if balance.PendingDays.IsNegative() {
				balance.PendingDays = decimal.Zero
			}
			balance.ClosingBalance = balance.ClosingBalance.Add(req.TotalDays)
			if _, err := s.repo.UpsertEmployeeLeaveBalance(txCtx, balance); err != nil {
				return err
			}
		}
		rejected, err = s.repo.UpdateLeaveRequestStatus(txCtx, legalEntityID, id, domain.RequestRejected, approverUserID, notes)
		return err
	})
	return rejected, err
}

func (s *leaveService) ListRequests(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeaveRequest, error) {
	return s.repo.ListLeaveRequests(ctx, legalEntityID)
}

func (s *leaveService) ListEmployeeRequests(ctx context.Context, employeeID uuid.UUID) ([]*domain.LeaveRequest, error) {
	return s.repo.ListEmployeeLeaveRequests(ctx, employeeID)
}
