package postgres

import (
	"context"
	"fmt"

	"erp/internal/authorization/domain"
	"erp/internal/authorization/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AssignmentRepository struct {
	db *database.DB
}

func NewAssignmentRepository(db *database.DB) ports.AssignmentRepository {
	return &AssignmentRepository{db: db}
}

func (r *AssignmentRepository) Assign(ctx context.Context, userID, roleID uuid.UUID, legalEntityID, branchID *uuid.UUID) (*domain.RoleAssignment, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.AssignUserRole(ctx, sqlc.AssignUserRoleParams{
		UserID:        toPgUUID(userID),
		RoleID:        toPgUUID(roleID),
		LegalEntityID: toPgUUIDPtr(legalEntityID),
		BranchID:      toPgUUIDPtr(branchID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to assign user role: %w", err)
	}

	return &domain.RoleAssignment{
		ID:            toUUID(row.ID),
		UserID:        toUUID(row.UserID),
		RoleID:        toUUID(row.RoleID),
		LegalEntityID: toUUIDPtr(row.LegalEntityID),
		BranchID:      toUUIDPtr(row.BranchID),
		AssignedAt:    row.AssignedAt.Time,
	}, nil
}

func (r *AssignmentRepository) Revoke(ctx context.Context, userID, roleID uuid.UUID, legalEntityID, branchID *uuid.UUID) error {
	queries := sqlc.New(r.db.DBTX(ctx))

	return queries.RevokeUserRole(ctx, sqlc.RevokeUserRoleParams{
		UserID:        toPgUUID(userID),
		RoleID:        toPgUUID(roleID),
		LegalEntityID: toPgUUIDPtr(legalEntityID),
		BranchID:      toPgUUIDPtr(branchID),
	})
}

func (r *AssignmentRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.RoleAssignment, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListUserAssignments(ctx, toPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to list user assignments: %w", err)
	}

	res := make([]*domain.RoleAssignment, len(rows))
	for i, row := range rows {
		res[i] = &domain.RoleAssignment{
			ID:                toUUID(row.ID),
			UserID:            toUUID(row.UserID),
			RoleID:            toUUID(row.RoleID),
			RoleCode:          row.RoleCode,
			RoleNameAr:        row.RoleNameAr,
			RoleNameEn:        row.RoleNameEn,
			LegalEntityID:     toUUIDPtr(row.LegalEntityID),
			LegalEntityNameAr: fromPgText(row.LegalEntityNameAr),
			LegalEntityNameEn: fromPgText(row.LegalEntityNameEn),
			BranchID:          toUUIDPtr(row.BranchID),
			BranchNameAr:      fromPgText(row.BranchNameAr),
			BranchNameEn:      fromPgText(row.BranchNameEn),
			AssignedAt:        row.AssignedAt.Time,
		}
	}
	return res, nil
}

func (r *AssignmentRepository) GetPermissionsInScope(ctx context.Context, userID uuid.UUID, legalEntityID *uuid.UUID) ([]string, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	var leID pgtype.UUID
	if legalEntityID != nil && *legalEntityID != uuid.Nil {
		leID = toPgUUID(*legalEntityID)
	} else {
		leID = pgtype.UUID{Valid: false}
	}

	perms, err := queries.GetUserPermissionsInScope(ctx, sqlc.GetUserPermissionsInScopeParams{
		UserID:        toPgUUID(userID),
		LegalEntityID: leID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate scoped permissions: %w", err)
	}

	return perms, nil
}
