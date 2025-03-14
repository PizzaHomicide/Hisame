package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	"github.com/PizzaHomicide/hisame/internal/log"
)

const (
	callbackPort = "19331"
	callbackPath = "/callback"
	tokenPath    = "/token"
	clientID     = "18776"
)

// Result represents the outcome of an authentication attempt
type Result struct {
	Token string
	Error error
}

// Auth manages the OAuth authentication flow with AniList
type Auth struct {
	LoginURL     *url.URL
	tokenChannel chan string
	httpServer   *http.Server
}

// NewAuth creates a new Auth instance
func NewAuth() *Auth {
	return &Auth{
		LoginURL:     generateAuthURL(),
		tokenChannel: make(chan string, 1),
		httpServer:   nil,
	}
}

// StartCallbackServer starts the HTTP server listening for the callback from AniList
func (auth *Auth) StartCallbackServer() error {
	log.Info("Starting auth callback server")

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, handleCallback)
	mux.HandleFunc(tokenPath, auth.handleToken())

	// Create auth listener early so we can report an error if we can't secure the port
	listener, err := net.Listen("tcp", ":"+callbackPort)
	if err != nil {
		log.Error("Could not listen on port", "port", callbackPort, "error", err)
		return err
	}

	auth.httpServer = &http.Server{
		Handler: mux,
	}

	go func() {
		if err := auth.httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Server error", "error", err)
		}
	}()

	return nil
}

// DoAuth performs the entire authentication flow and returns the result
func (auth *Auth) DoAuth() Result {
	// Start the callback server
	if err := auth.StartCallbackServer(); err != nil {
		return Result{Error: err}
	}

	// Open the browser with the login URL
	if err := OpenBrowser(auth.LoginURL.String()); err != nil {
		log.Warn("Failed to open browser automatically", "error", err)
		// Note: We continue the flow even if browser opening fails,
		// as the user can manually navigate to the URL
	}

	// Create a context with timeout for token waiting
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Wait for the token
	token, err := auth.WaitForToken(ctx)
	if err != nil {
		return Result{Error: err}
	}

	return Result{Token: token}
}

// WaitForToken waits for a token to be received via the callback
func (auth *Auth) WaitForToken(ctx context.Context) (string, error) {
	log.Debug("Waiting for token to arrive on /token endpoint")
	// Ensure the callback server is stopped after we finish waiting
	defer auth.StopCallbackServer()

	// Wait for the token to be received
	select {
	case <-ctx.Done():
		log.Debug("WaitForToken exiting because context is done")
		return "", ctx.Err()
	case token, ok := <-auth.tokenChannel:
		if !ok || token == "" {
			log.Warn("Failed to receive token")
			return "", errors.New("failed to receive token")
		}
		log.Info("Received token")
		return token, nil
	}
}

// StopCallbackServer stops the HTTP server
func (auth *Auth) StopCallbackServer() {
	if auth.httpServer == nil {
		log.Warn("Call to StopCallbackServer when server was not started")
		return
	}
	log.Debug("Stopping callback server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := auth.httpServer.Shutdown(ctx); err != nil {
		log.Error("Server shutdown failed", "error", err)
	}
	log.Debug("Callback server shutdown successfully")
}

// generateAuthURL creates the AniList OAuth URL
func generateAuthURL() *url.URL {
	loginURL, err := url.Parse(fmt.Sprintf("https://anilist.co/api/v2/oauth/authorize?client_id=%s&response_type=token", clientID))
	if err != nil {
		log.Error("Failed to generate auth url", "error", err)
		panic("Failed to generate auth url. Exiting application.")
	}
	return loginURL
}

// handleToken creates a handler for the token endpoint
func (auth *Auth) handleToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Received post to token endpoint")
		var data struct {
			Token string `json:"token"`
		}

		// Parse the token from the POST request body
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		log.Debug("Token decoded", "length", len(data.Token))

		// Send the token to the channel
		auth.tokenChannel <- data.Token

		// Send auth success response back
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "token stored"})
	}
}

// handleCallback handles the callback from AniList
func handleCallback(w http.ResponseWriter, r *http.Request) {
	htmlContent := `
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Hisame Auth</title>
        <script>
            window.onload = function() {
                const fragment = window.location.hash.substring(1);
                const params = new URLSearchParams(fragment);
                const token = params.get("access_token");

                if (token) {
                    fetch("/token", {
                        method: "POST",
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ token: token })
                    }).then(response => response.json())
                    .then(data => {
                        document.body.innerHTML = "<h1>Authentication successful!</h1><p>You can close this window and return to Hisame.</p>";
                    }).catch((error) => {
                        document.body.innerHTML = "<h1>Error retrieving token: " + error + "</h1>";
                    });
                } else {
                    document.body.innerHTML = "<h1>No token found in the URL fragment</h1>";
                }
            };
        </script>
    </head>
    <body>
        <h1>Processing OAuth Token...</h1>
    </body>
    </html>
    `
	w.Header().Set("Content-Type", "text/html")
	_, err := fmt.Fprint(w, htmlContent)
	if err != nil {
		log.Error("Error handling callback", "error", err)
	}
}

// OpenBrowser opens the specified URL in the default browser
func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default: // Linux and others
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
