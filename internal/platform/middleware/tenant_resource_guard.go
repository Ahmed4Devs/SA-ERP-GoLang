package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"erp/internal/platform/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const maxTenantResourceIDs = 100

// TenantOwnershipChecker reports whether a UUID belongs to a tenant-scoped row.
// A UUID with no known owner is allowed because it may identify a global resource.
type TenantOwnershipChecker interface {
	Ownership(ctx context.Context, resourceID, requestedTenantID uuid.UUID) (ownedByRequestedTenant, ownedByOtherTenant bool, err error)
}

// PostgresTenantOwnershipChecker discovers tenant-scoped tables from the schema.
// The generated lookup uses only catalog-provided, quoted identifiers.
type PostgresTenantOwnershipChecker struct {
	db       *database.DB
	once     sync.Once
	query    string
	queryErr error
}

func NewPostgresTenantOwnershipChecker(db *database.DB) *PostgresTenantOwnershipChecker {
	return &PostgresTenantOwnershipChecker{db: db}
}

func (c *PostgresTenantOwnershipChecker) Ownership(ctx context.Context, resourceID, requestedTenantID uuid.UUID) (bool, bool, error) {
	c.once.Do(func() { c.query, c.queryErr = c.buildQuery(ctx) })
	if c.queryErr != nil {
		return false, false, c.queryErr
	}
	var requested, other bool
	if err := c.db.Pool.QueryRow(ctx, c.query, resourceID, requestedTenantID).Scan(&requested, &other); err != nil {
		return false, false, fmt.Errorf("resolve tenant resource ownership: %w", err)
	}
	return requested, other, nil
}

func (c *PostgresTenantOwnershipChecker) buildQuery(ctx context.Context) (string, error) {
	rows, err := c.db.Pool.Query(ctx, `
		SELECT n.nspname, cls.relname
		FROM pg_class cls
		JOIN pg_namespace n ON n.oid = cls.relnamespace
		WHERE cls.relkind IN ('r', 'p')
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND EXISTS (
		      SELECT 1 FROM pg_attribute a
		      WHERE a.attrelid = cls.oid AND a.attname = 'id'
		        AND a.atttypid = 'uuid'::regtype AND NOT a.attisdropped
		  )
		  AND EXISTS (
		      SELECT 1 FROM pg_attribute a
		      WHERE a.attrelid = cls.oid AND a.attname = 'legal_entity_id'
		        AND a.atttypid = 'uuid'::regtype AND NOT a.attisdropped
		  )
		ORDER BY n.nspname, cls.relname`)
	if err != nil {
		return "", fmt.Errorf("discover tenant-scoped tables: %w", err)
	}
	defer rows.Close()

	parts := []string{
		`SELECT id AS resource_id, id AS legal_entity_id FROM organization.legal_entities WHERE id = $1`,
	}
	for rows.Next() {
		var schema, table string
		if err := rows.Scan(&schema, &table); err != nil {
			return "", fmt.Errorf("scan tenant-scoped table: %w", err)
		}
		identifier := pgx.Identifier{schema, table}.Sanitize()
		parts = append(parts, fmt.Sprintf("SELECT id AS resource_id, legal_entity_id FROM %s WHERE id = $1", identifier))
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate tenant-scoped tables: %w", err)
	}

	owners := strings.Join(parts, " UNION ALL ")
	return fmt.Sprintf(`
		WITH owned AS MATERIALIZED (%s)
		SELECT
			EXISTS (SELECT 1 FROM owned WHERE legal_entity_id = $2),
			EXISTS (SELECT 1 FROM owned WHERE legal_entity_id <> $2)`, owners), nil
}

// TenantResourceGuard rejects path, query, and JSON foreign-key UUIDs that are
// known to belong to a different legal entity before the handler can read or mutate them.
func TenantResourceGuard(checker TenantOwnershipChecker, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, ok := TenantFromContext(r.Context())
			if !ok || checker == nil {
				next.ServeHTTP(w, r)
				return
			}

			ids, err := tenantResourceIDs(r, tenantID)
			if err != nil {
				writeTenantError(w, http.StatusBadRequest, err.Error())
				return
			}
			for _, resourceID := range ids {
				ownedHere, ownedElsewhere, err := checker.Ownership(r.Context(), resourceID, tenantID)
				if err != nil {
					if logger != nil {
						logger.Error("Tenant ownership check failed", "error", err, "resource_id", resourceID, "legal_entity_id", tenantID)
					}
					writeTenantError(w, http.StatusServiceUnavailable, "Tenant ownership check unavailable")
					return
				}
				if !ownedHere && ownedElsewhere {
					writeTenantError(w, http.StatusNotFound, "Resource not found")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func tenantResourceIDs(r *http.Request, tenantID uuid.UUID) ([]uuid.UUID, error) {
	unique := make(map[uuid.UUID]struct{})
	add := func(candidate string) error {
		id, err := uuid.Parse(strings.TrimSpace(candidate))
		if err != nil || id == tenantID {
			return nil
		}
		unique[id] = struct{}{}
		if len(unique) > maxTenantResourceIDs {
			return fmt.Errorf("too many tenant resource identifiers; maximum is %d", maxTenantResourceIDs)
		}
		return nil
	}

	for _, segment := range strings.Split(strings.Trim(r.URL.Path, "/"), "/") {
		if err := add(segment); err != nil {
			return nil, err
		}
	}
	for key, values := range r.URL.Query() {
		if key == "legal_entity_id" || (!strings.HasSuffix(key, "_id") && key != "id") {
			continue
		}
		for _, value := range values {
			if err := add(value); err != nil {
				return nil, err
			}
		}
	}

	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.HasPrefix(contentType, "application/json") && r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, errorsForTenantBodyRead()
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		if len(body) > 0 {
			var value any
			if json.Unmarshal(body, &value) == nil {
				if err := collectTenantResourceIDs(value, "", tenantID, add); err != nil {
					return nil, err
				}
			}
		}
	}

	ids := make([]uuid.UUID, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}
	return ids, nil
}

func collectTenantResourceIDs(value any, key string, tenantID uuid.UUID, add func(string) error) error {
	switch typed := value.(type) {
	case map[string]any:
		for childKey, child := range typed {
			if err := collectTenantResourceIDs(child, childKey, tenantID, add); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range typed {
			if err := collectTenantResourceIDs(child, key, tenantID, add); err != nil {
				return err
			}
		}
	case string:
		if key != "legal_entity_id" && (key == "id" || strings.HasSuffix(key, "_id")) {
			return add(typed)
		}
	}
	return nil
}

func errorsForTenantBodyRead() error {
	return fmt.Errorf("unable to read request body for tenant ownership validation")
}
