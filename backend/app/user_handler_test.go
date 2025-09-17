package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	t.Run("Register User Successfully", func(t *testing.T) {

		payload := RegisterInput{
			Name:            "Test User",
			Email:           "testuser@example.com",
			Password:        "securepassword",
			ConfirmPassword: "securepassword",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equalf(t, http.StatusAccepted, resp.Code, "Expected status %d, got %d. Body: %s", http.StatusAccepted, resp.Code, resp.Body.String())

	})

	t.Run("Register User with Invalid Request Format", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{})

		req, _ := http.NewRequest("POST", "/api/v1/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, resp.Code, http.StatusBadRequest)

	})

	t.Run("Register Existing Verified User", func(t *testing.T) {
		user := CreateTestUser(t, app, "dupe@example.com", "Test User", []byte("securepassword"), true, false, false, 0, time.Now())
		user.Mnemonic = "mnemonic"
		user.AccountAddress = "sponseeAddress"
		user.Sponsored = true
		user.StripeCustomerID = "stripeCustomerID"
		require.NoError(t, app.db.UpdateUserByID(user))

		payload := RegisterInput{
			Name:            "New Name",
			Email:           user.Email,
			Password:        "newpassword",
			ConfirmPassword: "newpassword",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusConflict, resp.Code)

	})

	t.Run("Register Existing Not Verified User", func(t *testing.T) {
		user := CreateTestUser(t, app, "unverified@example.com", "Unverified User", []byte("securepassword"), false, false, false, 0, time.Now())

		payload := RegisterInput{
			Name:            "New Name",
			Email:           user.Email,
			Password:        "newpassword",
			ConfirmPassword: "newpassword",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusAccepted, resp.Code)
	})
}

func TestVerifyRegisterCode(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	user := CreateTestUser(t, app, "dupe@example.com", "Test User", []byte("securepassword"), false, false, false, 123, time.Now())
	t.Run("Test Verify Register Code", func(t *testing.T) {
		payload := VerifyCodeInput{
			Email: user.Email,
			Code:  user.Code,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusAccepted, resp.Code)

	})

	t.Run("Test Verify Register Code with Invalid request format", func(t *testing.T) {
		payload := VerifyCodeInput{
			Email: "dupe@example.com",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["message"], "Invalid request format")

	})
	t.Run("Test Verify Register Code with registered user", func(t *testing.T) {
		registeredUser := CreateTestUser(t, app, "registered@example.com", "Registered User", []byte("securepassword"), true, false, false, 123, time.Now())
		registeredUser.Mnemonic = "mnemonic"
		registeredUser.AccountAddress = "sponseeAddress"
		registeredUser.Sponsored = true
		registeredUser.StripeCustomerID = "stripeCustomerID"
		require.NoError(t, app.db.UpdateUserByID(registeredUser))

		payload := VerifyCodeInput{
			Email: registeredUser.Email,
			Code:  123,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusConflict, resp.Code)

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["error"], "User is already registered")
	})

	t.Run("Test Verify Register Code with wrong code", func(t *testing.T) {
		user2 := CreateTestUser(t, app, "dupe2@example.com", "Test User2", []byte("securepassword"), false, false, false, 123, time.Now())

		payload := VerifyCodeInput{
			Email: user2.Email,
			Code:  333,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["error"], "Invalid verification code")

	})

	t.Run("Test Verify Register Code with expired code", func(t *testing.T) {
		user2 := CreateTestUser(t, app, "test@example.com", "Test User", []byte("securepassword"), false, false, false, 123, time.Now().Add(-2*time.Hour))

		payload := VerifyCodeInput{
			Email: user2.Email,
			Code:  user2.Code,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/user/register/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["error"], "code has expired")

	})

}

func TestLoginUserHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router

	hashedPassword, _ := internal.HashAndSaltPassword([]byte("securepassword"))
	user := CreateTestUser(t, app, "loginuser@example.com", "Login User", hashedPassword, true, false, true, 0, time.Now())

	t.Run("Test LoginUserHandler with Invalid Request Format", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{"email": "abc"})
		req, _ := http.NewRequest("POST", "/api/v1/user/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test LoginUserHandler with non-existing user", func(t *testing.T) {

		payload := LoginInput{
			Email:    "notfound@example.com",
			Password: "irrelevant",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["error"], "email or password is incorrect")
	})
	t.Run("Test LoginUserHandler with wrong password", func(t *testing.T) {
		payload := LoginInput{
			Email:    user.Email,
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["error"], "email or password is incorrect")
	})
}

func TestRefreshTokenHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	t.Run("Test RefreshTokenHandler", func(t *testing.T) {

		user := CreateTestUser(t, app, "refreshtoken@example.com", "Refresh User", []byte("securepassword"), true, false, false, 0, time.Now())
		tokenPair, _ := app.handlers.tokenManager.CreateTokenPair(user.ID, user.Username, false)

		payload := RefreshTokenInput{
			RefreshToken: tokenPair.RefreshToken,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "access token refreshed successfully", result["message"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Test RefreshTokenHandler with Invalid Request Format", func(t *testing.T) {

		body, _ := json.Marshal(map[string]interface{}{})
		req, _ := http.NewRequest("POST", "/api/v1/user/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test RefreshTokenHandler with Invalid or Expired Token", func(t *testing.T) {

		payload := RefreshTokenInput{
			RefreshToken: "invalidtoken",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["error"], "Invalid or expired refresh token")
	})
}

func TestForgotPasswordHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	t.Run("Test ForgotPasswordHandler", func(t *testing.T) {

		user := CreateTestUser(t, app, "forgotuser@example.com", "Forgot User", []byte("securepassword"), true, false, false, 0, time.Now())

		payload := EmailInput{
			Email: user.Email,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Verification code sent", result["message"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Test ForgotPasswordHandler with Invalid Request format", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{})
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test ForgotPasswordHandler with non-existing user", func(t *testing.T) {

		payload := EmailInput{
			Email: "notfound@example.com",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["error"], "failed to get user")
	})

}

func TestVerifyForgetPasswordCodeHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	t.Run("Test VerifyForgetPasswordCodeHandler", func(t *testing.T) {

		user := CreateTestUser(t, app, "resetuser@example.com", "Reset User", []byte("securepassword"), false, false, false, 4231, time.Now())

		payload := VerifyCodeInput{
			Email: user.Email,
			Code:  user.Code,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Verification successful", result["message"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Test VerifyForgetPasswordCodeHandler with Invalid request format", func(t *testing.T) {

		body, _ := json.Marshal(map[string]interface{}{})
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test VerifyForgetPasswordCodeHandler with wrong code", func(t *testing.T) {

		user := CreateTestUser(t, app, "wrongreset@example.com", "Wrong Reset", []byte("securepassword"), false, false, false, 0, time.Now())

		assert.NoError(t, err)
		payload := VerifyCodeInput{
			Email: user.Email,
			Code:  9999,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["message"], "Invalid code")
	})

	t.Run("Test VerifyForgetPasswordCodeHandler with expired code", func(t *testing.T) {

		user := CreateTestUser(t, app, "expiredreset@example.com", "Expired Reset", []byte("securepassword"), false, false, false, 4231, time.Now().Add(-2*time.Hour))

		payload := VerifyCodeInput{
			Email: user.Email,
			Code:  user.Code,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["error"], "code has expired")
	})

	t.Run("Test VerifyForgetPasswordCodeHandler with non-existing user", func(t *testing.T) {

		payload := VerifyCodeInput{
			Email: "notfoundreset@example.com",
			Code:  4321,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/forgot_password/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["error"], "record not found")
	})
}

func TestChangePasswordHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	user := CreateTestUser(t, app, "changepass@example.com", "Change Pass", []byte("oldpassword"), true, false, false, 0, time.Now())

	t.Run("Test ChangePasswordHandler", func(t *testing.T) {
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)

		payload := ChangePasswordInput{
			Email:           user.Email,
			Password:        "newsecurepassword",
			ConfirmPassword: "newsecurepassword",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/api/v1/user/change_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusAccepted, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "password is updated successfully", result["message"])
	})

	t.Run("Test ChangePasswordHandler with Invalid Request format", func(t *testing.T) {
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		body, _ := json.Marshal(map[string]interface{}{})
		req, _ := http.NewRequest("PUT", "/api/v1/user/change_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test ChangePasswordHandler with passwords mismatch", func(t *testing.T) {
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		payload := ChangePasswordInput{
			Email:           user.Email,
			Password:        "newsecurepassword",
			ConfirmPassword: "differentpassword",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/api/v1/user/change_password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

func TestChargeBalanceHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	t.Run("Test ChargeBalance with Invalid Request format", func(t *testing.T) {

		user := CreateTestUser(t, app, "chargeuser@example.com", "Charge User", []byte("securepassword"), true, false, true, 0, time.Now())
		user.Mnemonic = "test-menmonic"
		err = app.handlers.db.UpdateUserByID(user)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		payload := ChargeBalanceInput{
			CardType: "visa",
			Amount:   10,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/balance/charge", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test ChargeBalance with invalid amount", func(t *testing.T) {

		email := "chargeuser3@example.com"
		username := "Charge User3"
		token := GetAuthToken(t, app, 1, email, username, false)
		payload := ChargeBalanceInput{
			CardType:     "visa",
			PaymentToken: "tok_test",
			Amount:       0,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/balance/charge", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test ChargeBalance with non-existing user", func(t *testing.T) {

		token := GetAuthToken(t, app, 1, "notfound@example.com", "Not Found", false)
		payload := ChargeBalanceInput{
			CardType:     "visa",
			PaymentToken: "tok_test",
			Amount:       100,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/charge_balance", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})
}

func TestGetUserHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	t.Run("Test Get user successfully", func(t *testing.T) {

		user := CreateTestUser(t, app, "getuser@example.com", "Get User", []byte("securepassword"), true, false, false, 0, time.Now())
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "User is retrieved successfully", result["message"])
		assert.NotNil(t, result["data"])
		userData := result["data"].(map[string]interface{})["user"].(map[string]interface{})
		assert.Equal(t, user.Email, userData["email"])
		assert.Equal(t, user.Username, userData["username"])
	})

	t.Run("Test Get non-existing user", func(t *testing.T) {

		token := GetAuthToken(t, app, 999, "notfound@example.com", "Not Found", false)
		req, _ := http.NewRequest("GET", "/api/v1/user/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		var result map[string]interface{}
		_ = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Contains(t, result["message"], "User is not found")
	})

}

func TestGetUserBalanceHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	t.Run("Test Get balance successfully", func(t *testing.T) {

		user := CreateTestUser(t, app, "balanceuser@example.com", "Balance User", []byte("securepassword"), true, false, true, 0, time.Now())

		assert.NoError(t, err)
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/balance", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Balance is fetched", result["message"])
		assert.NotNil(t, result["data"])
		data := result["data"].(map[string]interface{})
		assert.Contains(t, data, "balance_usd")
		assert.Contains(t, data, "debt_usd")
	})

	t.Run("Test Get balance for non-existing user", func(t *testing.T) {

		token := GetAuthToken(t, app, 999, "notfound@example.com", "Not Found", false)
		req, _ := http.NewRequest("GET", "/api/v1/user/balance", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["message"], "User is not found")
	})

}

func TestRedeemVoucherHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	user := CreateTestUser(t, app, "voucheruser@example.com", "Voucher User", []byte("securepassword"), true, false, true, 0, time.Now())

	voucher := &models.Voucher{
		ID:        1,
		Code:      "VOUCHER123",
		Value:     50.0,
		Redeemed:  false,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	err = app.handlers.db.CreateVoucher(voucher)
	assert.NoError(t, err)
	t.Run("Test redeem voucher successfully", func(t *testing.T) {

		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("PUT", "/api/v1/user/redeem/VOUCHER123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusAccepted, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Voucher is redeemed successfully. Money transfer in progress.", result["message"])
		assert.NotNil(t, result["data"])
		assert.NotEmpty(t, result["data"].(map[string]interface{})["workflow_id"])
	})

	t.Run("Test redeem non-existing voucher", func(t *testing.T) {
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("PUT", "/api/v1/user/redeem/Voucher123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Test redeem already redeemed voucher", func(t *testing.T) {
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/user/redeem/%s", voucher.Code), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test redeem expired voucher", func(t *testing.T) {
		voucher := &models.Voucher{
			ID:        3,
			Code:      "EXPIREDVOUCHER",
			Value:     20.0,
			Redeemed:  false,
			CreatedAt: time.Now().Add(-2 * time.Hour),
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		err = app.handlers.db.CreateVoucher(voucher)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("PUT", "/api/v1/user/redeem/EXPIREDVOUCHER", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Test redeem voucher for non-existing user", func(t *testing.T) {
		token := GetAuthToken(t, app, 999, "notfound@example.com", "Not Found", false)
		req, _ := http.NewRequest("PUT", "/api/v1/user/redeem/VOUCHER123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Test redeem voucher with missing code param", func(t *testing.T) {

		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("PUT", "/api/v1/user/redeem/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.True(t, resp.Code == http.StatusNotFound)
	})

}

func TestListSSHKeysHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	user := CreateTestUser(t, app, "sshuser@example.com", "SSH User", []byte("securepassword"), true, false, false, 0, time.Now())
	sshKey1 := &models.SSHKey{
		UserID:    user.ID,
		Name:      "key1",
		PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC1",
	}
	sshKey2 := &models.SSHKey{
		UserID:    user.ID,
		Name:      "key2",
		PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC2",
	}
	t.Run("Test list SSH keys with no keys", func(t *testing.T) {
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/ssh-keys", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "SSH keys retrieved successfully", result["message"])
		assert.NotNil(t, result["data"])
		sshKeys := result["data"].([]interface{})
		assert.Len(t, sshKeys, 0)
	})

	t.Run("Test list SSH keys with multiple keys", func(t *testing.T) {

		err = app.handlers.db.CreateSSHKey(sshKey1)
		assert.NoError(t, err)
		err = app.handlers.db.CreateSSHKey(sshKey2)
		assert.NoError(t, err)
		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		req, _ := http.NewRequest("GET", "/api/v1/user/ssh-keys", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "SSH keys retrieved successfully", result["message"])
		assert.NotNil(t, result["data"])
		sshKeys := result["data"].([]interface{})
		assert.Len(t, sshKeys, 2)
	})
}

func TestAddSSHKeyHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	t.Run("Add SSH key successfully", func(t *testing.T) {

		user := CreateTestUser(t, app, "addsshuser@example.com", "Add SSH User", []byte("securepassword"), true, false, false, 0, time.Now())

		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		payload := SSHKeyInput{
			Name:      "mykey",
			PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDzy9yGz+CsKhjYB3FLr27SaoPQVi/tOZDZ06LnO7NuVUj0yR3e7IJO26cxs6j7tRAGTrA7choRMlQJdCFQfkDCaAL+31fPSihHhB3kxUTnZymaWgZ6s/JxjI/2/kKcLjxMWpMYTs18ZdRJf1DgoiyTV6yhlxAhWJvMxTtC5++h5+Ir7mHoN5QdrRt5AjKEcTEJjoKC3it4itHz7w45hi4y07kFYIk4HcMGrInh1IC/BriU7xKlwYcP2tp0W4GIraDJoD8OR3cgcYd/AFXSnVDtomCq5MaKBUli6FWLCK7E3+0AtYxxLkQ/zFkPsYSFAGGqVp8uq2hI46d0TxhgcG2EsWiF/2yOjtMdX1ab3Ns23p8Q0l/8JxXn6WT9xhme9eb2v8UjukN0AR8j+hp5xoQuSEgXAxkg4PFEa2seYEcE8xZPOSavuQl4wEAjXH/1BHnRHxrBBWixN2xdclHRAKQRwR+EHg8wDQ0EAAxtoCCAVHOepBrmV0JDxJGHQ8euvbs= test@gmail.com",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/ssh-keys", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusCreated, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "SSH key added successfully", result["message"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Add SSH key with invalid request format", func(t *testing.T) {

		user := CreateTestUser(t, app, "addsshuser2@example.com", "Add SSH User2", []byte("securepassword"), true, false, false, 0, time.Now())

		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		// Missing public_key
		payload := SSHKeyInput{
			Name: "mykey2",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/ssh-keys", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Add SSH key with invalid SSH key format", func(t *testing.T) {

		user := CreateTestUser(t, app, "addsshuser3@example.com", "Add SSH User3", []byte("securepassword"), true, false, false, 0, time.Now())

		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		payload := SSHKeyInput{
			Name:      "badkey",
			PublicKey: "not-a-valid-ssh-key",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/user/ssh-keys", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result["error"], "invalid SSH key format")
	})

	t.Run("Add SSH key with duplicate public key", func(t *testing.T) {

		user := CreateTestUser(t, app, "addsshuser4@example.com", "Add SSH User4", []byte("securepassword"), true, false, false, 0, time.Now())

		token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
		publicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDzy9yGz+CsKhjYB3FLr27SaoPQVi/tOZDZ06LnO7NuVUj0yR3e7IJO26cxs6j7tRAGTrA7choRMlQJdCFQfkDCaAL+31fPSihHhB3kxUTnZymaWgZ6s/JxjI/2/kKcLjxMWpMYTs18ZdRJf1DgoiyTV6yhlxAhWJvMxTtC5++h5+Ir7mHoN5QdrRt5AjKEcTEJjoKC3it4itHz7w45hi4y07kFYIk4HcMGrInh1IC/BriU7xKlwYcP2tp0W4GIraDJoD8OR3cgcYd/AFXSnVDtomCq5MaKBUli6FWLCK7E3+0AtYxxLkQ/zFkPsYSFAGGqVp8uq2hI46d0TxhgcG2EsWiF/2yOjtMdX1ab3Ns23p8Q0l/8JxXn6WT9xhme9eb2v8UjukN0AR8j+hp5xoQuSEgXAxkg4PFEa2seYEcE8xZPOSavuQl4wEAjXH/1BHnRHxrBBWixN2xdclHRAKQRwR+EHg8wDQ0EAAxtoCCAVHOepBrmV0JDxJGHQ8euvbs= test@gmail.com"

		// Add first SSH key
		payload1 := SSHKeyInput{
			Name:      "keyA",
			PublicKey: publicKey,
		}
		body1, _ := json.Marshal(payload1)
		req1, _ := http.NewRequest("POST", "/api/v1/user/ssh-keys", bytes.NewReader(body1))
		req1.Header.Set("Authorization", "Bearer "+token)
		req1.Header.Set("Content-Type", "application/json")
		resp1 := httptest.NewRecorder()
		router.ServeHTTP(resp1, req1)
		assert.Equal(t, http.StatusCreated, resp1.Code)

		payload2 := SSHKeyInput{
			Name:      "keyB",
			PublicKey: publicKey,
		}
		body2, _ := json.Marshal(payload2)
		req2, _ := http.NewRequest("POST", "/api/v1/user/ssh-keys", bytes.NewReader(body2))
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", "application/json")
		resp2 := httptest.NewRecorder()
		router.ServeHTTP(resp2, req2)
		assert.Equal(t, http.StatusBadRequest, resp2.Code)
	})

}

func TestListUserPendingRecordsHandler(t *testing.T) {
	app, err := SetUp(t)
	require.NoError(t, err)
	router := app.router
	user := CreateTestUser(t, app, "pendinguser@example.com", "Pending User", []byte("securepassword"), true, false, false, 0, time.Now())
	token := GetAuthToken(t, app, user.ID, user.Email, user.Username, false)
	t.Run("Test list user pending records successfully", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/user/pending-records", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Pending records are retrieved successfully", result["message"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Test list user pending records with no records", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/user/pending-records", nil)
		assert.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Pending records are retrieved successfully", result["message"])
		assert.NotNil(t, result["data"])
	})

	t.Run("Test list user pending records with no token", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/user/pending-records", nil)
		assert.NoError(t, err)

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Test list user pending records with invalid token", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/user/pending-records", nil)
		assert.NoError(t, err)

		req.Header.Set("Authorization", "Bearer invalidtoken")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

}
