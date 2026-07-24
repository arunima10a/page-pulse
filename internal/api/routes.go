package api

import "net/http"

// NewRouter sets up the API routes and static file serving
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// API Endpoints
	mux.HandleFunc("/api/audit", AuditHandler)

	//Static Assets
	fs := http.FileServer(http.Dir("./public"))
	mux.Handle("/", fs)

	return mux
}