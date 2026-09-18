package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/leave/domain"
	"erp/internal/leave/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type leaveService struct {
	repo ports.LeaveRepository
	db   *database.DB
}

func NewLeaveService(repo ports.LeaveRepository, db *database.DB) ports.LeaveService {
	return &leaveService{
		repo: repo,
		db:   db,
	}
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

	balance, err := s.repo.GetEmployeeLeaveBalance(ctx, employeeID, leaveTypeID, fiscalYear)
	if err != nil {
		return nil, err
	}
	if balance == nil {
		balance = &domain.EmployeeLeaveBalance{
			EmployeeID:     employeeID,
			LeaveTypeID:    leaveTypeID,
			FiscalYear:     fiscalYear,
			AllocatedDays:  decimal.Zero,
			AccruedDays:    days,
			TakenDays:      decimal.Zero,
			PendingDays:    decimal.Zero,
			CarryoverDays:  decimal.Zero,
			ClosingBalance: days,
		}
	} else {
		balance.AccruedDays = balance.AccruedDays.Add(days)
		balance.ClosingBalance = balance.ClosingBalance.Add(days)
	}

	updated, err := s.repo.UpsertEmployeeLeaveBalance(ctx, balance)
	if err != nil {
		return nil, err
	}

	// Record transaction ledger entry
	_, _ = s.repo.CreateLeaveTransaction(ctx, &domain.LeaveTransaction{
		EmployeeID:      employeeID,
		LeaveTypeID:     leaveTypeID,
		FiscalYear:      fiscalYear,
		TransactionType: domain.TxAccrual,
		Days:            days,
		Notes:           &reason,
	})

	return updated, nil
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

	// 2. Validate leave balance
	fiscalYear := int32(req.StartDate.Year())
	balance, err := s.repo.GetEmployeeLeaveBalance(ctx, req.EmployeeID, req.LeaveTypeID, fiscalYear)
	if err != nil {
		return nil, err
	}

	// For annual/paid leave, check sufficiency
	lt, _ := s.repo.GetLeaveTypeByID(ctx, req.LegalEntityID, req.LeaveTypeID)
	if lt != nil && lt.Code == "ANNUAL" {
		if balance == nil || balance.ClosingBalance.LessThan(totalDays) {
			return nil, domain.ErrInsufficientBalance
		}

		// Reserve pending days
		balance.PendingDays = balance.PendingDays.Add(totalDays)
		balance.ClosingBalance = balance.ClosingBalance.Sub(totalDays)
		_, err = s.repo.UpsertEmployeeLeaveBalance(ctx, balance)
		if err != nil {
			return nil, err
		}
	}

	req.Status = domain.RequestPending
	return s.repo.CreateLeaveRequest(ctx, req)
}

func (s *leaveService) ApproveLeaveRequest(ctx context.Context, legalEntityID, id uuid.UUID, approverUserID uuid.UUID, notes string) (*domain.LeaveRequest, error) {
	req, err := s.repo.GetLeaveRequestByID(ctx, legalEntityID, id)
	if err != nil {
		return nil, err
	}
	if req.Status != domain.RequestPending {
		return nil, domain.ErrAlreadyProcessed
	}

	fiscalYear := int32(req.StartDate.Year())
	balance, _ := s.repo.GetEmployeeLeaveBalance(ctx, req.EmployeeID, req.LeaveTypeID, fiscalYear)
	if balance != nil {
		// Commit deduction: pending -> taken
		balance.PendingDays = balance.PendingDays.Sub(req.TotalDays)
		if balance.PendingDays.IsNegative() {
			balance.PendingDays = decimal.Zero
		}
		balance.TakenDays = balance.TakenDays.Add(req.TotalDays)
		_, _ = s.repo.UpsertEmployeeLeaveBalance(ctx, balance)

		// Record immutable transaction
		desc := fmt.Sprintf("Approved Leave Request %s (%s to %s)", req.RequestNumber, req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02"))
		_, _ = s.repo.CreateLeaveTransaction(ctx, &domain.LeaveTransaction{
			EmployeeID:      req.EmployeeID,
			LeaveTypeID:     req.LeaveTypeID,
			FiscalYear:      fiscalYear,
			TransactionType: domain.TxUsage,
			Days:            req.TotalDays,
			ReferenceID:     &req.ID,
			Notes:           &desc,
		})
	}

	return s.repo.UpdateLeaveRequestStatus(ctx, legalEntityID, id, domain.RequestApproved, approverUserID, notes)
}

func (s *leaveService) RejectLeaveRequest(ctx context.Context, legalEntityID, id uuid.UUID, approverUserID uuid.UUID, notes string) (*domain.LeaveRequest, error) {
	req, err := s.repo.GetLeaveRequestByID(ctx, legalEntityID, id)
	if err != nil {
		return nil, err
	}
	if req.Status != domain.RequestPending {
		return nil, domain.ErrAlreadyProcessed
	}

	// Revert pending reservation
	fiscalYear := int32(req.StartDate.Year())
	balance, _ := s.repo.GetEmployeeLeaveBalance(ctx, req.EmployeeID, req.LeaveTypeID, fiscalYear)
	if balance != nil {
		balance.PendingDays = balance.PendingDays.Sub(req.TotalDays)
		if balance.PendingDays.IsNegative() {
			balance.PendingDays = decimal.Zero
		}
		balance.ClosingBalance = balance.ClosingBalance.Add(req.TotalDays)
		_, _ = s.repo.UpsertEmployeeLeaveBalance(ctx, balance)
	}

	return s.repo.UpdateLeaveRequestStatus(ctx, legalEntityID, id, domain.RequestRejected, approverUserID, notes)
}

func (s *leaveService) ListRequests(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeaveRequest, error) {
	return s.repo.ListLeaveRequests(ctx, legalEntityID)
}

func (s *leaveService) ListEmployeeRequests(ctx context.Context, employeeID uuid.UUID) ([]*domain.LeaveRequest, error) {
	return s.repo.ListEmployeeLeaveRequests(ctx, employeeID)
}
