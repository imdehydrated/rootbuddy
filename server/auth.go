package server

import (
	"net/http"
	"strings"
)

func playerTokenFromRequest(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get("X-Player-Token"))
}
