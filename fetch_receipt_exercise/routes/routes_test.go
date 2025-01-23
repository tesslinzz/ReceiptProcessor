package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	router := gin.Default()
	RegisterRoutes(router)
	return router
}

func TestProcessReceipt_ValidInput(t *testing.T) {
	router := setupRouter()

	payload := `{
		"retailer": "Target",
		"purchaseDate": "2023-09-15",
		"purchaseTime": "14:30",
		"total": "35.50",
		"items": [
			{"shortDescription": "Apples", "price": "3.50"},
			{"shortDescription": "Bananas", "price": "2.00"}
		]
	}`
	req, _ := http.NewRequest("POST", "/receipts/process", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "id") // Ensure an ID is returned
}

func TestProcessReceipt_InvalidInput(t *testing.T) {
	router := setupRouter()

	payload := `{
		"retailer": "Target",
		"purchaseDate": "2023-09-15",
		"total": "35.50",
		"items": []
	}`
	req, _ := http.NewRequest("POST", "/receipts/process", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "BadRequest")
}

func TestGetPoints_ValidID(t *testing.T) {
	router := setupRouter()

	// Step 1: Add a receipt
	payload := `{
		"retailer": "Target",
		"purchaseDate": "2023-09-15",
		"purchaseTime": "14:30",
		"total": "35.50",
		"items": [
			{"shortDescription": "Apples", "price": "3.50"},
			{"shortDescription": "Bananas", "price": "2.00"}
		]
	}`
	req, _ := http.NewRequest("POST", "/receipts/process", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract ID from response
	var response map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	id := response["id"]

	// Step 2: Fetch points using the ID
	req, _ = http.NewRequest("GET", "/receipts/"+id+"/points", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "points")
}

func TestGetPoints_InvalidID(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/receipts/invalid-id/points", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NotFound")
}
