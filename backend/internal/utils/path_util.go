package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func ExpandPath(path string) (string, error) {
	if path == "" {
		return os.Getwd()
	}

	path = os.ExpandEnv(path)
	var err error
	if strings.HasPrefix(path, "~") {
		path, err = expandTilde(path)
		if err != nil {
			return "", fmt.Errorf("failed to expand tilde: %w", err)
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return filepath.Clean(absPath), nil
}

func expandTilde(path string) (string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if path == "~" {
		return homeDir, nil
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:]), nil
	}
	parts := strings.SplitN(path, "/", 2)
	username := parts[0][1:]

	user, err := user.Lookup(username)
	if err != nil {
		return "", fmt.Errorf("user %s not found: %w", username, err)
	}

	if len(parts) == 1 {
		return user.HomeDir, nil
	}
	return filepath.Join(user.HomeDir, parts[1]), nil

}
