package http

import (
	"encoding/json"
	"net/http"

	"context"
	auditports "erp/internal/audit/ports"
	iamhttp "erp/internal/iam/transport/http"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type ApprovalLimitHandler struct {
	db    *database.DB
	audit auditports.AuditService
}

func NewApprovalLimitHandler(db *database.DB, audit auditports.AuditService) *ApprovalLimitHandler {
	return &ApprovalLimitHandler{db: db, audit: audit}
}

type approvalLimit struct {
	LegalEntityID     uuid.UUID       `json:"legal_entity_id"`
	UserID            uuid.UUID       `json:"user_id"`
	ResourceType      string          `json:"resource_type"`
	MaximumBaseAmount decimal.Decimal `json:"maximum_base_amount"`
	Version           int64           `json:"version"`
}

func (h *ApprovalLimitHandler) List(w http.ResponseWriter, r *http.Request) {
	entity, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, 400, "INVALID_ENTITY", "legal_entity_id is required")
		return
	}
	rows, err := h.db.Pool.Query(r.Context(), `SELECT legal_entity_id,user_id,resource_type,maximum_base_amount,version FROM workflow.approval_limits WHERE legal_entity_id=$1 ORDER BY resource_type,user_id`, entity)
	if err != nil {
		writeError(w, 500, "READ_FAILED", "Could not read approval limits")
		return
	}
	defer rows.Close()
	items := []approvalLimit{}
	for rows.Next() {
		var item approvalLimit
		if err = rows.Scan(&item.LegalEntityID, &item.UserID, &item.ResourceType, &item.MaximumBaseAmount, &item.Version); err != nil {
			writeError(w, 500, "READ_FAILED", "Could not read approval limits")
			return
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		writeError(w, 500, "READ_FAILED", "Could not read approval limits")
		return
	}
	writeJSON(w, 200, map[string]any{"data": items})
}

func (h *ApprovalLimitHandler) Save(w http.ResponseWriter, r *http.Request) {
	actor := iamhttp.UserFromContext(r.Context())
	if actor == nil {
		writeError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}
	var req approvalLimit
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, 400, "INVALID_BODY", "Invalid approval limit")
		return
	}
	if req.LegalEntityID == uuid.Nil || req.UserID == uuid.Nil || req.MaximumBaseAmount.IsNegative() || !req.MaximumBaseAmount.Equal(req.MaximumBaseAmount.Round(6)) || req.Version < 0 {
		writeError(w, 400, "INVALID_LIMIT", "Entity, user, non-negative six-decimal limit and expected version are required")
		return
	}
	switch req.ResourceType {
	case "SALES_INVOICE", "TREASURY_PAYMENT", "PAYROLL_RUN":
	default:
		writeError(w, 400, "INVALID_RESOURCE", "Unsupported document type")
		return
	}
	err := transaction.NewManager(h.db.Pool).WithinTx(r.Context(), func(ctx context.Context) error {
		if err := h.audit.LockChain(ctx); err != nil {
			return err
		}
		var active bool
		if err := h.db.DBTX(ctx).QueryRow(ctx, "SELECT status='ACTIVE' FROM iam.users WHERE id=$1 FOR SHARE", req.UserID).Scan(&active); err != nil {
			return err
		}
		if !active {
			return fmt.Errorf("approver must be active")
		}
		var id uuid.UUID
		err := h.db.DBTX(ctx).QueryRow(ctx, `INSERT INTO workflow.approval_limits(legal_entity_id,user_id,resource_type,maximum_base_amount,updated_by)
		 SELECT $1,$2,$3,$4,$5 WHERE $6::bigint=0
		 ON CONFLICT (legal_entity_id,user_id,resource_type) DO NOTHING RETURNING id,version`, req.LegalEntityID, req.UserID, req.ResourceType, req.MaximumBaseAmount, actor.ID, req.Version).Scan(&id, &req.Version)
		if errors.Is(err, pgx.ErrNoRows) && req.Version > 0 {
			err = h.db.DBTX(ctx).QueryRow(ctx, `UPDATE workflow.approval_limits SET maximum_base_amount=$4,updated_by=$5,version=version+1,updated_at=now()
			 WHERE legal_entity_id=$1 AND user_id=$2 AND resource_type=$3 AND version=$6 RETURNING id,version`, req.LegalEntityID, req.UserID, req.ResourceType, req.MaximumBaseAmount, actor.ID, req.Version).Scan(&id, &req.Version)
		}
		if err != nil {
			return err
		}
		payload, err := json.Marshal(req)
		if err != nil {
			return err
		}
		_, err = h.audit.RecordEvent(ctx, auditports.RecordEventCommand{UserID: &actor.ID, LegalEntityID: &req.LegalEntityID, Action: "workflow.approval_limit.save", ResourceType: "workflow.approval_limit", ResourceID: id.String(), Payload: payload})
		return err
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, 409, "VERSION_CONFLICT", "Approval limit changed; refresh and retry")
		return
	}
	if err != nil {
		writeError(w, 400, "SAVE_FAILED", "Could not save approval limit")
		return
	}
	writeJSON(w, 200, map[string]any{"data": req})
}
