package path

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandPath_EmptyPath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	result, err := ExpandPath("")
	if err != nil {
		t.Errorf("Expected no error for empty path, got: %v", err)
	}

	if result != wd {
		t.Errorf("Expected working directory %s, got %s", wd, result)
	}
}

func TestExpandPath_AbsolutePath(t *testing.T) {
	absPath := "/tmp/test/path"
	result, err := ExpandPath(absPath)
	if err != nil {
		t.Errorf("Expected no error for absolute path, got: %v", err)
	}

	expected, err := filepath.Abs(absPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestExpandPath_RelativePath(t *testing.T) {
	relPath := "test/path"
	result, err := ExpandPath(relPath)
	if err != nil {
		t.Errorf("Expected no error for relative path, got: %v", err)
	}

	expected, err := filepath.Abs(relPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestExpandPath_TildeHome(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Cannot get user home directory: %v", err)
	}

	result, err := ExpandPath("~")
	if err != nil {
		t.Errorf("Expected no error for ~, got: %v", err)
	}

	if result != homeDir {
		t.Errorf("Expected home directory %s, got %s", homeDir, result)
	}
}

func TestExpandPath_TildeHomeWithPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Cannot get user home directory: %v", err)
	}

	result, err := ExpandPath("~/test/file.txt")
	if err != nil {
		t.Errorf("Expected no error for ~/test/file.txt, got: %v", err)
	}

	expected := filepath.Join(homeDir, "test", "file.txt")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestExpandPath_EnvironmentVariables(t *testing.T) {
	// Set a test environment variable
	testValue := "/tmp/test/env/path"
	os.Setenv("TEST_PATH", testValue)
	defer os.Unsetenv("TEST_PATH")

	result, err := ExpandPath("$TEST_PATH/subdir")
	if err != nil {
		t.Errorf("Expected no error for environment variable, got: %v", err)
	}

	expected := filepath.Join(testValue, "subdir")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestExpandPath_InvalidUser(t *testing.T) {
	_, err := ExpandPath("~nonexistentuser/path")
	if err == nil {
		t.Error("Expected error for nonexistent user, got nil")
	}

	if !strings.Contains(err.Error(), "user nonexistentuser not found") {
		t.Errorf("Expected user not found error, got: %v", err)
	}
}

func TestExpandPath_CleanPath(t *testing.T) {
	// Test that the path is cleaned (removes redundant separators, etc.)
	input := "/tmp//test/../test/path/"
	result, err := ExpandPath(input)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := "/tmp/test/path"
	if result != expected {
		t.Errorf("Expected cleaned path %s, got %s", expected, result)
	}
}

func TestExpandTilde_OnlyTilde(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Cannot get user home directory: %v", err)
	}

	result, err := expandTilde("~")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result != homeDir {
		t.Errorf("Expected %s, got %s", homeDir, result)
	}
}

func TestExpandTilde_WithPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Cannot get user home directory: %v", err)
	}

	result, err := expandTilde("~/subdir/file.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := filepath.Join(homeDir, "subdir", "file.txt")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
