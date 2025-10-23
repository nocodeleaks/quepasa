package metrics

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// DashboardHandler serves the QuePasa metrics dashboard
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("serving dashboard")

	// Get the absolute path to the dashboard.html file
	dashboardPath := filepath.Join("views", "dashboard.html")

	// Open and read the dashboard file
	file, err := os.Open(dashboardPath)
	if err != nil {
		log.Errorf("error opening dashboard file: %v", err)
		http.Error(w, "Dashboard file not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Copy file contents to response
	_, err = io.Copy(w, file)
	if err != nil {
		log.Errorf("error serving dashboard: %v", err)
		http.Error(w, "Error serving dashboard", http.StatusInternalServerError)
		return
	}
}
