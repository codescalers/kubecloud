package utils

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandPath(t *testing.T) {
	// Get current user for test cases
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		// Empty path
		{
			name:     "empty path",
			input:    "",
			expected: cwd,
			hasError: false,
		},
		// Bare tilde
		{
			name:     "bare tilde",
			input:    "~",
			expected: currentUser.HomeDir,
			hasError: false,
		},
		// Tilde with slash (current user)
		{
			name:     "tilde with nested path",
			input:    "~/documents/projects/test",
			expected: filepath.Join(currentUser.HomeDir, "documents", "projects", "test"),
			hasError: false,
		},
		// Tilde with username
		{
			name:     "tilde with current username",
			input:    "~" + currentUser.Username,
			expected: currentUser.HomeDir,
			hasError: false,
		},
		// Root user (should exist on most systems)
		{
			name:     "tilde with root username",
			input:    "~root",
			expected: "/root",
			hasError: false,
		},
		{
			name:     "tilde with root username and path",
			input:    "~root/.ssh",
			expected: "/root/.ssh",
			hasError: false,
		},
		// Non-existent user
		{
			name:     "tilde with non-existent user",
			input:    "~nonexistentuser12345",
			expected: "",
			hasError: true,
		},
		{
			name:     "tilde with invalid username characters",
			input:    "~user@invalid",
			expected: "",
			hasError: true,
		},
		{
			name:     "tilde with username containing spaces",
			input:    "~user name",
			expected: "",
			hasError: true,
		},
		{
			name:     "tilde with username containing slashes",
			input:    "~user/name",
			expected: "",
			hasError: true,
		},
		// Absolute paths (should not be modified)
		{
			name:     "absolute path",
			input:    "/absolute/path",
			expected: "/absolute/path",
			hasError: false,
		},
		{
			name:     "absolute path with tilde in middle",
			input:    "/path/with~tilde",
			expected: "/path/with~tilde",
			hasError: false,
		},
		// Relative paths (should not be modified)
		{
			name:     "relative path",
			input:    "relative/path",
			expected: filepath.Join(cwd, "relative", "path"),
			hasError: false,
		},
		{
			name:     "relative path with tilde",
			input:    "path~with~tildes",
			expected: filepath.Join(cwd, "path~with~tildes"),
			hasError: false,
		},
		// Environment variable expansion
		{
			name:     "path with environment variable",
			input:    "$HOME/documents",
			expected: filepath.Join(currentUser.HomeDir, "documents"),
			hasError: false,
		},
		// Edge cases
		{
			name:     "multiple slashes",
			input:    "~/documents//projects///test",
			expected: filepath.Join(currentUser.HomeDir, "documents", "projects", "test"),
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandPath(tt.input)

			if tt.hasError {
				assert.Error(t, err, "expandPath(%q) expected error, got nil", tt.input)
				return
			}

			assert.NoError(t, err, "expandPath(%q) unexpected error: %v", tt.input, err)

			// Clean the result for comparison
			expected := filepath.Clean(tt.expected)
			assert.Equal(t, expected, result, "expandPath(%q)", tt.input)
		})
	}
}
