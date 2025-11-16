package models

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound         = errors.New("user is not found")
	ErrSSHKeyNotFound       = errors.New("ssh key is not found")
	ErrSSHKeyAlreadyExists  = errors.New("ssh key already exists")
	ErrClusterNotFound      = errors.New("cluster is not found")
	ErrUserNodeNotFound     = errors.New("user node is not found")
	ErrNotificationNotFound = errors.New("notification is not found")
)

func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	if pgErr, ok := err.(*pq.Error); ok {
		// 23505 is the code for unique constraint violation
		return pgErr.Code == "23505"
	}

	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) {
		// 23505 is the code for unique constraint violation
		return pgxErr.Code == "23505"
	}

	var sqlLiteErr sqlite3.Error
	if errors.As(err, &sqlLiteErr) {
		if sqlLiteErr.Code == sqlite3.ErrConstraint {
			return sqlLiteErr.ExtendedCode == sqlite3.ErrConstraintUnique || sqlLiteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey
		}
	}
	return false
}
