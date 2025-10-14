package models

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLParse_DatabaseDSNs(t *testing.T) {
	tests := []struct {
		name       string
		dsn        string
		wantScheme string
		wantHost   string
		wantPath   string
		wantOpaque string
	}{
		{
			name:       "postgres standard",
			dsn:        "postgres://user:pass@localhost:5432/dbname?sslmode=disable&timezone=UTC",
			wantScheme: "postgres",
			wantHost:   "localhost:5432",
			wantPath:   "/dbname",
		},
		{
			name:       "sqlite absolute path 3 slashes",
			dsn:        "sqlite3:///absolute/path.db",
			wantScheme: "sqlite3",
			wantHost:   "",
			wantPath:   "/absolute/path.db",
		},
		{
			name:       "sqlite host only treated as relative base",
			dsn:        "sqlite3://relative.db",
			wantScheme: "sqlite3",
			wantHost:   "relative.db",
			wantPath:   "",
		},
		{
			name:       "sqlite single slash absolute",
			dsn:        "sqlite:/absolute/path.db",
			wantScheme: "sqlite",
			wantHost:   "",
			wantPath:   "/absolute/path.db",
		},
		{
			name:       "sqlite absolute path 3 slashes (sqlite)",
			dsn:        "sqlite:///just/abs.db",
			wantScheme: "sqlite",
			wantHost:   "",
			wantPath:   "/just/abs.db",
		},
		{
			name:       "sqlite in-memory opaque",
			dsn:        "sqlite3::memory:",
			wantScheme: "sqlite3",
			wantOpaque: ":memory:",
		},
		{
			name:       "mysql standard",
			dsn:        "mysql://user:pass@localhost:3306/dbname?charset=utf8mb4",
			wantScheme: "mysql",
			wantHost:   "localhost:3306",
			wantPath:   "/dbname",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.dsn)
			assert.NoError(t, err, "url.Parse(%q) error", tc.dsn)
			assert.Equal(t, tc.wantScheme, u.Scheme, "scheme")
			assert.Equal(t, tc.wantHost, u.Host, "host")
			assert.Equal(t, tc.wantPath, u.Path, "path")
			assert.Equal(t, tc.wantOpaque, u.Opaque, "opaque")
		})
	}
}
