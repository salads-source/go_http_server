package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/salads-source/go_http_server/db"
	"github.com/stretchr/testify/assert"
)

func TestGetEvents(t *testing.T) {
	_, router := setupTestRouter(t)

	tests := []struct {
		name           string
		expectedStatus int
		setupEvents    bool
	}{
		{
			name:           "get all events - empty",
			expectedStatus: http.StatusOK,
			setupEvents:    false,
		},
		{
			name:           "get all events - with data",
			expectedStatus: http.StatusOK,
			setupEvents:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEvents {
				userId := createTestUser(t, db.DB, "eventuser@example.com", "password123")
				_, err := db.DB.Exec(
					"INSERT INTO events(name, description, location, dateTime, user_id) VALUES (?, ?, ?, ?, ?)",
					"Test Event", "Test Description", "Test Location", time.Now(), userId,
				)
				if err != nil {
					t.Fatalf("Failed to create test event: %v", err)
				}
			}

			req, _ := http.NewRequest(http.MethodGet, "/events", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.setupEvents {
				var events []interface{}
				json.Unmarshal(w.Body.Bytes(), &events)
				assert.Greater(t, len(events), 0)
			} else {
				var events []interface{}
				json.Unmarshal(w.Body.Bytes(), &events)
				assert.Equal(t, 0, len(events))
			}
		})
	}
}

func TestGetEvent(t *testing.T) {
	_, router := setupTestRouter(t)
	userId := createTestUser(t, db.DB, "eventuser@example.com", "password123")

	tests := []struct {
		name           string
		eventID        string
		expectedStatus int
		setupEvent     bool
	}{
		{
			name:           "get event by id - success",
			eventID:        "1",
			expectedStatus: http.StatusOK,
			setupEvent:     true,
		},
		{
			name:           "get event by id - invalid id",
			eventID:        "invalid",
			expectedStatus: http.StatusBadRequest,
			setupEvent:     false,
		},
		{
			name:           "get event by id - not found",
			eventID:        "999",
			expectedStatus: http.StatusInternalServerError,
			setupEvent:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEvent {
				_, err := db.DB.Exec(
					"INSERT INTO events(name, description, location, dateTime, user_id) VALUES (?, ?, ?, ?, ?)",
					"Test Event", "Test Description", "Test Location", time.Now(), userId,
				)
				if err != nil {
					t.Fatalf("Failed to create test event: %v", err)
				}
			}

			req, _ := http.NewRequest(http.MethodGet, "/events/"+tt.eventID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var event map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &event)
				assert.Contains(t, event, "ID")
				assert.Contains(t, event, "Name")
			}
		})
	}
}

func TestCreateEvent(t *testing.T) {
	_, router := setupTestRouter(t)
	userId := createTestUser(t, db.DB, "eventcreator@example.com", "password123")
	token := generateTestToken(t, "eventcreator@example.com", userId)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		authToken      string
		expectedStatus int
	}{
		{
			name: "successful event creation",
			payload: map[string]interface{}{
				"name":        "New Event",
				"description": "Event Description",
				"location":    "Event Location",
				"dateTime":    time.Now().Format(time.RFC3339),
			},
			authToken:      token,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing authorization",
			payload: map[string]interface{}{
				"name":        "New Event",
				"description": "Event Description",
				"location":    "Event Location",
				"dateTime":    time.Now().Format(time.RFC3339),
			},
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid token",
			payload: map[string]interface{}{
				"name":        "New Event",
				"description": "Event Description",
				"location":    "Event Location",
				"dateTime":    time.Now().Format(time.RFC3339),
			},
			authToken:      "invalid_token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "missing name",
			payload: map[string]interface{}{
				"description": "Event Description",
				"location":    "Event Location",
				"dateTime":    time.Now().Format(time.RFC3339),
			},
			authToken:      token,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing description",
			payload: map[string]interface{}{
				"name":     "New Event",
				"location": "Event Location",
				"dateTime": time.Now().Format(time.RFC3339),
			},
			authToken:      token,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing location",
			payload: map[string]interface{}{
				"name":        "New Event",
				"description": "Event Description",
				"dateTime":    time.Now().Format(time.RFC3339),
			},
			authToken:      token,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing dateTime",
			payload: map[string]interface{}{
				"name":        "New Event",
				"description": "Event Description",
				"location":    "Event Location",
			},
			authToken:      token,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			if tt.authToken != "" {
				req.Header.Set("Authorization", tt.authToken)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Event created!", response["message"])
				assert.Contains(t, response, "event")
			}
		})
	}
}

func TestUpdateEvent(t *testing.T) {
	_, router := setupTestRouter(t)
	userId := createTestUser(t, db.DB, "eventupdater@example.com", "password123")
	otherUserId := createTestUser(t, db.DB, "otheruser@example.com", "password123")
	token := generateTestToken(t, "eventupdater@example.com", userId)
	otherToken := generateTestToken(t, "otheruser@example.com", otherUserId)

	tests := []struct {
		name           string
		eventID        string
		payload        map[string]interface{}
		authToken      string
		expectedStatus int
		setupEvent     bool
		eventOwnerID   int64
	}{
		{
			name:    "successful event update",
			eventID: "1",
			payload: map[string]interface{}{
				"name":        "Updated Event",
				"description": "Updated Description",
				"location":    "Updated Location",
				"dateTime":    time.Now().Format(time.RFC3339),
			},
			authToken:      token,
			expectedStatus: http.StatusOK,
			setupEvent:     true,
			eventOwnerID:   userId,
		},
		{
			name:    "unauthorized - not owner",
			eventID: "1",
			payload: map[string]interface{}{
				"name":        "Updated Event",
				"description": "Updated Description",
				"location":    "Updated Location",
				"dateTime":    time.Now().Format(time.RFC3339),
			},
			authToken:      otherToken,
			expectedStatus: http.StatusUnauthorized,
			setupEvent:     true,
			eventOwnerID:   userId,
		},
		{
			name:    "missing authorization",
			eventID: "1",
			payload: map[string]interface{}{
				"name":        "Updated Event",
				"description": "Updated Description",
				"location":    "Updated Location",
				"dateTime":    time.Now().Format(time.RFC3339),
			},
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
			setupEvent:     true,
			eventOwnerID:   userId,
		},
		{
			name:    "invalid event id",
			eventID: "invalid",
			payload: map[string]interface{}{
				"name":        "Updated Event",
				"description": "Updated Description",
				"location":    "Updated Location",
				"dateTime":    time.Now().Format(time.RFC3339),
			},
			authToken:      token,
			expectedStatus: http.StatusBadRequest,
			setupEvent:     false,
		},
		{
			name:    "event not found",
			eventID: "999",
			payload: map[string]interface{}{
				"name":        "Updated Event",
				"description": "Updated Description",
				"location":    "Updated Location",
				"dateTime":    time.Now().Format(time.RFC3339),
			},
			authToken:      token,
			expectedStatus: http.StatusInternalServerError,
			setupEvent:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEvent {
				_, err := db.DB.Exec(
					"INSERT INTO events(name, description, location, dateTime, user_id) VALUES (?, ?, ?, ?, ?)",
					"Original Event", "Original Description", "Original Location", time.Now(), tt.eventOwnerID,
				)
				if err != nil {
					t.Fatalf("Failed to create test event: %v", err)
				}
			}

			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPut, "/events/"+tt.eventID, bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			if tt.authToken != "" {
				req.Header.Set("Authorization", tt.authToken)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Event updated successfully!", response["message"])
			}
		})
	}
}

func TestDeleteEvent(t *testing.T) {
	_, router := setupTestRouter(t)
	userId := createTestUser(t, db.DB, "eventdeleter@example.com", "password123")
	otherUserId := createTestUser(t, db.DB, "otheruser2@example.com", "password123")
	token := generateTestToken(t, "eventdeleter@example.com", userId)
	otherToken := generateTestToken(t, "otheruser2@example.com", otherUserId)

	tests := []struct {
		name           string
		eventID        string
		authToken      string
		expectedStatus int
		setupEvent     bool
		eventOwnerID   int64
	}{
		{
			name:           "successful event deletion",
			eventID:        "1",
			authToken:      token,
			expectedStatus: http.StatusOK,
			setupEvent:     true,
			eventOwnerID:   userId,
		},
		{
			name:           "unauthorized - not owner",
			eventID:        "1",
			authToken:      otherToken,
			expectedStatus: http.StatusUnauthorized,
			setupEvent:     true,
			eventOwnerID:   userId,
		},
		{
			name:           "missing authorization",
			eventID:        "1",
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
			setupEvent:     true,
			eventOwnerID:   userId,
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
					"Event to Delete", "Description", "Location", time.Now(), tt.eventOwnerID,
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

			req, _ := http.NewRequest(http.MethodDelete, "/events/"+actualEventID, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", tt.authToken)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "event deleted successfully", response["message"])
			}
		})
	}
}

