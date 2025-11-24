package api

import (
	"net/http"
	"strings"
)

// WithCORS wraps handlers and sets CORS headers.
// It echoes back the Origin header for a small allowlist so credentials (cookies)
// can be sent from the frontend. Avoid using a wildcard when Credential=true.
func WithCORS(handler http.HandlerFunc) http.HandlerFunc {
	allowed := []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		// When served inside Docker, the frontend container hostname may be different;
		// the frontend uses BACKEND_URL=http://backend:8080, but browser requests will
		// still originate from the host (localhost:3000). Keep the list small and safe.
	}

	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			for _, a := range allowed {
				if strings.EqualFold(a, origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		// Fallback: if Access-Control-Allow-Origin wasn't set and the request is from
		// localhost on port 3000 using an IP form (e.g., 0.0.0.0), allow it by echoing
		// any origin that contains ":3000" — this is convenient for local testing.
		if w.Header().Get("Access-Control-Allow-Origin") == "" && origin != "" && strings.Contains(origin, ":3000") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Protocol, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}
