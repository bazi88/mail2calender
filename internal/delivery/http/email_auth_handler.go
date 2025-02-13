package http

import (
	"html/template"
	"net/http"
)

// EmailAuthHandler xử lý các yêu cầu xác thực email
type EmailAuthHandler struct {
	emailAuthService EmailAuthService
}

// NewEmailAuthHandler tạo một EmailAuthHandler mới
func NewEmailAuthHandler(emailAuthService EmailAuthService) *EmailAuthHandler {
	return &EmailAuthHandler{
		emailAuthService: emailAuthService,
	}
}

// HandleCallback xử lý callback từ OAuth provider
func (h *EmailAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	err := h.emailAuthService.ExchangeCodeForToken(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Email Authorization Success</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				margin: 0;
				padding: 20px;
				background-color: #f5f5f5;
			}
			.container {
				max-width: 600px;
				margin: 40px auto;
				padding: 20px;
				background-color: white;
				border-radius: 8px;
				box-shadow: 0 2px 4px rgba(0,0,0,0.1);
				text-align: center;
			}
			h1 {
				color: #2c3e50;
				margin-bottom: 20px;
			}
			.success-message {
				color: #27ae60;
				font-size: 24px;
				margin: 20px 0;
			}
			p {
				color: #7f8c8d;
				line-height: 1.6;
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
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
