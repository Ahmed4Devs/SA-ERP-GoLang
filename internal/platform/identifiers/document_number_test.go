package identifiers

import (
	"strings"
	"testing"
	"time"
)

func TestNewDocumentNumberFormatAndUniqueness(t *testing.T) {
	at := time.Date(2026, time.September, 18, 0, 0, 0, 0, time.UTC)
	seen := make(map[string]struct{}, 10_000)
	for range 10_000 {
		number := NewDocumentNumber(" so ", at)
		if !strings.HasPrefix(number, "SO-2026-") || len(number) != len("SO-2026-")+12 {
			t.Fatalf("unexpected document number format: %q", number)
		}
		if _, exists := seen[number]; exists {
			t.Fatalf("duplicate document number generated: %s", number)
		}
		seen[number] = struct{}{}
	}
}
