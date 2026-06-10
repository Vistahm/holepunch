package server

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"holepunch/internal/auth"
)

// New creates a configured HTTP server.
func New(dir string, authenticator auth.Authenticator, mode string, quiet bool) *http.Server {
	fs := http.FileServer(http.Dir(dir))

	// Wrap with auth middleware (pass mode so it can set cookies for token auth)
	handler := auth.Middleware(authenticator, mode, fs)

	mux := http.NewServeMux()
	mux.Handle("/", handler)

	return &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// PrintInfo displays connection details.
func PrintInfo(dir string, port int, externalIP string, creds *auth.Credentials) {
	absDir, _ := filepath.Abs(dir)
	fmt.Printf("\n📁 Serving: %s\n", absDir)
	fmt.Printf("🔗 Local:   http://localhost:%d\n", port)

	if externalIP != "" {
		fmt.Printf("🌐 Remote:  %s\n", creds.FormatURL(externalIP, port))
	}

	fmt.Println("\nPress Ctrl+C to stop...")
}
