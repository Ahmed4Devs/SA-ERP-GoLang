package transaction

import (
	"context"
	"testing"
)

func TestTxContextHelpers(t *testing.T) {
	ctx := context.Background()

	// Initial context should not contain a transaction
	if tx, ok := TxFromContext(ctx); ok || tx != nil {
		t.Errorf("expected no transaction in fresh context, got %v", tx)
	}
}
