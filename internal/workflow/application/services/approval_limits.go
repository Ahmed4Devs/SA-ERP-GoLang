package services

import (
	"context"
	"errors"
	"fmt"

	"erp/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// CheckApprovalLimit compares a server-computed base-currency amount to the
// approver's configured authority. A zero limit blocks every positive amount.
// Absence of a limit retains the assigned workflow role's unrestricted authority.
func CheckApprovalLimit(ctx context.Context, db *database.DB, entity, actor uuid.UUID, resourceType string, baseAmount decimal.Decimal) error {
	if entity == uuid.Nil || actor == uuid.Nil || baseAmount.IsNegative() {
		return fmt.Errorf("valid entity, approver and non-negative base amount are required")
	}
	var maximum decimal.Decimal
	err := db.DBTX(ctx).QueryRow(ctx, `SELECT maximum_base_amount FROM workflow.approval_limits
		WHERE legal_entity_id=$1 AND user_id=$2 AND resource_type=$3 FOR SHARE`, entity, actor, resourceType).Scan(&maximum)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if baseAmount.GreaterThan(maximum) {
		return fmt.Errorf("document base amount exceeds the approver's configured limit")
	}
	return nil
}
