package storage

import (
	"fmt"

	guuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// uuidToString formats a pgtype.UUID as a standard string.
func uuidToString(u pgtype.UUID) string {
	return guuid.UUID(u.Bytes).String()
}

// parseUUID converts a string to pgtype.UUID.
func parseUUID(id string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid ID %q: %w", id, err)
	}
	return uuid, nil
}
