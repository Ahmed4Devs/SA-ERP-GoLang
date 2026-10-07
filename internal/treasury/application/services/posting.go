package services

import (
	"context"
	"encoding/json"
	"fmt"

	auditports "erp/internal/audit/ports"
	"erp/internal/treasury/domain"
	"github.com/google/uuid"
)

func validateChecker(maker *uuid.UUID, actor uuid.UUID) error {
	if maker == nil || *maker == uuid.Nil {
		return domain.ErrMissingMaker
	}
	if *maker == actor {
		return domain.ErrMakerCannotCheck
	}
	return nil
}

func (s *treasuryService) requiredAccount(ctx context.Context, entity uuid.UUID, account *uuid.UUID) (uuid.UUID, error) {
	if account == nil || *account == uuid.Nil {
		return uuid.Nil, fmt.Errorf("treasury GL account mapping is required")
	}
	return s.ledger.Account(ctx, entity, account, "")
}

func (s *treasuryService) recordTransition(ctx context.Context, entity, actor uuid.UUID, action, table string, id uuid.UUID, before, after any) error {
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return err
	}
	_, err = s.audit.RecordEvent(ctx, auditports.RecordEventCommand{
		UserID: &actor, LegalEntityID: &entity, Action: "treasury." + action,
		ResourceType: "treasury." + table, ResourceID: id.String(), Payload: afterJSON,
		Changes: []auditports.EntityChangeCommand{{TableName: "treasury." + table, RecordID: id.String(), BeforeState: beforeJSON, AfterState: afterJSON}},
	})
	return err
}
