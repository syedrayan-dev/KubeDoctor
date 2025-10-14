/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp"`
	Code      int    `json:"code"`
}

// WriteJSONResponse writes a JSON response with the given status code
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// WriteErrorResponse writes an error response with safe error messages
func WriteErrorResponse(w http.ResponseWriter, statusCode int, err error, userMessage string) {
	// Log the actual error for debugging (server-side only)
	if err != nil {
		log.Printf("API Error: %v", err)
	}

	// Create user-friendly error response
	errorResp := ErrorResponse{
		Error:     userMessage,
		Timestamp: time.Now().Format(time.RFC3339),
		Code:      statusCode,
	}

	WriteJSONResponse(w, statusCode, errorResp)
}

// WriteSuccessResponse writes a success response with data
func WriteSuccessResponse(w http.ResponseWriter, data interface{}) {
	WriteJSONResponse(w, http.StatusOK, data)
}

// WriteCreatedResponse writes a 201 Created response
func WriteCreatedResponse(w http.ResponseWriter, data interface{}) {
	WriteJSONResponse(w, http.StatusCreated, data)
}

// WriteNoContentResponse writes a 204 No Content response
func WriteNoContentResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WriteBadRequestError writes a 400 Bad Request error response
func WriteBadRequestError(w http.ResponseWriter, err error, userMessage string) {
	WriteErrorResponse(w, http.StatusBadRequest, err, userMessage)
}

// WriteNotFoundError writes a 404 Not Found error response
func WriteNotFoundError(w http.ResponseWriter, err error, userMessage string) {
	WriteErrorResponse(w, http.StatusNotFound, err, userMessage)
}

// WriteInternalServerError writes a 500 Internal Server Error response
func WriteInternalServerError(w http.ResponseWriter, err error, userMessage string) {
	WriteErrorResponse(w, http.StatusInternalServerError, err, userMessage)
}

// WriteUnauthorizedError writes a 401 Unauthorized error response
func WriteUnauthorizedError(w http.ResponseWriter, err error, userMessage string) {
	WriteErrorResponse(w, http.StatusUnauthorized, err, userMessage)
}

// WriteForbiddenError writes a 403 Forbidden error response
func WriteForbiddenError(w http.ResponseWriter, err error, userMessage string) {
	WriteErrorResponse(w, http.StatusForbidden, err, userMessage)
}
