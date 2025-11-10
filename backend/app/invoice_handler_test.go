package app

import (
	"encoding/json"
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAllInvoicesHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	require.NotNil(t, app.handlers.fileStorage)

	adminUser := CreateTestUser(t, app, "admin@example.com", "Admin User", []byte("securepassword"), true, true, true, 0, time.Now())
	nonAdminUser := CreateTestUser(t, app, "user@example.com", "Normal User", []byte("securepassword"), true, false, true, 0, time.Now())

	t.Run("Test List all invoices with empty list", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		req, _ := http.NewRequest("GET", "/api/v1/invoices", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.NotNil(t, result["data"])
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok)
		invoicesRaw, ok := data["invoices"]
		assert.True(t, ok)
		invoicesBytes, err := json.Marshal(invoicesRaw)
		assert.NoError(t, err)
		var invoices []models.Invoice
		err = json.Unmarshal(invoicesBytes, &invoices)
		assert.NoError(t, err)
		assert.Len(t, invoices, 0)
	})

	invoice1 := &models.Invoice{
		UserID:    adminUser.ID,
		Total:     100.0,
		Tax:       10.0,
		CreatedAt: time.Now(),
	}
	invoice2 := &models.Invoice{
		UserID:    nonAdminUser.ID,
		Total:     200.0,
		Tax:       20.0,
		CreatedAt: time.Now(),
	}
	err = app.handlers.db.CreateInvoice(invoice1)
	require.NoError(t, err)
	err = app.handlers.db.CreateInvoice(invoice2)
	require.NoError(t, err)

	pdf1, err := internal.CreateInvoicePDF(*invoice1, *adminUser, app.config.Invoice)
	require.NoError(t, err)
	pdf2, err := internal.CreateInvoicePDF(*invoice2, *nonAdminUser, app.config.Invoice)
	require.NoError(t, err)

	_, err = app.handlers.fileStorage.WriteInvoiceFile(adminUser.ID, invoice1.ID, pdf1)
	require.NoError(t, err)
	_, err = app.handlers.fileStorage.WriteInvoiceFile(nonAdminUser.ID, invoice2.ID, pdf2)
	require.NoError(t, err)

	t.Run("Test List all invoices successfully", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		req, _ := http.NewRequest("GET", "/api/v1/invoices", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Invoices are retrieved successfully", result["message"])
		assert.NotNil(t, result["data"])
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok)
		invoicesRaw, ok := data["invoices"]
		assert.True(t, ok)
		invoicesBytes, err := json.Marshal(invoicesRaw)
		assert.NoError(t, err)
		var invoices []models.Invoice
		err = json.Unmarshal(invoicesBytes, &invoices)
		assert.NoError(t, err)
		assert.Len(t, invoices, 2)
		var found1, found2 bool
		for _, inv := range invoices {
			if inv.UserID == adminUser.ID {
				found1 = true
			}
			if inv.UserID == nonAdminUser.ID {
				found2 = true
			}
		}
		assert.True(t, found1, "Admin's invoice should be in the list")
		assert.True(t, found2, "Normal user's invoice should be in the list")

		storedPDF1, err := app.handlers.fileStorage.ReadInvoiceFile(adminUser.ID, invoice1.ID)
		require.NoError(t, err)
		assert.Equal(t, pdf1, storedPDF1)
		assert.Greater(t, len(storedPDF1), 100)

		storedPDF2, err := app.handlers.fileStorage.ReadInvoiceFile(nonAdminUser.ID, invoice2.ID)
		require.NoError(t, err)
		assert.Equal(t, pdf2, storedPDF2)
		assert.Greater(t, len(storedPDF2), 100)

		assert.NotEqual(t, storedPDF1, storedPDF2)
	})

	t.Run("Test List all invoices with no token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/invoices", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test List all invoices with non-admin user", func(t *testing.T) {
		token := GetAuthToken(t, app, nonAdminUser.ID, nonAdminUser.Email, nonAdminUser.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/invoices", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestListUserInvoicesHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	user := CreateTestUser(t, app, "user@example.com", "Test User", []byte("securepassword"), true, false, true, 0, time.Now())

	t.Run("Test List user invoices with empty list", func(t *testing.T) {
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/invoice", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.NotNil(t, result["data"])
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok)
		invoicesRaw, ok := data["invoices"]
		assert.True(t, ok)
		invoicesBytes, err := json.Marshal(invoicesRaw)
		assert.NoError(t, err)
		var invoices []models.Invoice
		err = json.Unmarshal(invoicesBytes, &invoices)
		assert.NoError(t, err)
		assert.Len(t, invoices, 0)
	})

	invoice1 := &models.Invoice{
		UserID:    user.ID,
		Total:     100.0,
		Tax:       10.0,
		CreatedAt: time.Now(),
	}
	err = app.handlers.db.CreateInvoice(invoice1)
	require.NoError(t, err)

	pdfContent, err := internal.CreateInvoicePDF(*invoice1, *user, app.config.Invoice)
	require.NoError(t, err)
	require.NotEmpty(t, pdfContent)

	_, err = app.handlers.fileStorage.WriteInvoiceFile(user.ID, invoice1.ID, pdfContent)
	require.NoError(t, err)

	t.Run("Test List user invoices successfully", func(t *testing.T) {
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/invoice", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Invoices are retrieved successfully", result["message"])

		data := result["data"].(map[string]interface{})
		invoicesRaw := data["invoices"].([]interface{})
		assert.Len(t, invoicesRaw, 1)

		storedPDF, err := app.handlers.fileStorage.ReadInvoiceFile(user.ID, invoice1.ID)
		require.NoError(t, err)
		assert.Equal(t, pdfContent, storedPDF)
		assert.Greater(t, len(storedPDF), 100)
	})

	t.Run("Test List user invoices with no token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/user/invoice", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

}

func TestDownloadInvoiceHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	require.NotNil(t, app.handlers.fileStorage)

	user1 := CreateTestUser(t, app, "user1@example.com", "User One", []byte("securepassword"), true, false, false, 0, time.Now())

	invoice := &models.Invoice{
		ID:        1,
		UserID:    user1.ID,
		Total:     100.0,
		Tax:       10.0,
		CreatedAt: time.Now(),
	}
	err = app.handlers.db.CreateInvoice(invoice)
	require.NoError(t, err)

	t.Run("Download an invoice successfully", func(t *testing.T) {
		pdfContent, err := internal.CreateInvoicePDF(*invoice, *user1, app.config.Invoice)
		require.NoError(t, err)
		require.NotEmpty(t, pdfContent)
		require.Greater(t, len(pdfContent), 100)

		fileName, err := app.handlers.fileStorage.WriteInvoiceFile(user1.ID, invoice.ID, pdfContent)
		require.NoError(t, err)
		assert.Contains(t, fileName, fmt.Sprintf("user-%d", user1.ID))
		assert.Contains(t, fileName, fmt.Sprintf("invoice-%d", invoice.ID))

		filePath := filepath.Join(app.config.FileStoragePath, "invoices", fileName)
		assert.FileExists(t, filePath)

		fileInfo, err := os.Stat(filePath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())

		token := GetAuthToken(t, app, user1.ID, user1.Email, user1.Username, false)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/user/invoice/%d", invoice.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "application/pdf", resp.Header().Get("Content-Type"))

		responseBody := resp.Body.Bytes()
		assert.NotEmpty(t, responseBody)
		assert.Greater(t, len(responseBody), 100)
		assert.Equal(t, pdfContent, responseBody)

		if len(responseBody) >= 4 {
			assert.Equal(t, "%PDF", string(responseBody[:4]))
		}

		contentDisposition := resp.Header().Get("Content-Disposition")
		assert.Contains(t, contentDisposition, "attachment")
		assert.Contains(t, contentDisposition, fileName)
	})

	t.Run("Download invoice with no token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/user/invoice/%d", invoice.ID), nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Download non-existing invoice", func(t *testing.T) {
		token := GetAuthToken(t, app, user1.ID, user1.Email, user1.Username, false)
		maxID := invoice.ID + 1
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/user/invoice/%d", maxID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Download invoice with invalid invoice id", func(t *testing.T) {
		token := GetAuthToken(t, app, user1.ID, user1.Email, user1.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/invoice/abc", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Download invoice missing from file storage generates on-demand", func(t *testing.T) {
		user2 := CreateTestUser(t, app, "user2@example.com", "User Two", []byte("password"), true, false, true, 0, time.Now())

		invoice2 := &models.Invoice{
			UserID:    user2.ID,
			Total:     99.99,
			Tax:       9.99,
			CreatedAt: time.Now(),
		}
		err := app.db.CreateInvoice(invoice2)
		require.NoError(t, err)

		token := GetAuthToken(t, app, user2.ID, user2.Email, user2.Username, false)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/user/invoice/%d", invoice2.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "application/pdf", resp.Header().Get("Content-Type"))

		responseBody := resp.Body.Bytes()
		assert.NotEmpty(t, responseBody)
		assert.Greater(t, len(responseBody), 100)

		if len(responseBody) >= 4 {
			assert.Equal(t, "%PDF", string(responseBody[:4]))
		}

		storedContent, err := app.handlers.fileStorage.ReadInvoiceFile(user2.ID, invoice2.ID)
		require.NoError(t, err)
		assert.Equal(t, responseBody, storedContent)
	})

	t.Run("User cannot download another user's invoice", func(t *testing.T) {
		user3 := CreateTestUser(t, app, "user3@example.com", "User Three", []byte("password"), true, false, true, 0, time.Now())

		invoice3 := &models.Invoice{
			UserID:    user1.ID,
			Total:     250.00,
			Tax:       25.00,
			CreatedAt: time.Now(),
		}
		err := app.db.CreateInvoice(invoice3)
		require.NoError(t, err)

		pdfContent, err := internal.CreateInvoicePDF(*invoice3, *user1, app.config.Invoice)
		require.NoError(t, err)
		_, err = app.handlers.fileStorage.WriteInvoiceFile(user1.ID, invoice3.ID, pdfContent)
		require.NoError(t, err)

		token := GetAuthToken(t, app, user3.ID, user3.Email, user3.Username, false)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/user/invoice/%d", invoice3.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusForbidden, resp.Code)
		assert.NotEqual(t, "application/pdf", resp.Header().Get("Content-Type"))
	})
}
