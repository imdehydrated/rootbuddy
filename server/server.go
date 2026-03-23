package server

import "net/http"

func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", HandleHealthCheck)
	mux.HandleFunc("GET /api/health", HandleHealthCheck)
	mux.HandleFunc("POST /actions/valid", HandleValidActions)
	mux.HandleFunc("POST /api/actions/valid", HandleValidActions)
	mux.HandleFunc("POST /actions/apply", HandleApplyAction)
	mux.HandleFunc("POST /api/actions/apply", HandleApplyAction)
	mux.HandleFunc("POST /battles/resolve", HandleResolveBattle)
	mux.HandleFunc("POST /api/battles/resolve", HandleResolveBattle)
	return withCORS(mux)
}
