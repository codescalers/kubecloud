package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	InvoicesDir    = "invoices"
	KubeconfigsDir = "kubeconfigs"
)

type FileStorageService struct {
	baseDir string
	mu      sync.Mutex
}

func NewFileStorageService(baseDir string) (*FileStorageService, error) {
	s := &FileStorageService{baseDir: baseDir}
	if err := s.ensureFileStorageSubdirs(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileStorageService) ensureFileStorageSubdirs() error {
	if strings.TrimSpace(s.baseDir) == "" {
		return fmt.Errorf("file storage base directory cannot be empty")
	}
	if err := os.MkdirAll(filepath.Join(s.baseDir, InvoicesDir), 0o700); err != nil {
		return fmt.Errorf("failed to create invoices directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(s.baseDir, KubeconfigsDir), 0o700); err != nil {
		return fmt.Errorf("failed to create kubeconfigs directory: %w", err)
	}
	return nil
}

func (s *FileStorageService) InvoiceFileName(userID, invoiceID int) string {
	return fmt.Sprintf("user-%d-invoice-%d.pdf", userID, invoiceID)
}

func (s *FileStorageService) kubeconfigFileName(userID, clusterID int, projectName string) string {
	safeProject := strings.ReplaceAll(strings.TrimSpace(projectName), " ", "-")
	return fmt.Sprintf("user-%d-cluster-%d-%s-kubeconfig.yaml", userID, clusterID, safeProject)
}

func (s *FileStorageService) buildPath(dir string, fileName string) string {
	return filepath.Join(s.baseDir, dir, fileName)
}

func (s *FileStorageService) WriteInvoiceFile(userID, invoiceID int, data []byte) (string, error) {
	fileName := s.InvoiceFileName(userID, invoiceID)
	absPath := s.buildPath(InvoicesDir, fileName)
	if err := s.atomicWriteFile(absPath, data, 0o600); err != nil {
		return "", err
	}
	return fileName, nil
}

func (s *FileStorageService) ReadInvoiceFile(userID, invoiceID int) ([]byte, error) {
	fileName := s.InvoiceFileName(userID, invoiceID)
	absPath := s.buildPath(InvoicesDir, fileName)
	return os.ReadFile(absPath)
}

func (s *FileStorageService) WriteKubeconfigFile(userID, clusterID int, projectName string, data []byte) (string, error) {
	fileName := s.kubeconfigFileName(userID, clusterID, projectName)
	absPath := s.buildPath(KubeconfigsDir, fileName)
	if err := s.atomicWriteFile(absPath, data, 0o600); err != nil {
		return "", err
	}
	return fileName, nil
}

func (s *FileStorageService) ReadKubeconfigFile(userID, clusterID int, projectName string) ([]byte, error) {
	fileName := s.kubeconfigFileName(userID, clusterID, projectName)
	absPath := s.buildPath(KubeconfigsDir, fileName)
	return os.ReadFile(absPath)
}

func (s *FileStorageService) DeleteKubeconfigFile(userID, clusterID int, projectName string) error {
	fileName := s.kubeconfigFileName(userID, clusterID, projectName)
	absPath := s.buildPath(KubeconfigsDir, fileName)
	if err := os.Remove(absPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func (s *FileStorageService) atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(path)

	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tmpFile = nil

	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
