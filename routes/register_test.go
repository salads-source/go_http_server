package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/salads-source/go_http_server/db"
	"github.com/stretchr/testify/assert"
)

func TestGetRegistrations(t *testing.T) {
	_, router := setupTestRouter(t)

	tests := []struct {
		name           string
		expectedStatus int
		setupRegs      bool
	}{
		{
			name:           "get all registrations - empty",
			expectedStatus: http.StatusOK,
			setupRegs:      false,
		},
		{
			name:           "get all registrations - with data",
			expectedStatus: http.StatusOK,
			setupRegs:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupRegs {
				userId := createTestUser(t, db.DB, "reguser@example.com", "password123")
				eventResult, err := db.DB.Exec(
					"INSERT INTO events(name, description, location, dateTime, user_id) VALUES (?, ?, ?, ?, ?)",
					"Test Event", "Test Description", "Test Location", time.Now(), userId,
				)
				if err != nil {
					t.Fatalf("Failed to create test event: %v", err)
				}
				eventID, _ := eventResult.LastInsertId()

				_, err = db.DB.Exec(
					"INSERT INTO registrations(event_id, user_id) VALUES (?, ?)",
					eventID, userId,
				)
				if err != nil {
					t.Fatalf("Failed to create test registration: %v", err)
				}
			}

			req, _ := http.NewRequest(http.MethodGet, "/registrations", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.setupRegs {
				assert.Equal(t, "Registrations fetched successfully", response["message"])
				registrations, ok := response["registrations"].([]interface{})
				assert.True(t, ok)
				assert.Greater(t, len(registrations), 0)
			} else {
				registrations, ok := response["registrations"].([]interface{})
				if ok {
					assert.Equal(t, 0, len(registrations))
				} else {
					// If registrations is nil, that's also acceptable for empty case
					assert.Nil(t, response["registrations"])
				}
			}
		})
	}
}

func TestGetRegistration(t *testing.T) {
	_, router := setupTestRouter(t)
	userId := createTestUser(t, db.DB, "reguser@example.com", "password123")

	tests := []struct {
		name           string
		registrationID string
		expectedStatus int
		setupReg       bool
	}{
		{
			name:           "get registration by id - success",
			registrationID: "1",
			expectedStatus: http.StatusOK,
			setupReg:       true,
		},
		{
			name:           "get registration by id - invalid id",
			registrationID: "invalid",
			expectedStatus: http.StatusBadRequest,
			setupReg:       false,
		},
		{
			name:           "get registration by id - not found",
			registrationID: "999",
			expectedStatus: http.StatusInternalServerError,
			setupReg:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupReg {
				eventResult, err := db.DB.Exec(
					"INSERT INTO events(name, description, location, dateTime, user_id) VALUES (?, ?, ?, ?, ?)",
					"Test Event", "Test Description", "Test Location", time.Now(), userId,
				)
				if err != nil {
					t.Fatalf("Failed to create test event: %v", err)
				}
				eventID, _ := eventResult.LastInsertId()

				_, err = db.DB.Exec(
					"INSERT INTO registrations(event_id, user_id) VALUES (?, ?)",
					eventID, userId,
				)
				if err != nil {
					t.Fatalf("Failed to create test registration: %v", err)
				}
			}

			req, _ := http.NewRequest(http.MethodGet, "/registrations/"+tt.registrationID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var registration map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &registration)
				assert.Contains(t, registration, "ID")
				assert.Contains(t, registration, "EventID")
				assert.Contains(t, registration, "UserID")
			}
		})
	}
}

func TestRegisterForEvent(t *testing.T) {
	_, router := setupTestRouter(t)
	userId := createTestUser(t, db.DB, "registeruser@example.com", "password123")
	eventOwnerId := createTestUser(t, db.DB, "eventowner@example.com", "password123")
	token := generateTestToken(t, "registeruser@example.com", userId)

	tests := []struct {
		name           string
		eventID        string
		authToken      string
		expectedStatus int
		setupEvent     bool
	}{
		{
			name:           "successful registration",
			eventID:        "1",
			authToken:      token,
			expectedStatus: http.StatusCreated,
			setupEvent:     true,
		},
		{
			name:           "missing authorization",
			eventID:        "1",
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
			setupEvent:     true,
		},
		{
			name:           "invalid token",
			eventID:        "1",
			authToken:      "invalid_token",
			expectedStatus: http.StatusUnauthorized,
			setupEvent:     true,
		},
		{
			name:           "invalid event id",
			eventID:        "invalid",
			authToken:      token,
			expectedStatus: http.StatusBadRequest,
			setupEvent:     false,
		},
		{
			name:           "event not found",
			eventID:        "999",
			authToken:      token,
			expectedStatus: http.StatusInternalServerError,
			setupEvent:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualEventID string = tt.eventID
			if tt.setupEvent {
				result, err := db.DB.Exec(
					"INSERT INTO events(name, description, location, dateTime, user_id) VALUES (?, ?, ?, ?, ?)",
					"Test Event", "Test Description", "Test Location", time.Now(), eventOwnerId,
				)
				if err != nil {
					t.Fatalf("Failed to create test event: %v", err)
				}
				insertedID, _ := result.LastInsertId()
				// Use the actual inserted ID if the test specifies "1"
				if tt.eventID == "1" {
					actualEventID = strconv.FormatInt(insertedID, 10)
				}
			}

			req, _ := http.NewRequest(http.MethodPost, "/events/"+actualEventID+"/register", nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", tt.authToken)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "User registered for event successfully", response["message"])
			}
		})
	}
}

func TestCancelRegistration(t *testing.T) {
	_, router := setupTestRouter(t)
	userId := createTestUser(t, db.DB, "canceluser@example.com", "password123")
	eventOwnerId := createTestUser(t, db.DB, "eventowner2@example.com", "password123")
	token := generateTestToken(t, "canceluser@example.com", userId)

	tests := []struct {
		name           string
		eventID        string
		authToken      string
		expectedStatus int
		setupEvent     bool
		setupReg       bool
	}{
		{
			name:           "successful cancellation",
			eventID:        "1",
			authToken:      token,
			expectedStatus: http.StatusOK,
			setupEvent:     true,
			setupReg:       true,
		},
		{
			name:           "missing authorization",
			eventID:        "1",
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
			setupEvent:     true,
			setupReg:       true,
		},
		{
			name:           "invalid token",
			eventID:        "1",
			authToken:      "invalid_token",
			expectedStatus: http.StatusUnauthorized,
			setupEvent:     true,
			setupReg:       true,
		},
		{
			name:           "invalid event id",
			eventID:        "invalid",
			authToken:      token,
			expectedStatus: http.StatusBadRequest,
			setupEvent:     false,
			setupReg:       false,
		},
		{
			name:           "no registration to cancel",
			eventID:        "1",
			authToken:      token,
			expectedStatus: http.StatusOK,
			setupEvent:     true,
			setupReg:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualEventID string = tt.eventID
			var eventID int64
			if tt.setupEvent {
				result, err := db.DB.Exec(
					"INSERT INTO events(name, description, location, dateTime, user_id) VALUES (?, ?, ?, ?, ?)",
					"Test Event", "Test Description", "Test Location", time.Now(), eventOwnerId,
				)
				if err != nil {
					t.Fatalf("Failed to create test event: %v", err)
				}
				eventID, _ = result.LastInsertId()
				// Use the actual inserted ID if the test specifies "1"
				if tt.eventID == "1" {
					actualEventID = strconv.FormatInt(eventID, 10)
				}
			}

			if tt.setupReg {
				_, err := db.DB.Exec(
					"INSERT INTO registrations(event_id, user_id) VALUES (?, ?)",
					eventID, userId,
				)
				if err != nil {
					t.Fatalf("Failed to create test registration: %v", err)
				}
			}

			req, _ := http.NewRequest(http.MethodDelete, "/events/"+actualEventID+"/register", nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", tt.authToken)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Registration for event cancelled successfully", response["message"])
			}
		})
	}
}

