package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

// Authenticator validates incoming requests
type Authenticator func(r *http.Request) bool

// Credentials holds the info needed to access the server.
type Credentials struct {
	Mode      string
	User      string
	Pass      string
	Token     string
	AccessURL string
}

// Print displays credentials to the terminal
func (c *Credentials) Print() {
	switch c.Mode {
	case "basic":
		fmt.Printf("Auth: Basic Auth (username: %s)\n", c.User)
		fmt.Printf("Password: %s\n", c.Pass)
	case "token":
		fmt.Printf("Auth: Token-based\n")
		fmt.Printf("Token: %s\n", c.Token)
		fmt.Println("⚠️  Save this token — you'll need it to access the files.")
	case "none":
		fmt.Println("⚠️  No authentication enabled — anyone can access the server.")
	}
}

// FormatURL returns the access URL with the given external IP
func (c *Credentials) FormatURL(ip string, port int) string {
	url := strings.Replace(c.AccessURL, "EXTERNAL_IP", ip, 1)
	url = strings.Replace(url, "PORT", fmt.Sprintf("%d", port), 1)
	return url
}

// Setup creates an authenticator based on provided flags
func Setup(user, pass, token string) (Authenticator, *Credentials) {
	creds := &Credentials{}

	// Option 1: Basic Auth
	if user != "" && pass != "" {
		creds.Mode = "basic"
		creds.User = user
		creds.Pass = pass
		creds.AccessURL = "http://EXTERNAL_IP:PORT/"

		auth := func(r *http.Request) bool {
			u, p, ok := r.BasicAuth()
			return ok &&
				subtle.ConstantTimeCompare([]byte(u), []byte(user)) == 1 &&
				subtle.ConstantTimeCompare([]byte(p), []byte(pass)) == 1
		}
		return auth, creds
	}

	// Option 2: Pre-shared token
	if token != "" {
		hash := hashToken(token)
		creds.Mode = "token"
		creds.Token = token
		creds.AccessURL = "http://EXTERNAL_IP:PORT/?token=" + token

		auth := createTokenAuth(hash)
		return auth, creds
	}

	// Option 3: Generate random token
	rawToken := generateToken(16)
	hash := hashToken(rawToken)
	creds.Mode = "token"
	creds.Token = rawToken
	creds.AccessURL = "http://EXTERNAL_IP:PORT/?token=" + rawToken

	auth := createTokenAuth(hash)
	return auth, creds
}

// createTokenAuth returns an authenticator that checks both URL param and cookie
func createTokenAuth(expectedHash string) Authenticator {
	return func(r *http.Request) bool {
		// Check URL query parameter
		if t := r.URL.Query().Get("token"); t != "" {
			if subtle.ConstantTimeCompare([]byte(hashToken(t)), []byte(expectedHash)) == 1 {
				return true
			}
		}

		// Check cookie
		if cookie, err := r.Cookie("holepunch_token"); err == nil {
			if subtle.ConstantTimeCompare([]byte(hashToken(cookie.Value)), []byte(expectedHash)) == 1 {
				return true
			}
		}

		return false
	}
}

// Middleware wraps a handler with authentication.
// On first token access via URL, it sets a cookie so subsequent requests work.
func Middleware(auth Authenticator, mode string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth(r) {
			if mode == "basic" {
				w.Header().Set("WWW-Authenticate", `Basic realm="HolePunch"`)
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// If token auth and token was passed in URL, set it as a cookie
		// so navigation within the directory listing works without the URL param
		if mode == "token" {
			if t := r.URL.Query().Get("token"); t != "" {
				http.SetCookie(w, &http.Cookie{
					Name:     "holepunch_token",
					Value:    t,
					Path:     "/",
					HttpOnly: true,
				})
			}
		}

		next.ServeHTTP(w, r)
	})
}

func generateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
