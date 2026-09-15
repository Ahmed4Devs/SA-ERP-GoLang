package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const GenesisHash = "GENESIS"

// CanonicalizeJSON produces deterministic JSON with sorted keys for consistent hashing.
func CanonicalizeJSON(data []byte) []byte {
	if len(data) == 0 {
		return []byte("{}")
	}
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return data
	}
	canonical, err := json.Marshal(raw)
	if err != nil {
		return data
	}
	return canonical
}

// ComputeRecordHash calculates the deterministic SHA-256 hash for an audit record.
func ComputeRecordHash(
	previousHash string,
	userID *uuid.UUID,
	legalEntityID *uuid.UUID,
	action string,
	resourceType string,
	resourceID string,
	payload []byte,
	occurredAt time.Time,
) string {
	if previousHash == "" {
		previousHash = GenesisHash
	}

	userStr := "none"
	if userID != nil {
		userStr = userID.String()
	}

	leStr := "none"
	if legalEntityID != nil {
		leStr = legalEntityID.String()
	}

	// Canonicalize payload JSON representation
	canonicalPayload := CanonicalizeJSON(payload)

	// PostgreSQL TIMESTAMPTZ has microsecond resolution
	normTime := occurredAt.UTC().Truncate(time.Microsecond)
	timeStr := normTime.Format(time.RFC3339Nano)

	// Chain content: prevHash|user|legalEntity|action|resourceType|resourceID|payload|timestamp
	content := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s",
		previousHash,
		userStr,
		leStr,
		action,
		resourceType,
		resourceID,
		string(canonicalPayload),
		timeStr,
	)

	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// ValidateRecordHash checks if an AuditEvent's RecordHash matches its contents.
func ValidateRecordHash(event *AuditEvent) bool {
	expected := ComputeRecordHash(
		event.PreviousHash,
		event.UserID,
		event.LegalEntityID,
		event.Action,
		event.ResourceType,
		event.ResourceID,
		event.Payload,
		event.OccurredAt,
	)
	return expected == event.RecordHash
}
