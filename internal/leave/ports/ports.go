package ports

import (
	"context"

	"erp/internal/leave/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LeaveRepository interface {
	// Leave Types
	CreateLeaveType(ctx context.Context, lt *domain.LeaveType) (*domain.LeaveType, error)
	GetLeaveTypeByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.LeaveType, error)
	GetLeaveTypeByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.LeaveType, error)
	ListLeaveTypes(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeaveType, error)

	// Policies
	CreateLeavePolicy(ctx context.Context, p *domain.LeavePolicy) (*domain.LeavePolicy, error)
	GetLeavePolicyByType(ctx context.Context, legalEntityID, leaveTypeID uuid.UUID) (*domain.LeavePolicy, error)
	ListLeavePolicies(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeavePolicy, error)

	// Balances
	GetEmployeeLeaveBalance(ctx context.Context, employeeID, leaveTypeID uuid.UUID, fiscalYear int32) (*domain.EmployeeLeaveBalance, error)
	UpsertEmployeeLeaveBalance(ctx context.Context, b *domain.EmployeeLeaveBalance) (*domain.EmployeeLeaveBalance, error)
	ListEmployeeLeaveBalances(ctx context.Context, employeeID uuid.UUID, fiscalYear int32) ([]*domain.EmployeeLeaveBalance, error)

	// Transactions
	CreateLeaveTransaction(ctx context.Context, tx *domain.LeaveTransaction) (*domain.LeaveTransaction, error)
	ListLeaveTransactions(ctx context.Context, employeeID, leaveTypeID uuid.UUID, fiscalYear int32) ([]*domain.LeaveTransaction, error)

	// Requests
	CreateLeaveRequest(ctx context.Context, req *domain.LeaveRequest) (*domain.LeaveRequest, error)
	GetLeaveRequestByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.LeaveRequest, error)
	GetLeaveRequestByNumber(ctx context.Context, legalEntityID uuid.UUID, requestNumber string) (*domain.LeaveRequest, error)
	ListLeaveRequests(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeaveRequest, error)
	ListEmployeeLeaveRequests(ctx context.Context, employeeID uuid.UUID) ([]*domain.LeaveRequest, error)
	UpdateLeaveRequestStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.LeaveStatusCode, reviewedBy uuid.UUID, notes string) (*domain.LeaveRequest, error)
}

type LeaveService interface {
	// Types & Policies
	CreateLeaveType(ctx context.Context, lt *domain.LeaveType) (*domain.LeaveType, error)
	ListLeaveTypes(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeaveType, error)
	DefinePolicy(ctx context.Context, p *domain.LeavePolicy) (*domain.LeavePolicy, error)
	ListPolicies(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeavePolicy, error)

	// Balances
	GetEmployeeBalances(ctx context.Context, employeeID uuid.UUID, fiscalYear int32) ([]*domain.EmployeeLeaveBalance, error)
	AccrueLeave(ctx context.Context, employeeID, leaveTypeID uuid.UUID, fiscalYear int32, days decimal.Decimal, reason string) (*domain.EmployeeLeaveBalance, error)

	// Requests Workflow
	SubmitLeaveRequest(ctx context.Context, req *domain.LeaveRequest) (*domain.LeaveRequest, error)
	ApproveLeaveRequest(ctx context.Context, legalEntityID, id uuid.UUID, approverUserID uuid.UUID, notes string) (*domain.LeaveRequest, error)
	RejectLeaveRequest(ctx context.Context, legalEntityID, id uuid.UUID, approverUserID uuid.UUID, notes string) (*domain.LeaveRequest, error)
	ListRequests(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.LeaveRequest, error)
	ListEmployeeRequests(ctx context.Context, employeeID uuid.UUID) ([]*domain.LeaveRequest, error)
}
