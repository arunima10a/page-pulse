package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/arunima10a/page-pulse/internal/auditor"
)

// AuditHandler processes the incoming audit request
func AuditHandler(w http.ResponseWriter, r *http.Request) {
	// Extract and Validate Input
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		sendError(w, "URL parameter is required", http.StatusBadRequest)
		return
	}

	// Basic check: must be absolute URL (http/https)
	parsedURL, err := url.ParseRequestURI(targetURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		sendError(w, "Invalid URL format. Please include http:// or https://", http.StatusBadRequest)
		return
	}

	// Perform the Audit
	report, err := auditor.AnalyzeURL(targetURL)
	if err != nil {
		// If the error contains "unsupported content type", it's a 415 Unsupported Media Type
		if strings.Contains(err.Error(), "unsupported content type") {
			sendError(w, err.Error(), http.StatusUnsupportedMediaType) 
			return
		}
		// If it's a timeout or DNS error
		sendError(w, err.Error(), http.StatusBadGateway) // 502 indicates the external site failed
		return
	}

	// Return the JSON Report
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

// sendError is a helper to ensure all errors follow the same JSON structure
func sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
