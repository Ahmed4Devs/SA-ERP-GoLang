package identifiers

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NewDocumentNumber returns a human-readable, collision-resistant identifier.
// The random component contains 48 bits from a UUID and is safe to generate
// concurrently across API and worker instances without shared process state.
func NewDocumentNumber(prefix string, at time.Time) string {
	cleanPrefix := strings.ToUpper(strings.TrimSpace(prefix))
	if at.IsZero() {
		at = time.Now().UTC()
	}
	random := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:12])
	return fmt.Sprintf("%s-%s-%s", cleanPrefix, at.Format("2006"), random)
}
