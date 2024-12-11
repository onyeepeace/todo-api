package handlers

import "net/http"

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "OK", http.StatusOK)
}
