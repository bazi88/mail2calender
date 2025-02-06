package http

import (
	"html/template"
	"net/http"

	"mail2calendar/internal/domain/email_auth"

	"github.com/go-chi/chi/v5"
)

type EmailAuthHandler struct {
	authService *email_auth.EmailAuthService
}

func NewEmailAuthHandler(authService *email_auth.EmailAuthService) *EmailAuthHandler {
	return &EmailAuthHandler{authService: authService}
}

func (h *EmailAuthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/auth", h.showAuthPage)
	r.Get("/auth/login", h.handleLogin)
	r.Get("/auth/callback", h.handleCallback)
}

func (h *EmailAuthHandler) showAuthPage(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Mail2Calendar - Email Authorization</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				margin: 0;
				padding: 20px;
				display: flex;
				justify-content: center;
				align-items: center;
				min-height: 100vh;
				background-color: #f5f5f5;
			}
			.container {
				background-color: white;
				padding: 40px;
				border-radius: 8px;
				box-shadow: 0 2px 4px rgba(0,0,0,0.1);
				text-align: center;
			}
			h1 {
				color: #333;
				margin-bottom: 20px;
			}
			p {
				color: #666;
				margin-bottom: 30px;
				line-height: 1.5;
			}
			.button {
				display: inline-block;
				padding: 12px 24px;
				background-color: #4285f4;
				color: white;
				text-decoration: none;
				border-radius: 4px;
				transition: background-color 0.3s;
			}
			.button:hover {
				background-color: #3367d6;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Mail2Calendar Setup</h1>
			<p>To get started, please authorize access to your email and calendar.<br>This will allow the system to automatically process your emails and create calendar events.</p>
			<a href="/auth/login" class="button">Authorize Access</a>
		</div>
	</body>
	</html>
	`

	tmpl, err := template.New("auth").Parse(html)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, nil)
}

func (h *EmailAuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	authURL := h.authService.GetAuthURL()
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *EmailAuthHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// In a real application, you would get the userID from the session
	userID := "default-user"

	if err := h.authService.HandleCallback(r.Context(), code, userID); err != nil {
		http.Error(w, "Failed to complete authentication", http.StatusInternalServerError)
		return
	}

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Mail2Calendar - Setup Complete</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				margin: 0;
				padding: 20px;
				display: flex;
				justify-content: center;
				align-items: center;
				min-height: 100vh;
				background-color: #f5f5f5;
			}
			.container {
				background-color: white;
				padding: 40px;
				border-radius: 8px;
				box-shadow: 0 2px 4px rgba(0,0,0,0.1);
				text-align: center;
			}
			h1 {
				color: #333;
				margin-bottom: 20px;
			}
			.success-message {
				color: #4caf50;
				font-size: 18px;
				margin-bottom: 20px;
			}
			p {
				color: #666;
				margin-bottom: 10px;
				line-height: 1.5;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Setup Complete!</h1>
			<div class="success-message">✓ Authorization successful</div>
			<p>Your email and calendar are now connected.</p>
			<p>You can close this window and the system will continue running in the background.</p>
		</div>
	</body>
	</html>
	`

	tmpl, err := template.New("success").Parse(html)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, nil)
}
