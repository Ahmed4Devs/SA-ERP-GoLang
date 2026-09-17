package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"erp/internal/documents/domain"
	"erp/internal/documents/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// DocumentHandler handles HTTP requests for the documents management platform.
type DocumentHandler struct {
	svc ports.Service
}

// NewDocumentHandler creates a new DocumentHandler instance.
func NewDocumentHandler(svc ports.Service) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

// Routes configures HTTP routes for documents.
func (h *DocumentHandler) Routes(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		if authMiddleware != nil {
			protected.Use(authMiddleware)
		}

		protected.Post("/upload", h.Upload)
		protected.Get("/", h.List)
		protected.Get("/summary", h.Summary)
		protected.Get("/{id}", h.Get)
		protected.Get("/{id}/download", h.Download)
		protected.Post("/{id}/versions", h.UploadVersion)
		protected.Delete("/{id}", h.Archive)
		protected.Post("/{id}/attach", h.Attach)
		protected.Delete("/{id}/attach", h.Detach)
		protected.Get("/entities/{type}/{id}", h.ListEntityAttachments)
	})

	return r
}

// Upload handles multipart file upload.
func (h *DocumentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Max upload size 50MB
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
		return
	}

	leIDStr := r.FormValue("legal_entity_id")
	if leIDStr == "" {
		leIDStr = r.URL.Query().Get("legal_entity_id")
	}
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required in multipart form")
		return
	}
	defer file.Close()

	title := r.FormValue("title")
	if strings.TrimSpace(title) == "" {
		title = header.Filename
	}

	category := domain.FileCategory(strings.ToUpper(strings.TrimSpace(r.FormValue("category"))))
	if category == "" {
		category = domain.CategoryGeneral
	}

	folder := r.FormValue("folder_path")
	if folder == "" {
		folder = "/"
	}

	var entityIDPtr *uuid.UUID
	entityType := r.FormValue("entity_type")
	if entIDStr := r.FormValue("entity_id"); entIDStr != "" {
		if parsedID, err := uuid.Parse(entIDStr); err == nil {
			entityIDPtr = &parsedID
		}
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	params := domain.UploadFileParams{
		LegalEntityID:    leID,
		Title:            title,
		Description:      r.FormValue("description"),
		Category:         category,
		FolderPath:       folder,
		OriginalFilename: header.Filename,
		ContentType:      contentType,
		Content:          file,
		EntityType:       entityType,
		EntityID:         entityIDPtr,
	}

	created, err := h.svc.UploadFile(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// List returns documents matching query filters.
func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	filter := domain.FileFilter{
		LegalEntityID:   leID,
		IncludeArchived: r.URL.Query().Get("include_archived") == "true",
		Limit:           50,
		Offset:          0,
	}

	if catStr := r.URL.Query().Get("category"); catStr != "" {
		cat := domain.FileCategory(strings.ToUpper(catStr))
		filter.Category = &cat
	}

	if folderStr := r.URL.Query().Get("folder_path"); folderStr != "" {
		filter.FolderPath = &folderStr
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = &search
	}

	if entType := r.URL.Query().Get("entity_type"); entType != "" {
		filter.EntityType = &entType
	}

	if entIDStr := r.URL.Query().Get("entity_id"); entIDStr != "" {
		if parsed, err := uuid.Parse(entIDStr); err == nil {
			filter.EntityID = &parsed
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			filter.Limit = val
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			filter.Offset = val
		}
	}

	files, total, err := h.svc.ListFiles(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":   files,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// Get returns details for a single document.
func (h *DocumentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid file id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	file, err := h.svc.GetFile(r.Context(), leID, id)
	if err != nil {
		if errors.Is(err, domain.ErrFileNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, file)
}

// Download streams the binary content of a file.
func (h *DocumentHandler) Download(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid file id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	var versionIDPtr *uuid.UUID
	if verIDStr := r.URL.Query().Get("version_id"); verIDStr != "" {
		if parsedVer, err := uuid.Parse(verIDStr); err == nil {
			versionIDPtr = &parsedVer
		}
	}

	stream, ver, err := h.svc.DownloadFile(r.Context(), leID, id, versionIDPtr)
	if err != nil {
		if errors.Is(err, domain.ErrFileNotFound) || errors.Is(err, domain.ErrVersionNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", ver.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(ver.ByteSize, 10))
	w.Header().Set("ETag", fmt.Sprintf("\"%s\"", ver.SHA256Hash))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", ver.OriginalFilename))

	_, _ = io.Copy(w, stream)
}

// UploadVersion adds a new version to a file.
func (h *DocumentHandler) UploadVersion(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid file id")
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
		return
	}

	leIDStr := r.FormValue("legal_entity_id")
	if leIDStr == "" {
		leIDStr = r.URL.Query().Get("legal_entity_id")
	}
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required in multipart form")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	params := domain.UploadVersionParams{
		FileID:           id,
		LegalEntityID:    leID,
		OriginalFilename: header.Filename,
		ContentType:      contentType,
		Content:          file,
	}

	ver, err := h.svc.UploadVersion(r.Context(), params)
	if err != nil {
		if errors.Is(err, domain.ErrFileNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ver)
}

// Archive marks a file as archived.
func (h *DocumentHandler) Archive(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid file id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	if err := h.svc.ArchiveFile(r.Context(), leID, id); err != nil {
		if errors.Is(err, domain.ErrFileNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

// Attach links a document to an entity.
func (h *DocumentHandler) Attach(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid file id")
		return
	}

	var req struct {
		EntityType string    `json:"entity_type"`
		EntityID   uuid.UUID `json:"entity_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json request body: "+err.Error())
		return
	}

	if err := h.svc.AttachToEntity(r.Context(), id, req.EntityType, req.EntityID, nil); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "attached"})
}

// Detach removes a link between a document and an entity.
func (h *DocumentHandler) Detach(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid file id")
		return
	}

	var req struct {
		EntityType string    `json:"entity_type"`
		EntityID   uuid.UUID `json:"entity_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json request body: "+err.Error())
		return
	}

	if err := h.svc.DetachFromEntity(r.Context(), id, req.EntityType, req.EntityID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "detached"})
}

// ListEntityAttachments returns attachments for an entity.
func (h *DocumentHandler) ListEntityAttachments(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "type")
	entityID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entity id")
		return
	}

	attachments, err := h.svc.GetEntityAttachments(r.Context(), entityType, entityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, attachments)
}

// Summary returns high-level metrics for the document storage platform.
func (h *DocumentHandler) Summary(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	summary, err := h.svc.GetSummary(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
