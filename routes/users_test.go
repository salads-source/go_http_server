package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salads-source/go_http_server/db"
	"github.com/stretchr/testify/assert"
)

func TestSignup(t *testing.T) {
	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name: "successful signup",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]string{
				"message": "User created successfully",
			},
		},
		{
			name: "missing email",
			payload: map[string]string{
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]string{
				"message": "Could not parse request",
			},
		},
		{
			name: "missing password",
			payload: map[string]string{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]string{
				"message": "Could not parse request",
			},
		},
		{
			name: "empty payload",
			payload: map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]string{
				"message": "Could not parse request",
			},
		},
		{
			name: "duplicate email",
			payload: map[string]string{
				"email":    "duplicate@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]string{
				"message": "Could not save user, try again later",
			},
		},
	}

	_, router := setupTestRouter(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create duplicate user for duplicate email test
			if tt.name == "duplicate email" {
				createTestUser(t, db.DB, "duplicate@example.com", "password123")
			}

			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]string
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, tt.expectedBody["message"], response["message"])
		})
	}
}

func TestLogin(t *testing.T) {
	_, router := setupTestRouter(t)

	// Create a test user
	testEmail := "login@example.com"
	testPassword := "password123"
	createTestUser(t, db.DB, testEmail, testPassword)

	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
		checkToken     bool
	}{
		{
			name: "successful login",
			payload: map[string]string{
				"email":    testEmail,
				"password": testPassword,
			},
			expectedStatus: http.StatusCreated,
			checkToken:     true,
		},
		{
			name: "invalid email",
			payload: map[string]string{
				"email":    "nonexistent@example.com",
				"password": testPassword,
			},
			expectedStatus: http.StatusUnauthorized,
			checkToken:     false,
		},
		{
			name: "invalid password",
			payload: map[string]string{
				"email":    testEmail,
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			checkToken:     false,
		},
		{
			name: "missing email",
			payload: map[string]string{
				"password": testPassword,
			},
			expectedStatus: http.StatusBadRequest,
			checkToken:     false,
		},
		{
			name: "missing password",
			payload: map[string]string{
				"email": testEmail,
			},
			expectedStatus: http.StatusBadRequest,
			checkToken:     false,
		},
		{
			name: "empty payload",
			payload: map[string]string{},
			expectedStatus: http.StatusBadRequest,
			checkToken:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.checkToken {
				assert.Contains(t, response, "token")
				assert.NotEmpty(t, response["token"])
				assert.Equal(t, "Login Successfully", response["message"])
			} else if tt.expectedStatus == http.StatusBadRequest {
				assert.Equal(t, "Could not parse request", response["message"])
			} else {
				assert.Contains(t, response, "message")
			}
		})
	}
}

