package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileStorageService(t *testing.T) {
	t.Run("Create service successfully", func(t *testing.T) {
		tmpDir := t.TempDir()

		service, err := NewFileStorageService(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		invoicesDir := filepath.Join(tmpDir, InvoicesDir)
		kubeconfigsDir := filepath.Join(tmpDir, KubeconfigsDir)

		assert.DirExists(t, invoicesDir)
		assert.DirExists(t, kubeconfigsDir)

		invoiceInfo, err := os.Stat(invoicesDir)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), invoiceInfo.Mode().Perm())

		kubeconfigInfo, err := os.Stat(kubeconfigsDir)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), kubeconfigInfo.Mode().Perm())
	})

	t.Run("Fail with empty base directory", func(t *testing.T) {
		_, err := NewFileStorageService("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file storage base directory cannot be empty")
	})

	t.Run("Fail with whitespace-only base directory", func(t *testing.T) {
		_, err := NewFileStorageService("   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file storage base directory cannot be empty")
	})

	t.Run("Create service with nested directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "nested", "path", "to", "storage")

		service, err := NewFileStorageService(nestedDir)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.DirExists(t, filepath.Join(nestedDir, InvoicesDir))
		assert.DirExists(t, filepath.Join(nestedDir, KubeconfigsDir))
	})
}

func TestWriteAndReadInvoiceFile(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewFileStorageService(tmpDir)
	assert.NoError(t, err)

	t.Run("Write and read invoice file successfully", func(t *testing.T) {
		userID := 1
		invoiceID := 100
		testData := []byte("This is a test PDF content for invoice")

		fileName, err := service.WriteInvoiceFile(userID, invoiceID, testData)
		assert.NoError(t, err)
		assert.Equal(t, "user-1-invoice-100.pdf", fileName)

		filePath := filepath.Join(tmpDir, InvoicesDir, fileName)
		assert.FileExists(t, filePath)

		fileInfo, err := os.Stat(filePath)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())

		readData, err := service.ReadInvoiceFile(userID, invoiceID)
		assert.NoError(t, err)
		assert.Equal(t, testData, readData)
	})

	t.Run("Overwrite existing invoice file", func(t *testing.T) {
		userID := 2
		invoiceID := 200
		originalData := []byte("Original invoice data")
		updatedData := []byte("Updated invoice data - much longer content")

		_, err := service.WriteInvoiceFile(userID, invoiceID, originalData)
		assert.NoError(t, err)

		_, err = service.WriteInvoiceFile(userID, invoiceID, updatedData)
		assert.NoError(t, err)
		readData, err := service.ReadInvoiceFile(userID, invoiceID)
		assert.NoError(t, err)
		assert.Equal(t, updatedData, readData)
	})

	t.Run("Write empty invoice file", func(t *testing.T) {
		userID := 3
		invoiceID := 300
		emptyData := []byte{}

		fileName, err := service.WriteInvoiceFile(userID, invoiceID, emptyData)
		assert.NoError(t, err)

		readData, err := service.ReadInvoiceFile(userID, invoiceID)
		assert.NoError(t, err)
		assert.Equal(t, emptyData, readData)

		filePath := filepath.Join(tmpDir, InvoicesDir, fileName)
		assert.FileExists(t, filePath)
	})

	t.Run("Read non-existent invoice file", func(t *testing.T) {
		userID := 999
		invoiceID := 999

		_, err := service.ReadInvoiceFile(userID, invoiceID)
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestWriteAndReadKubeconfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewFileStorageService(tmpDir)
	assert.NoError(t, err)

	t.Run("Write and read kubeconfig file successfully", func(t *testing.T) {
		userID := 1
		clusterID := 10
		projectName := "test-cluster"
		kubeconfigData := []byte(`apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://example.com
  name: test-cluster
`)

		fileName, err := service.WriteKubeconfigFile(userID, clusterID, projectName, kubeconfigData)
		assert.NoError(t, err)
		assert.Equal(t, "user-1-cluster-10-test-cluster-kubeconfig.yaml", fileName)

		filePath := filepath.Join(tmpDir, KubeconfigsDir, fileName)
		assert.FileExists(t, filePath)

		fileInfo, err := os.Stat(filePath)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())

		readData, err := service.ReadKubeconfigFile(userID, clusterID, projectName)
		assert.NoError(t, err)
		assert.Equal(t, kubeconfigData, readData)
	})

	t.Run("Write kubeconfig for project with spaces in name", func(t *testing.T) {
		userID := 2
		clusterID := 20
		projectName := "my cluster 2"
		kubeconfigData := []byte("kubeconfig content")

		fileName, err := service.WriteKubeconfigFile(userID, clusterID, projectName, kubeconfigData)
		assert.NoError(t, err)
		assert.Contains(t, fileName, "my-cluster-2")

		readData, err := service.ReadKubeconfigFile(userID, clusterID, projectName)
		assert.NoError(t, err)
		assert.Equal(t, kubeconfigData, readData)
	})

	t.Run("Read non-existent kubeconfig file", func(t *testing.T) {
		userID := 999
		clusterID := 999
		projectName := "non-existent"

		_, err := service.ReadKubeconfigFile(userID, clusterID, projectName)
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestDeleteKubeconfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewFileStorageService(tmpDir)
	assert.NoError(t, err)

	t.Run("Delete existing kubeconfig file", func(t *testing.T) {
		userID := 1
		clusterID := 10
		projectName := "test-cluster"
		kubeconfigData := []byte("test kubeconfig")

		fileName, err := service.WriteKubeconfigFile(userID, clusterID, projectName, kubeconfigData)
		assert.NoError(t, err)

		filePath := filepath.Join(tmpDir, KubeconfigsDir, fileName)
		assert.FileExists(t, filePath)

		err = service.DeleteKubeconfigFile(userID, clusterID, projectName)
		assert.NoError(t, err)

		assert.NoFileExists(t, filePath)
	})

	t.Run("Delete non-existent kubeconfig file returns no error", func(t *testing.T) {
		userID := 999
		clusterID := 999
		projectName := "non-existent"

		err := service.DeleteKubeconfigFile(userID, clusterID, projectName)
		assert.NoError(t, err)
	})

	t.Run("Delete same file twice", func(t *testing.T) {
		userID := 2
		clusterID := 20
		projectName := "test-cluster-2"
		kubeconfigData := []byte("test kubeconfig 2")

		_, err := service.WriteKubeconfigFile(userID, clusterID, projectName, kubeconfigData)
		assert.NoError(t, err)

		err = service.DeleteKubeconfigFile(userID, clusterID, projectName)
		assert.NoError(t, err)

		err = service.DeleteKubeconfigFile(userID, clusterID, projectName)
		assert.NoError(t, err)
	})
}
