package postgres

import (
	"context"
	"errors"

	"erp/internal/leave/domain"
	"erp/internal/leave/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type leaveRepository struct {
	db *database.DB
	q  *sqlc.Queries
}

func NewLeaveRepository(db *database.DB) ports.LeaveRepository {
	return &leaveRepository{
		db: db,
		q:  sqlc.New(db.Pool),
	}
}

// ============================================================================
// LEAVE TYPES
// ============================================================================

func (r *leaveRepository) CreateLeaveType(ctx context.Context, lt *domain.LeaveType) (*domain.LeaveType, error) {
	if lt.ID == uuid.Nil {
		lt.ID = uuid.New()
	}
	row, err := r.q.CreateLeaveType(ctx, sqlc.CreateLeaveTypeParams{
		ID:                 toPgUUID(lt.ID),
		LegalEntityID:      toPgUUID(lt.LegalEntityID),
		Code:               lt.Code,
		NameEn:             lt.NameEn,
		NameAr:             lt.NameAr,
		IsPaid:             lt.IsPaid,
		RequiresAttachment: lt.RequiresAttachment,
		DeductsFromAnnual:  lt.DeductsFromAnnual,
		MaxDaysPerYear:     toPgNumericPtr(lt.MaxDaysPerYear),
		IsActive:           lt.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return mapLeaveType(row), nil
}

func (r *leaveRepository) GetLeaveTypeByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.LeaveType, error) {
	row, err := r.q.GetLeaveTypeByID(ctx, sqlc.GetLeaveTypeByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLeaveTypeNotFound
		}
		return nil, err
	}
	return mapLeaveType(row), nil
}

func (r *leaveRepository) GetLeaveTypeByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.LeaveType, error) {
	row, err := r.q.GetLeaveTypeByCode(ctx, sqlc.GetLeaveTypeByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLeaveTypeNotFound
		}
		return nil, err
	}
	return mapLeaveType(row), nil
}

func (r *leaveRepository) ListLeaveTypes(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeaveType, error) {
	rows, err := r.q.ListLeaveTypes(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.LeaveType, len(rows))
	for i, row := range rows {
		res[i] = mapLeaveType(row)
	}
	return res, nil
}

// ============================================================================
// POLICIES
// ============================================================================

func (r *leaveRepository) CreateLeavePolicy(ctx context.Context, p *domain.LeavePolicy) (*domain.LeavePolicy, error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	row, err := r.q.CreateLeavePolicy(ctx, sqlc.CreateLeavePolicyParams{
		ID:                     toPgUUID(p.ID),
		LegalEntityID:          toPgUUID(p.LegalEntityID),
		LeaveTypeID:            toPgUUID(p.LeaveTypeID),
		AccrualFrequency:       string(p.AccrualFrequency),
		AccrualRatePerPeriod:   toPgNumeric(p.AccrualRatePerPeriod),
		MaxCarryoverDays:       toPgNumeric(p.MaxCarryoverDays),
		ProbationRestricted:    p.ProbationRestricted,
		EncashmentAllowed:      p.EncashmentAllowed,
		MinServiceDaysRequired: p.MinServiceDaysRequired,
		IsActive:               p.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return mapLeavePolicy(row), nil
}

func (r *leaveRepository) GetLeavePolicyByType(ctx context.Context, legalEntityID, leaveTypeID uuid.UUID) (*domain.LeavePolicy, error) {
	row, err := r.q.GetLeavePolicyByType(ctx, sqlc.GetLeavePolicyByTypeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		LeaveTypeID:   toPgUUID(leaveTypeID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPolicyNotFound
		}
		return nil, err
	}
	return mapLeavePolicy(row), nil
}

func (r *leaveRepository) ListLeavePolicies(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeavePolicy, error) {
	rows, err := r.q.ListLeavePolicies(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.LeavePolicy, len(rows))
	for i, row := range rows {
		code := row.LeaveTypeCode
		nameEn := row.LeaveTypeNameEn
		nameAr := row.LeaveTypeNameAr
		p := mapLeavePolicy(sqlc.LeavePolicies{
			ID:                     row.ID,
			LegalEntityID:          row.LegalEntityID,
			LeaveTypeID:            row.LeaveTypeID,
			AccrualFrequency:       row.AccrualFrequency,
			AccrualRatePerPeriod:   row.AccrualRatePerPeriod,
			MaxCarryoverDays:       row.MaxCarryoverDays,
			ProbationRestricted:    row.ProbationRestricted,
			EncashmentAllowed:      row.EncashmentAllowed,
			MinServiceDaysRequired: row.MinServiceDaysRequired,
			IsActive:               row.IsActive,
			CreatedAt:              row.CreatedAt,
			UpdatedAt:              row.UpdatedAt,
		})
		p.LeaveTypeCode = &code
		p.LeaveTypeNameEn = &nameEn
		p.LeaveTypeNameAr = &nameAr
		res[i] = p
	}
	return res, nil
}

// ============================================================================
// BALANCES
// ============================================================================

func (r *leaveRepository) GetEmployeeLeaveBalance(ctx context.Context, employeeID, leaveTypeID uuid.UUID, fiscalYear int32) (*domain.EmployeeLeaveBalance, error) {
	row, err := r.q.GetEmployeeLeaveBalance(ctx, sqlc.GetEmployeeLeaveBalanceParams{
		EmployeeID:  toPgUUID(employeeID),
		LeaveTypeID: toPgUUID(leaveTypeID),
		FiscalYear:  fiscalYear,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapLeaveBalance(row), nil
}

func (r *leaveRepository) UpsertEmployeeLeaveBalance(ctx context.Context, b *domain.EmployeeLeaveBalance) (*domain.EmployeeLeaveBalance, error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	row, err := r.q.UpsertEmployeeLeaveBalance(ctx, sqlc.UpsertEmployeeLeaveBalanceParams{
		ID:             toPgUUID(b.ID),
		EmployeeID:     toPgUUID(b.EmployeeID),
		LeaveTypeID:    toPgUUID(b.LeaveTypeID),
		FiscalYear:     b.FiscalYear,
		AllocatedDays:  toPgNumeric(b.AllocatedDays),
		AccruedDays:    toPgNumeric(b.AccruedDays),
		TakenDays:      toPgNumeric(b.TakenDays),
		PendingDays:    toPgNumeric(b.PendingDays),
		CarryoverDays:  toPgNumeric(b.CarryoverDays),
		ClosingBalance: toPgNumeric(b.ClosingBalance),
	})
	if err != nil {
		return nil, err
	}
	return mapLeaveBalance(row), nil
}

func (r *leaveRepository) ListEmployeeLeaveBalances(ctx context.Context, employeeID uuid.UUID, fiscalYear int32) ([]*domain.EmployeeLeaveBalance, error) {
	rows, err := r.q.ListEmployeeLeaveBalances(ctx, sqlc.ListEmployeeLeaveBalancesParams{
		EmployeeID: toPgUUID(employeeID),
		FiscalYear: fiscalYear,
	})
	if err != nil {
		return nil, err
	}
	res := make([]*domain.EmployeeLeaveBalance, len(rows))
	for i, row := range rows {
		code := row.LeaveTypeCode
		nameEn := row.LeaveTypeNameEn
		nameAr := row.LeaveTypeNameAr
		b := mapLeaveBalance(sqlc.LeaveEmployeeBalances{
			ID:             row.ID,
			EmployeeID:     row.EmployeeID,
			LeaveTypeID:    row.LeaveTypeID,
			FiscalYear:     row.FiscalYear,
			AllocatedDays:  row.AllocatedDays,
			AccruedDays:    row.AccruedDays,
			TakenDays:      row.TakenDays,
			PendingDays:    row.PendingDays,
			CarryoverDays:  row.CarryoverDays,
			ClosingBalance: row.ClosingBalance,
			UpdatedAt:      row.UpdatedAt,
		})
		b.LeaveTypeCode = &code
		b.LeaveTypeNameEn = &nameEn
		b.LeaveTypeNameAr = &nameAr
		res[i] = b
	}
	return res, nil
}

// ============================================================================
// TRANSACTIONS
// ============================================================================

func (r *leaveRepository) CreateLeaveTransaction(ctx context.Context, tx *domain.LeaveTransaction) (*domain.LeaveTransaction, error) {
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	row, err := r.q.CreateLeaveTransaction(ctx, sqlc.CreateLeaveTransactionParams{
		ID:              toPgUUID(tx.ID),
		EmployeeID:      toPgUUID(tx.EmployeeID),
		LeaveTypeID:     toPgUUID(tx.LeaveTypeID),
		FiscalYear:      tx.FiscalYear,
		TransactionType: string(tx.TransactionType),
		Days:            toPgNumeric(tx.Days),
		ReferenceID:     toPgUUIDPtr(tx.ReferenceID),
		Notes:           toPgTextPtr(tx.Notes),
	})
	if err != nil {
		return nil, err
	}
	return mapLeaveTransaction(row), nil
}

func (r *leaveRepository) ListLeaveTransactions(ctx context.Context, employeeID, leaveTypeID uuid.UUID, fiscalYear int32) ([]*domain.LeaveTransaction, error) {
	rows, err := r.q.ListLeaveTransactions(ctx, sqlc.ListLeaveTransactionsParams{
		EmployeeID:  toPgUUID(employeeID),
		LeaveTypeID: toPgUUID(leaveTypeID),
		FiscalYear:  fiscalYear,
	})
	if err != nil {
		return nil, err
	}
	res := make([]*domain.LeaveTransaction, len(rows))
	for i, row := range rows {
		res[i] = mapLeaveTransaction(row)
	}
	return res, nil
}

// ============================================================================
// REQUESTS
// ============================================================================

func (r *leaveRepository) CreateLeaveRequest(ctx context.Context, req *domain.LeaveRequest) (*domain.LeaveRequest, error) {
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	row, err := r.q.CreateLeaveRequest(ctx, sqlc.CreateLeaveRequestParams{
		ID:                     toPgUUID(req.ID),
		LegalEntityID:          toPgUUID(req.LegalEntityID),
		EmployeeID:            toPgUUID(req.EmployeeID),
		LeaveTypeID:            toPgUUID(req.LeaveTypeID),
		RequestNumber:          req.RequestNumber,
		StartDate:              toPgDate(req.StartDate),
		EndDate:                toPgDate(req.EndDate),
		TotalDays:              toPgNumeric(req.TotalDays),
		ReturnDate:             toPgDate(req.ReturnDate),
		ReplacementEmployeeID: toPgUUIDPtr(req.ReplacementEmployeeID),
		Reason:                 req.Reason,
		AttachmentUrl:          toPgTextPtr(req.AttachmentURL),
		Status:                 string(req.Status),
	})
	if err != nil {
		return nil, err
	}
	return mapLeaveRequest(row), nil
}

func (r *leaveRepository) GetLeaveRequestByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.LeaveRequest, error) {
	row, err := r.q.GetLeaveRequestByID(ctx, sqlc.GetLeaveRequestByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLeaveRequestNotFound
		}
		return nil, err
	}
	return mapLeaveRequest(row), nil
}

func (r *leaveRepository) GetLeaveRequestByNumber(ctx context.Context, legalEntityID uuid.UUID, requestNumber string) (*domain.LeaveRequest, error) {
	row, err := r.q.GetLeaveRequestByNumber(ctx, sqlc.GetLeaveRequestByNumberParams{
		LegalEntityID: toPgUUID(legalEntityID),
		RequestNumber: requestNumber,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLeaveRequestNotFound
		}
		return nil, err
	}
	return mapLeaveRequest(row), nil
}

func (r *leaveRepository) ListLeaveRequests(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeaveRequest, error) {
	rows, err := r.q.ListLeaveRequests(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.LeaveRequest, len(rows))
	for i, row := range rows {
		empNum := row.EmployeeNumber
		nameEn := row.FirstNameEn + " " + row.LastNameEn
		nameAr := row.FirstNameAr + " " + row.LastNameAr
		typeCode := row.LeaveTypeCode
		typeNameEn := row.LeaveTypeNameEn
		typeNameAr := row.LeaveTypeNameAr

		req := mapLeaveRequest(sqlc.LeaveRequests{
			ID:                     row.ID,
			LegalEntityID:          row.LegalEntityID,
			EmployeeID:            row.EmployeeID,
			LeaveTypeID:            row.LeaveTypeID,
			RequestNumber:          row.RequestNumber,
			StartDate:              row.StartDate,
			EndDate:                row.EndDate,
			TotalDays:              row.TotalDays,
			ReturnDate:             row.ReturnDate,
			ReplacementEmployeeID: row.ReplacementEmployeeID,
			Reason:                 row.Reason,
			AttachmentUrl:          row.AttachmentUrl,
			Status:                 row.Status,
			ReviewedByUserID:      row.ReviewedByUserID,
			ReviewedAt:            row.ReviewedAt,
			ReviewNotes:           row.ReviewNotes,
			CreatedAt:             row.CreatedAt,
			UpdatedAt:             row.UpdatedAt,
		})
		req.EmployeeNumber = &empNum
		req.EmployeeNameEn = &nameEn
		req.EmployeeNameAr = &nameAr
		req.LeaveTypeCode = &typeCode
		req.LeaveTypeNameEn = &typeNameEn
		req.LeaveTypeNameAr = &typeNameAr
		res[i] = req
	}
	return res, nil
}

func (r *leaveRepository) ListEmployeeLeaveRequests(ctx context.Context, employeeID uuid.UUID) ([]*domain.LeaveRequest, error) {
	rows, err := r.q.ListEmployeeLeaveRequests(ctx, toPgUUID(employeeID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.LeaveRequest, len(rows))
	for i, row := range rows {
		typeCode := row.LeaveTypeCode
		typeNameEn := row.LeaveTypeNameEn
		typeNameAr := row.LeaveTypeNameAr
		req := mapLeaveRequest(sqlc.LeaveRequests{
			ID:                     row.ID,
			LegalEntityID:          row.LegalEntityID,
			EmployeeID:            row.EmployeeID,
			LeaveTypeID:            row.LeaveTypeID,
			RequestNumber:          row.RequestNumber,
			StartDate:              row.StartDate,
			EndDate:                row.EndDate,
			TotalDays:              row.TotalDays,
			ReturnDate:             row.ReturnDate,
			ReplacementEmployeeID: row.ReplacementEmployeeID,
			Reason:                 row.Reason,
			AttachmentUrl:          row.AttachmentUrl,
			Status:                 row.Status,
			ReviewedByUserID:      row.ReviewedByUserID,
			ReviewedAt:            row.ReviewedAt,
			ReviewNotes:           row.ReviewNotes,
			CreatedAt:             row.CreatedAt,
			UpdatedAt:             row.UpdatedAt,
		})
		req.LeaveTypeCode = &typeCode
		req.LeaveTypeNameEn = &typeNameEn
		req.LeaveTypeNameAr = &typeNameAr
		res[i] = req
	}
	return res, nil
}

func (r *leaveRepository) UpdateLeaveRequestStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.LeaveStatusCode, reviewedBy uuid.UUID, notes string) (*domain.LeaveRequest, error) {
	row, err := r.q.UpdateLeaveRequestStatus(ctx, sqlc.UpdateLeaveRequestStatusParams{
		ID:               toPgUUID(id),
		LegalEntityID:    toPgUUID(legalEntityID),
		Status:           string(status),
		ReviewedByUserID: toPgUUIDPtr(&reviewedBy),
		ReviewNotes:      toPgTextPtr(&notes),
	})
	if err != nil {
		return nil, err
	}
	return mapLeaveRequest(row), nil
}

// ============================================================================
// MAPPERS
// ============================================================================

func mapLeaveType(r sqlc.LeaveTypes) *domain.LeaveType {
	return &domain.LeaveType{
		ID:                 toUUID(r.ID),
		LegalEntityID:      toUUID(r.LegalEntityID),
		Code:               r.Code,
		NameEn:             r.NameEn,
		NameAr:             r.NameAr,
		IsPaid:             r.IsPaid,
		RequiresAttachment: r.RequiresAttachment,
		DeductsFromAnnual:  r.DeductsFromAnnual,
		MaxDaysPerYear:     fromPgNumericPtr(r.MaxDaysPerYear),
		IsActive:           r.IsActive,
		CreatedAt:          fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:          fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapLeavePolicy(r sqlc.LeavePolicies) *domain.LeavePolicy {
	return &domain.LeavePolicy{
		ID:                     toUUID(r.ID),
		LegalEntityID:          toUUID(r.LegalEntityID),
		LeaveTypeID:            toUUID(r.LeaveTypeID),
		AccrualFrequency:       domain.AccrualFrequency(r.AccrualFrequency),
		AccrualRatePerPeriod:   fromPgNumeric(r.AccrualRatePerPeriod),
		MaxCarryoverDays:       fromPgNumeric(r.MaxCarryoverDays),
		ProbationRestricted:    r.ProbationRestricted,
		EncashmentAllowed:      r.EncashmentAllowed,
		MinServiceDaysRequired: r.MinServiceDaysRequired,
		IsActive:               r.IsActive,
		CreatedAt:              fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:              fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapLeaveBalance(r sqlc.LeaveEmployeeBalances) *domain.EmployeeLeaveBalance {
	return &domain.EmployeeLeaveBalance{
		ID:             toUUID(r.ID),
		EmployeeID:     toUUID(r.EmployeeID),
		LeaveTypeID:    toUUID(r.LeaveTypeID),
		FiscalYear:     r.FiscalYear,
		AllocatedDays:  fromPgNumeric(r.AllocatedDays),
		AccruedDays:    fromPgNumeric(r.AccruedDays),
		TakenDays:      fromPgNumeric(r.TakenDays),
		PendingDays:    fromPgNumeric(r.PendingDays),
		CarryoverDays:  fromPgNumeric(r.CarryoverDays),
		ClosingBalance: fromPgNumeric(r.ClosingBalance),
		UpdatedAt:      fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapLeaveTransaction(r sqlc.LeaveTransactions) *domain.LeaveTransaction {
	return &domain.LeaveTransaction{
		ID:              toUUID(r.ID),
		EmployeeID:      toUUID(r.EmployeeID),
		LeaveTypeID:     toUUID(r.LeaveTypeID),
		FiscalYear:      r.FiscalYear,
		TransactionType: domain.TransactionType(r.TransactionType),
		Days:            fromPgNumeric(r.Days),
		ReferenceID:     toUUIDPtr(r.ReferenceID),
		Notes:           fromPgTextPtr(r.Notes),
		CreatedAt:       fromPgTimestamptz(r.CreatedAt),
	}
}

func mapLeaveRequest(r sqlc.LeaveRequests) *domain.LeaveRequest {
	return &domain.LeaveRequest{
		ID:                    toUUID(r.ID),
		LegalEntityID:         toUUID(r.LegalEntityID),
		EmployeeID:            toUUID(r.EmployeeID),
		LeaveTypeID:           toUUID(r.LeaveTypeID),
		RequestNumber:         r.RequestNumber,
		StartDate:             fromPgDate(r.StartDate),
		EndDate:               fromPgDate(r.EndDate),
		TotalDays:             fromPgNumeric(r.TotalDays),
		ReturnDate:            fromPgDate(r.ReturnDate),
		ReplacementEmployeeID: toUUIDPtr(r.ReplacementEmployeeID),
		Reason:                r.Reason,
		AttachmentURL:         fromPgTextPtr(r.AttachmentUrl),
		Status:                domain.LeaveStatusCode(r.Status),
		ReviewedByUserID:      toUUIDPtr(r.ReviewedByUserID),
		ReviewedAt:            fromPgTimestamptzPtr(r.ReviewedAt),
		ReviewNotes:           fromPgTextPtr(r.ReviewNotes),
		CreatedAt:             fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:             fromPgTimestamptz(r.UpdatedAt),
	}
}
