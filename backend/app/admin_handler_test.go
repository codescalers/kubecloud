package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kubecloud/models"
)

func TestListUsersHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	adminUser := CreateTestUser(t, app, "admin@example.com", "Admin User", []byte("securepassword"), true, true, false, 0, time.Now())
	normalUser := CreateTestUser(t, app, "user@example.com", "Normal User", []byte("securepassword"), true, false, false, 0, time.Now())

	t.Run("Test List all users successfully", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var usersResp map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &usersResp)
		assert.NoError(t, err)
		assert.Equal(t, "Users are retrieved successfully", usersResp["message"])

		data, ok := usersResp["data"].(map[string]interface{})
		assert.True(t, ok)
		usersRaw, ok := data["users"]
		assert.True(t, ok)
		usersBytes, err := json.Marshal(usersRaw)
		assert.NoError(t, err)

		var users []models.User
		err = json.Unmarshal(usersBytes, &users)
		assert.NoError(t, err)

		var foundAdmin, foundUser bool
		for _, user := range users {
			if user.Email == adminUser.Email {
				foundAdmin = true
			}
			if user.Email == normalUser.Email {
				foundUser = true
			}
		}
		assert.True(t, foundAdmin, "Admin user should be in the list")
		assert.True(t, foundUser, "Normal user should be in the list")
	})

	t.Run("Test List users with no token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test List users with non-admin credentials", func(t *testing.T) {
		token := GetAuthToken(t, app, normalUser.ID, normalUser.Email, normalUser.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestDeleteUsersHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	adminUser := CreateTestUser(t, app, "admin@example.com", "Admin User", []byte("securepassword"), true, true, false, 0, time.Now())
	nonAdminUser := CreateTestUser(t, app, "user@example.com", "Normal User", []byte("securepassword"), true, false, false, 0, time.Now())

	t.Run("Test Delete user successfully", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, adminUser.Admin)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/2", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "User is deleted successfully", result["message"])
	})

	t.Run("Test Admin deletes its account", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, adminUser.Admin)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("Test Delete with invalid user id", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, adminUser.Admin)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/aaa", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test Delete with no user id", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, adminUser.Admin)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Test Delete non-existing user", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, adminUser.Admin)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/9999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Test Delete user with no token", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v1/users/2", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test Delete with non-admin user", func(t *testing.T) {
		token := GetAuthToken(t, app, nonAdminUser.ID, nonAdminUser.Email, nonAdminUser.Username, false)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/2", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestGenerateVouchersHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	adminUser := CreateTestUser(t, app, "admin@example.com", "Admin User", []byte("securepassword"), true, true, false, 0, time.Now())
	nonAdminUser := CreateTestUser(t, app, "user@example.com", "Normal User", []byte("securepassword"), true, false, false, 0, time.Now())

	t.Run("Test GenerateVouchers successfully", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		payload := GenerateVouchersInput{
			Count:       2,
			Value:       10.0,
			ExpireAfter: 7,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/vouchers/generate", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Vouchers are generated successfully", result["message"])
		assert.NotNil(t, result["data"])
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok)
		vouchersRaw, ok := data["vouchers"]
		assert.True(t, ok)
		vouchersBytes, err := json.Marshal(vouchersRaw)
		assert.NoError(t, err)
		var vouchers []models.Voucher
		err = json.Unmarshal(vouchersBytes, &vouchers)
		assert.NoError(t, err)
		assert.Len(t, vouchers, 2)
	})

	t.Run("Test GenerateVouchers with invalid request format", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		body, _ := json.Marshal(map[string]interface{}{})
		req, _ := http.NewRequest("POST", "/api/v1/vouchers/generate", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test GenerateVouchers with no token", func(t *testing.T) {
		payload := GenerateVouchersInput{
			Count:       1,
			Value:       5.0,
			ExpireAfter: 3,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/vouchers/generate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test GenerateVouchers with non-admin user", func(t *testing.T) {
		token := GetAuthToken(t, app, nonAdminUser.ID, nonAdminUser.Email, nonAdminUser.Username, false)
		payload := GenerateVouchersInput{
			Count:       1,
			Value:       5.0,
			ExpireAfter: 3,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/vouchers/generate", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestListVouchersHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	adminUser := CreateTestUser(t, app, "admin@example.com", "Admin User", []byte("securepassword"), true, true, false, 0, time.Now())
	nonAdminUser := CreateTestUser(t, app, "user@example.com", "Normal User", []byte("securepassword"), true, false, false, 0, time.Now())

	voucher1 := &models.Voucher{
		Code:      "VOUCHER1",
		Value:     10.0,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	voucher2 := &models.Voucher{
		Code:      "VOUCHER2",
		Value:     20.0,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
	err = app.handlers.db.CreateVoucher(voucher1)
	require.NoError(t, err)
	err = app.handlers.db.CreateVoucher(voucher2)
	require.NoError(t, err)

	t.Run("Test List Vouchers successfully", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		req, _ := http.NewRequest("GET", "/api/v1/vouchers", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		var vouchersResp map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &vouchersResp)
		assert.NoError(t, err)
		assert.Equal(t, "Vouchers are Retrieved successfully", vouchersResp["message"])
		data, ok := vouchersResp["data"].(map[string]interface{})
		assert.True(t, ok)
		vouchersRaw, ok := data["vouchers"]
		assert.True(t, ok)
		vouchersBytes, err := json.Marshal(vouchersRaw)
		assert.NoError(t, err)
		var vouchers []models.Voucher
		err = json.Unmarshal(vouchersBytes, &vouchers)
		assert.NoError(t, err)
		var found1, found2 bool
		for _, v := range vouchers {
			if v.Code == voucher1.Code {
				found1 = true
			}
			if v.Code == voucher2.Code {
				found2 = true
			}
		}
		assert.True(t, found1, "Voucher1 should be in the list")
		assert.True(t, found2, "Voucher2 should be in the list")
	})

	t.Run("Test ListVouchersHandler with no token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/vouchers", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test ListVouchersHandler with non-admin user", func(t *testing.T) {
		token := GetAuthToken(t, app, nonAdminUser.ID, nonAdminUser.Email, nonAdminUser.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/vouchers", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestCreditUserHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	adminUser := CreateTestUser(t, app, "admin@example.com", "Admin User", []byte("securepassword"), true, true, true, 0, time.Now())
	normalUser := CreateTestUser(t, app, "user@example.com", "Normal User", []byte("securepassword"), true, false, true, 0, time.Now())

	t.Run("Test Credit user successfully", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		payload := CreditRequestInput{
			AmountUSD: 1,
			Memo:      "Manual credit",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/users/2/credit", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "User is credited successfully", result["message"])
		assert.NotNil(t, result["data"])
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, normalUser.Email, data["user"])
		assert.EqualValues(t, 1, data["amount"])
		assert.Equal(t, "Manual credit", data["memo"])
	})

	t.Run("Test Credit user with invalid request format", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		body, _ := json.Marshal(map[string]interface{}{}) // missing required fields
		req, _ := http.NewRequest("POST", "/api/v1/users/2/credit", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test Credit user with invalid user id", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		payload := CreditRequestInput{
			AmountUSD: 1,
			Memo:      "Manual credit",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/users/abc/credit", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test Credit non-existing user", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		payload := CreditRequestInput{
			AmountUSD: 1,
			Memo:      "Manual credit",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/users/9999/credit", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("Test Credit user with no token", func(t *testing.T) {
		payload := CreditRequestInput{
			AmountUSD: 1,
			Memo:      "Manual credit",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/users/2/credit", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test Credit user with non-admin user", func(t *testing.T) {
		token := GetAuthToken(t, app, normalUser.ID, normalUser.Email, normalUser.Username, false)
		payload := CreditRequestInput{
			AmountUSD: 1,
			Memo:      "Manual credit",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/users/2/credit", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestListPendingRecordsHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	adminUser := CreateTestUser(t, app, "admin@example.com", "Admin User", []byte("securepassword"), true, true, false, 0, time.Now())
	nonAdminUser := CreateTestUser(t, app, "user@example.com", "Normal User", []byte("securepassword"), true, false, false, 0, time.Now())

	t.Run("Test ListPendingRecordsHandler successfully", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		req, _ := http.NewRequest("GET", "/api/v1/pending-records", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("Test ListPendingRecordsHandler with no token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/pending-records", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test ListPendingRecordsHandler with non-admin user", func(t *testing.T) {
		token := GetAuthToken(t, app, nonAdminUser.ID, nonAdminUser.Email, nonAdminUser.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/pending-records", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("Test ListPendingRecordsHandler with non-existing user", func(t *testing.T) {
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/pending-records/%d", nonAdminUser.ID+1), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

}

func TestSendMailToAllUsersHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	adminUser := CreateTestUser(t, app, "admin@example.com", "Admin User", []byte("securepassword"), true, true, true, 0, time.Now())
	normalUser := CreateTestUser(t, app, "user@example.com", "Normal User", []byte("securepassword"), true, false, false, 0, time.Now())

	t.Run("Test Send email with non-admin user", func(t *testing.T) {
		body, writer := createMultipartEmailForm(t, "Test Subject", "Test email body")
		token := GetAuthToken(t, app, normalUser.ID, normalUser.Email, normalUser.Username, false)

		req, _ := http.NewRequest("POST", "/api/v1/users/mail", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("Test Send email validates concurrency handling", func(t *testing.T) {
		for i := 0; i < 25; i++ {
			CreateTestUser(t, app, fmt.Sprintf("testuser%d@example.com", i), fmt.Sprintf("Test User %d", i), []byte("securepassword"), true, false, false, 0, time.Now())
		}

		body, writer := createMultipartEmailForm(t, "Concurrency Test", "Testing concurrent email delivery")
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)

		req, _ := http.NewRequest("POST", "/api/v1/users/mail", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		result := extractSendMailResponse(t, resp.Body)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, 27, result.TotalUsers)
		assert.Equal(t, 27, result.SuccessfulEmails)
		assert.Equal(t, 0, result.FailedEmailsCount)
	})

	t.Run("Test Send email with partial success - some emails fail", func(t *testing.T) {
		app, err := SetUp(t)
		require.NoError(t, err)
		router := app.router

		adminUser := CreateTestUser(t, app, "admin@example.com", "Admin User", []byte("securepassword"), true, true, true, 0, time.Now())
		body, writer := createMultipartEmailForm(t, "Partial Success Test", "Testing partial email delivery success")
		token := GetAuthToken(t, app, adminUser.ID, adminUser.Email, adminUser.Username, true)

		CreateTestUser(t, app, "invalid-email", "Invalid User", []byte("password"), true, false, false, 0, time.Now())

		req, _ := http.NewRequest("POST", "/api/v1/users/mail", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		result := extractSendMailResponse(t, resp.Body)
		assert.Equal(t, 2, result.TotalUsers)
		assert.Equal(t, 1, result.SuccessfulEmails)
		assert.Equal(t, 1, result.FailedEmailsCount)
	})

}

func createMultipartEmailForm(t *testing.T, subject, body string) (*bytes.Buffer, *multipart.Writer) {
	t.Helper()
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	err := writer.WriteField("subject", subject)
	require.NoError(t, err)

	err = writer.WriteField("body", body)
	require.NoError(t, err)

	writer.Close()
	return &buffer, writer
}

func extractSendMailResponse(t *testing.T, responseBody *bytes.Buffer) SendMailResponse {
	t.Helper()
	var apiResponse APIResponse
	err := json.Unmarshal(responseBody.Bytes(), &apiResponse)
	require.NoError(t, err)
	resultBytes, err := json.Marshal(apiResponse.Data)
	require.NoError(t, err)

	var result SendMailResponse
	err = json.Unmarshal(resultBytes, &result)
	require.NoError(t, err)

	return result
}
