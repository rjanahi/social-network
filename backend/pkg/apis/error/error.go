package error

import "net/http"

func ErrorHandler(w http.ResponseWriter, r *http.Request, errNum int) {
	// Set the content type
	w.Header().Set("Content-Type", "text/html")

	// Check if the status code is not already set
	if w.Header().Get("X-Status-Code") == "" {
		switch errNum {
		case 400:
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
		case 404:
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
		case 500:
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		default:
			w.WriteHeader(http.StatusInternalServerError) // Default to 500
		}
		w.Header().Set("X-Status-Code", "true") // Mark that the status code has been set
	}
}
