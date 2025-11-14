package persistence

import (
	"github.com/lib/pq"
)

// isUniqueViolation checks if an error is due to a database unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	if pgErr, ok := err.(*pq.Error); ok {
		return pgErr.Code == "23505"
	}

	return false
}
