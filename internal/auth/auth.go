package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/DishIs/fce-cli/internal/config"
	"github.com/DishIs/fce-cli/internal/display"
)

const (
	loginBaseURL   = "https://freecustom.email/api/cli-auth"
	callbackPort   = 9876
	callbackPath   = "/callback"
	timeoutSeconds = 120
)

// Login opens the browser to the auth page, starts a local server to receive
// the API key, then stores it securely.
func Login() error {
	// Check if already logged in
	if config.IsLoggedIn() {
		cfg, _ := config.LoadConfig()
		display.Warn("Already logged in" + func() string {
			if cfg != nil && cfg.Plan != "" {
				return " (" + display.PlanBadge(cfg.Plan) + ")"
			}
			return ""
		}())
		display.Info("Run `fce logout` first to switch accounts.")
		return nil
	}

	// Start local callback server
	keyCh  := make(chan string, 1)
	errCh  := make(chan error, 1)
	server := startCallbackServer(keyCh, errCh)

	// Build the auth URL — includes the callback port so the site knows where to redirect
	authURL := fmt.Sprintf("%s?callback=http://localhost:%d%s", loginBaseURL, callbackPort, callbackPath)

	display.Step(1, 3, "Opening browser…")
	display.Info(fmt.Sprintf("If the browser doesn't open, visit:\n    %s", authURL))
	fmt.Println()

	if err := openBrowser(authURL); err != nil {
		display.Warn("Could not open browser automatically.")
		display.Info("Open this URL manually:")
		fmt.Printf("\n    %s\n\n", authURL)
	}

	display.Step(2, 3, "Waiting for authentication…")
	display.Info("Complete login in the browser. This window will update automatically.")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSeconds*time.Second)
	defer cancel()

	var apiKey string
	select {
	case key := <-keyCh:
		apiKey = key
	case err := <-errCh:
		server.Shutdown(ctx)
		return fmt.Errorf("auth error: %w", err)
	case <-ctx.Done():
		server.Shutdown(ctx)
		return fmt.Errorf("login timed out after %d seconds", timeoutSeconds)
	}

	server.Shutdown(ctx)

	if apiKey == "" {
		return fmt.Errorf("received empty API key")
	}

	display.Step(3, 3, "Saving credentials…")
	if err := config.SaveAPIKey(apiKey); err != nil {
		return fmt.Errorf("failed to save API key: %w", err)
	}

	// Mark first login done
	cfg, _ := config.LoadConfig()
	isFirst := cfg.FirstLogin
	cfg.FirstLogin = false
	_ = config.SaveConfig(cfg)

	fmt.Println()
	if isFirst {
		display.PrintLogo()
	}
	display.Success("Logged in successfully!")
	display.Info("Run `fce status` to see your account details.")
	fmt.Println()

	return nil
}

// Logout removes the stored API key
func Logout() error {
	if !config.IsLoggedIn() {
		display.Warn("Not currently logged in.")
		return nil
	}
	if err := config.DeleteAPIKey(); err != nil {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}
	// Reset config
	cfg := &config.Config{FirstLogin: true}
	_ = config.SaveConfig(cfg)
	display.Success("Logged out.")
	return nil
}

// ── Local callback server ─────────────────────────────────────────────────────

func startCallbackServer(keyCh chan<- string, errCh chan<- error) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("api_key")
		if key == "" {
			errCh <- fmt.Errorf("no api_key in callback")
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, callbackHTMLError)
			return
		}
		keyCh <- key
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, callbackHTMLSuccess)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", callbackPort),
		Handler: mux,
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		errCh <- fmt.Errorf("could not start callback server: %w", err)
		return server
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			// ignore — normal on shutdown
		}
	}()

	return server
}

// ── Browser opener ────────────────────────────────────────────────────────────

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default: // linux
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// ── Callback HTML pages ───────────────────────────────────────────────────────

const callbackHTMLSuccess = `<!DOCTYPE html><html><head><meta charset="utf-8">
<title>FreeCustom.Email CLI</title>
<style>
  *{margin:0;padding:0;box-sizing:border-box}
  body{background:#000;color:#fff;font-family:monospace;display:flex;align-items:center;justify-content:center;min-height:100vh;flex-direction:column;gap:16px}
  .icon{font-size:48px}
  h1{font-size:18px;font-weight:600;letter-spacing:-.01em}
  p{font-size:13px;color:#666;line-height:1.6;text-align:center;max-width:320px}
  .badge{border:1px solid #333;border-radius:4px;padding:4px 10px;font-size:11px;color:#999;margin-top:8px}
</style>
</head><body>
<div class="icon">✓</div>
<h1>Authentication successful</h1>
<p>You're now logged in to the FreeCustom.Email CLI. You can close this tab.</p>
<div class="badge">fce</div>
</body></html>`

const callbackHTMLError = `<!DOCTYPE html><html><head><meta charset="utf-8">
<title>FreeCustom.Email CLI</title>
<style>
  *{margin:0;padding:0;box-sizing:border-box}
  body{background:#000;color:#fff;font-family:monospace;display:flex;align-items:center;justify-content:center;min-height:100vh;flex-direction:column;gap:16px}
  .icon{font-size:48px}
  h1{font-size:18px;font-weight:600}
  p{font-size:13px;color:#666;text-align:center;max-width:320px}
</style>
</head><body>
<div class="icon">✗</div>
<h1>Authentication failed</h1>
<p>Something went wrong. Please try running <code>fce login</code> again.</p>
</body></html>`
