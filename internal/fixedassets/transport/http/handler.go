package http

import (
	"encoding/json"
	iamhttp "erp/internal/iam/transport/http"
	"net/http"
	"time"

	"erp/internal/fixedassets/domain"
	"erp/internal/fixedassets/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type FixedAssetsHandler struct {
	svc ports.Service
}

func NewFixedAssetsHandler(svc ports.Service) *FixedAssetsHandler {
	return &FixedAssetsHandler{svc: svc}
}

func (h *FixedAssetsHandler) Routes(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		if authMiddleware != nil {
			protected.Use(authMiddleware)
		}

		// Categories & Locations
		protected.Get("/categories", h.ListCategories)
		protected.Post("/categories", h.CreateCategory)
		protected.Get("/locations", h.ListLocations)
		protected.Post("/locations", h.CreateLocation)
		protected.Get("/books", h.ListBooks)

		// Assets
		protected.Get("/assets", h.ListAssets)
		protected.Post("/assets", h.RegisterAsset)
		protected.Get("/assets/{id}", h.GetAsset)
		protected.Post("/assets/{id}/assign", h.AssignAsset)
		protected.Post("/assets/{id}/transfer", h.TransferAsset)
		protected.Post("/assets/{id}/dispose", h.DisposeAsset)
		protected.Get("/assets/{id}/schedule", h.GetAssetSchedule)

		// Depreciation Runs
		protected.Get("/depreciation/runs", h.ListDepreciationRuns)
		protected.Post("/depreciation/runs/calculate", h.RunMonthlyDepreciation)
		protected.Get("/depreciation/runs/{id}", h.GetDepreciationRun)
		protected.Post("/depreciation/runs/{id}/post", h.PostDepreciationRun)

		// Transfers & Dashboard
		protected.Get("/transfers", h.ListTransfers)
		protected.Get("/dashboard", h.GetDashboardSummary)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ----------------------------------------------------------------------
// Categories & Locations
// ----------------------------------------------------------------------

func (h *FixedAssetsHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	cats, err := h.svc.ListAssetCategories(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

type CreateCategoryRequest struct {
	LegalEntityID                    string          `json:"legal_entity_id"`
	Code                             string          `json:"code"`
	NameEn                           string          `json:"name_en"`
	NameAr                           string          `json:"name_ar"`
	DescriptionEn                    string          `json:"description_en"`
	DescriptionAr                    string          `json:"description_ar"`
	DepreciationMethod               string          `json:"depreciation_method"`
	UsefulLifeMonths                 int             `json:"useful_life_months"`
	SalvageValueRate                 decimal.Decimal `json:"salvage_value_rate"`
	AssetAccountID                   *string         `json:"asset_account_id"`
	AccumulatedDepreciationAccountID *string         `json:"accumulated_depreciation_account_id"`
	DepreciationExpenseAccountID     *string         `json:"depreciation_expense_account_id"`
}

func (h *FixedAssetsHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}

	var assetAccID, accumAccID, expAccID *uuid.UUID
	if req.AssetAccountID != nil {
		if id, err := uuid.Parse(*req.AssetAccountID); err == nil {
			assetAccID = &id
		}
	}
	if req.AccumulatedDepreciationAccountID != nil {
		if id, err := uuid.Parse(*req.AccumulatedDepreciationAccountID); err == nil {
			accumAccID = &id
		}
	}
	if req.DepreciationExpenseAccountID != nil {
		if id, err := uuid.Parse(*req.DepreciationExpenseAccountID); err == nil {
			expAccID = &id
		}
	}

	cat, err := h.svc.CreateAssetCategory(r.Context(), domain.AssetCategory{
		LegalEntityID:                    leID,
		Code:                             req.Code,
		NameEn:                           req.NameEn,
		NameAr:                           req.NameAr,
		DescriptionEn:                    req.DescriptionEn,
		DescriptionAr:                    req.DescriptionAr,
		DepreciationMethod:               req.DepreciationMethod,
		UsefulLifeMonths:                 req.UsefulLifeMonths,
		SalvageValueRate:                 req.SalvageValueRate,
		AssetAccountID:                   assetAccID,
		AccumulatedDepreciationAccountID: accumAccID,
		DepreciationExpenseAccountID:     expAccID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, cat)
}

func (h *FixedAssetsHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	locs, err := h.svc.ListAssetLocations(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, locs)
}

type CreateLocationRequest struct {
	LegalEntityID string `json:"legal_entity_id"`
	Code          string `json:"code"`
	NameEn        string `json:"name_en"`
	NameAr        string `json:"name_ar"`
	Building      string `json:"building"`
	Floor         string `json:"floor"`
	Room          string `json:"room"`
}

func (h *FixedAssetsHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var req CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}

	loc, err := h.svc.CreateAssetLocation(r.Context(), domain.AssetLocation{
		LegalEntityID: leID,
		Code:          req.Code,
		NameEn:        req.NameEn,
		NameAr:        req.NameAr,
		Building:      req.Building,
		Floor:         req.Floor,
		Room:          req.Room,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, loc)
}

func (h *FixedAssetsHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	books, err := h.svc.ListDepreciationBooks(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, books)
}

// ----------------------------------------------------------------------
// Assets
// ----------------------------------------------------------------------

type RegisterAssetRequest struct {
	LegalEntityID      string          `json:"legal_entity_id"`
	CategoryID         string          `json:"category_id"`
	LocationID         *string         `json:"location_id"`
	AssetNumber        string          `json:"asset_number"`
	SerialNumber       string          `json:"serial_number"`
	Barcode            string          `json:"barcode"`
	NameEn             string          `json:"name_en"`
	NameAr             string          `json:"name_ar"`
	DescriptionEn      string          `json:"description_en"`
	DescriptionAr      string          `json:"description_ar"`
	AcquisitionDate    string          `json:"acquisition_date"`
	InServiceDate      string          `json:"in_service_date"`
	AcquisitionCost    decimal.Decimal `json:"acquisition_cost"`
	SalvageValue       decimal.Decimal `json:"salvage_value"`
	UsefulLifeMonths   int             `json:"useful_life_months"`
	DepreciationMethod string          `json:"depreciation_method"`
}

func (h *FixedAssetsHandler) RegisterAsset(w http.ResponseWriter, r *http.Request) {
	var req RegisterAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}
	catID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category_id")
		return
	}

	var locID *uuid.UUID
	if req.LocationID != nil && *req.LocationID != "" {
		if id, err := uuid.Parse(*req.LocationID); err == nil {
			locID = &id
		}
	}

	acqDate, _ := time.Parse("2006-01-02", req.AcquisitionDate)
	if acqDate.IsZero() {
		acqDate = time.Now()
	}

	inServDate, _ := time.Parse("2006-01-02", req.InServiceDate)
	if inServDate.IsZero() {
		inServDate = acqDate
	}

	asset := domain.Asset{
		LegalEntityID:      leID,
		CategoryID:         catID,
		LocationID:         locID,
		AssetNumber:        req.AssetNumber,
		SerialNumber:       req.SerialNumber,
		Barcode:            req.Barcode,
		NameEn:             req.NameEn,
		NameAr:             req.NameAr,
		DescriptionEn:      req.DescriptionEn,
		DescriptionAr:      req.DescriptionAr,
		AcquisitionDate:    acqDate,
		InServiceDate:      inServDate,
		AcquisitionCost:    req.AcquisitionCost,
		SalvageValue:       req.SalvageValue,
		UsefulLifeMonths:   req.UsefulLifeMonths,
		DepreciationMethod: req.DepreciationMethod,
	}

	saved, err := h.svc.RegisterAsset(r.Context(), asset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, saved)
}

func (h *FixedAssetsHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	assets, err := h.svc.ListAssets(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, assets)
}

func (h *FixedAssetsHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid asset id")
		return
	}

	asset, err := h.svc.GetAsset(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, asset)
}

type AssignAssetRequest struct {
	AssignedToType string  `json:"assigned_to_type"`
	EmployeeID     *string `json:"employee_id"`
	DepartmentID   *string `json:"department_id"`
	AssignedDate   string  `json:"assigned_date"`
	Notes          string  `json:"notes"`
}

func (h *FixedAssetsHandler) AssignAsset(w http.ResponseWriter, r *http.Request) {
	assetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid asset id")
		return
	}

	var req AssignAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var empID, deptID *uuid.UUID
	if req.EmployeeID != nil && *req.EmployeeID != "" {
		if id, err := uuid.Parse(*req.EmployeeID); err == nil {
			empID = &id
		}
	}
	if req.DepartmentID != nil && *req.DepartmentID != "" {
		if id, err := uuid.Parse(*req.DepartmentID); err == nil {
			deptID = &id
		}
	}

	asDate, _ := time.Parse("2006-01-02", req.AssignedDate)
	if asDate.IsZero() {
		asDate = time.Now()
	}

	assignment, err := h.svc.AssignAsset(r.Context(), domain.AssetAssignment{
		AssetID:        assetID,
		AssignedToType: req.AssignedToType,
		EmployeeID:     empID,
		DepartmentID:   deptID,
		AssignedDate:   asDate,
		Notes:          req.Notes,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, assignment)
}

type TransferAssetRequest struct {
	LegalEntityID  string  `json:"legal_entity_id"`
	ToLocationID   string  `json:"to_location_id"`
	FromLocationID *string `json:"from_location_id"`
	ToEmployeeID   *string `json:"to_employee_id"`
	FromEmployeeID *string `json:"from_employee_id"`
	Reason         string  `json:"reason"`
}

func (h *FixedAssetsHandler) TransferAsset(w http.ResponseWriter, r *http.Request) {
	assetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid asset id")
		return
	}

	var req TransferAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, _ := uuid.Parse(req.LegalEntityID)
	toLocID, err := uuid.Parse(req.ToLocationID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to_location_id")
		return
	}

	var fromLocID, toEmpID, fromEmpID *uuid.UUID
	if req.FromLocationID != nil && *req.FromLocationID != "" {
		if id, err := uuid.Parse(*req.FromLocationID); err == nil {
			fromLocID = &id
		}
	}
	if req.ToEmployeeID != nil && *req.ToEmployeeID != "" {
		if id, err := uuid.Parse(*req.ToEmployeeID); err == nil {
			toEmpID = &id
		}
	}
	if req.FromEmployeeID != nil && *req.FromEmployeeID != "" {
		if id, err := uuid.Parse(*req.FromEmployeeID); err == nil {
			fromEmpID = &id
		}
	}

	transfer, err := h.svc.TransferAsset(r.Context(), domain.Transfer{
		LegalEntityID:  leID,
		AssetID:        assetID,
		ToLocationID:   toLocID,
		FromLocationID: fromLocID,
		ToEmployeeID:   toEmpID,
		FromEmployeeID: fromEmpID,
		Reason:         req.Reason,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, transfer)
}

type DisposeAssetRequest struct {
	ImpairmentAccountID *uuid.UUID      `json:"impairment_account_id"`
	ProceedsAccountID   *uuid.UUID      `json:"proceeds_account_id"`
	GainLossAccountID   *uuid.UUID      `json:"gain_loss_account_id"`
	LegalEntityID       string          `json:"legal_entity_id"`
	DisposalType        string          `json:"disposal_type"`
	ProceedsAmount      decimal.Decimal `json:"proceeds_amount"`
	Reason              string          `json:"reason"`
}

func (h *FixedAssetsHandler) DisposeAsset(w http.ResponseWriter, r *http.Request) {
	actor := iamhttp.UserFromContext(r.Context())
	if actor == nil || actor.ID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	assetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid asset id")
		return
	}

	var req DisposeAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, _ := uuid.Parse(req.LegalEntityID)

	disp, err := h.svc.DisposeAsset(r.Context(), domain.Disposal{
		LegalEntityID:       leID,
		AssetID:             assetID,
		DisposalType:        req.DisposalType,
		ProceedsAmount:      req.ProceedsAmount,
		ImpairmentAccountID: req.ImpairmentAccountID, ProceedsAccountID: req.ProceedsAccountID, GainLossAccountID: req.GainLossAccountID, CreatedBy: &actor.ID,
		Reason: req.Reason,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, disp)
}

func (h *FixedAssetsHandler) GetAssetSchedule(w http.ResponseWriter, r *http.Request) {
	assetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid asset id")
		return
	}
	bookID, err := uuid.Parse(r.URL.Query().Get("book_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "book_id query parameter is required")
		return
	}

	scheds, err := h.svc.GetAssetSchedule(r.Context(), assetID, bookID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, scheds)
}

// ----------------------------------------------------------------------
// Depreciation Runs
// ----------------------------------------------------------------------

type RunMonthlyDepreciationRequest struct {
	LegalEntityID string `json:"legal_entity_id"`
	BookID        string `json:"book_id"`
	Year          int    `json:"year"`
	Month         int    `json:"month"`
}

func (h *FixedAssetsHandler) RunMonthlyDepreciation(w http.ResponseWriter, r *http.Request) {
	actor := iamhttp.UserFromContext(r.Context())
	if actor == nil || actor.ID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var req RunMonthlyDepreciationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}
	bookID, err := uuid.Parse(req.BookID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book_id")
		return
	}

	run, lines, err := h.svc.RunMonthlyDepreciation(r.Context(), leID, bookID, req.Year, req.Month, &actor.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"run":   run,
		"lines": lines,
	})
}

func (h *FixedAssetsHandler) PostDepreciationRun(w http.ResponseWriter, r *http.Request) {
	actor := iamhttp.UserFromContext(r.Context())
	if actor == nil || actor.ID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}

	posted, err := h.svc.PostDepreciationRun(r.Context(), runID, &actor.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, posted)
}

func (h *FixedAssetsHandler) GetDepreciationRun(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}

	run, lines, err := h.svc.GetDepreciationRun(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"run":   run,
		"lines": lines,
	})
}

func (h *FixedAssetsHandler) ListDepreciationRuns(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	runs, err := h.svc.ListDepreciationRuns(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (h *FixedAssetsHandler) ListTransfers(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	transfers, err := h.svc.ListTransfers(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, transfers)
}

func (h *FixedAssetsHandler) GetDashboardSummary(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	summary, err := h.svc.GetDashboardSummary(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
