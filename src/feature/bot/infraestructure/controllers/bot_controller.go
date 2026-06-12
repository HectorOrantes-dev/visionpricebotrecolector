package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/application"
)

type BotController struct {
	useCase *application.FetchAndSaveProductsUseCase
}

func NewBotController(useCase *application.FetchAndSaveProductsUseCase) *BotController {
	return &BotController{
		useCase: useCase,
	}
}

func (c *BotController) HandleRoot(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := req.URL.Query().Get("code")
	if code != "" {
		// Attempt to exchange the code for access token
		clientID := os.Getenv("ML_CLIENT_ID")
		clientSecret := os.Getenv("ML_CLIENT_SECRET")
		
		tokenURL := "https://api.mercadolibre.com/oauth/token"
		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)
		data.Set("code", code)
		
		// Determine redirect_uri based on request host
		scheme := "http"
		if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		redirectURI := fmt.Sprintf("%s://%s", scheme, req.Host)
		// Fallback to Render URL if host is local but we are debugging Render
		if strings.Contains(req.Host, "localhost") || strings.Contains(req.Host, "127.0.0.1") || strings.Contains(req.Host, "8080") {
			// Try to read redirect uri or default to Render
			redirectURI = "https://visionpricebotrecolector.onrender.com"
		}
		data.Set("redirect_uri", redirectURI)

		tokenReq, err := http.NewRequestWithContext(req.Context(), "POST", tokenURL, strings.NewReader(data.Encode()))
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html><head><title>OAuth Error</title></head>
			<body style="font-family: Arial, sans-serif; padding: 40px; text-align: center;">
				<h1 style="color: #e74c3c;">Error</h1>
				<p>No se pudo crear la petición de autenticación: %v</p>
			</body></html>`, err)
			return
		}
		tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		tokenReq.Header.Set("Accept", "application/json")
		tokenReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")

		client := &http.Client{}
		resp, err := client.Do(tokenReq)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html><head><title>OAuth Error</title></head>
			<body style="font-family: Arial, sans-serif; padding: 40px; text-align: center;">
				<h1 style="color: #e74c3c;">Error de red</h1>
				<p>Error de conexión al servidor de Mercado Libre: %v</p>
			</body></html>`, err)
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Mercado Libre OAuth Error</title>
				<style>
					body { font-family: Arial, sans-serif; background-color: #f4f4f9; padding: 40px; color: #333; }
					.container { max-width: 600px; margin: 0 auto; background: #fff; padding: 30px; border-radius: 8px; box-shadow: 0 4px 15px rgba(0,0,0,0.1); }
					h1 { color: #e74c3c; }
					.error-box { background: #fdf2f2; border-left: 5px solid #e74c3c; padding: 15px; font-family: monospace; word-break: break-all; margin: 15px 0; }
					.instruction { font-size: 14px; margin-top: 20px; line-height: 1.6; }
				</style>
			</head>
			<body>
				<div class="container">
					<h1>Error en el intercambio de código</h1>
					<p>Mercado Libre devolvió un código de estado %d:</p>
					<div class="error-box">%s</div>
					<p class="instruction">
						Asegúrate de que estás usando un código reciente (los códigos de Mercado Libre expiran en 10 minutos y solo sirven para un único uso).<br>
						También verifica que tus variables <code>ML_CLIENT_ID</code> y <code>ML_CLIENT_SECRET</code> en tu archivo <code>.env</code> o configuración de Render sean correctas.<br><br>
						<strong>Redirect URI intentado:</strong> <code>%s</code>
					</p>
				</div>
			</body>
			</html>`, resp.StatusCode, string(respBody), redirectURI)
			return
		}

		var tokenData map[string]interface{}
		_ = json.Unmarshal(respBody, &tokenData)

		accessToken, _ := tokenData["access_token"].(string)
		refreshToken, _ := tokenData["refresh_token"].(string)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Mercado Libre OAuth Success</title>
			<style>
				body { font-family: Arial, sans-serif; background-color: #f4f4f9; padding: 40px; color: #333; }
				.container { max-width: 650px; margin: 0 auto; background: #fff; padding: 30px; border-radius: 8px; box-shadow: 0 4px 15px rgba(0,0,0,0.1); }
				h1 { color: #2ecc71; }
				.token-box { background: #eef9f2; border-left: 5px solid #2ecc71; padding: 15px; font-family: monospace; word-break: break-all; margin: 15px 0; font-size: 13px; }
				.instruction { font-size: 14px; margin-top: 20px; line-height: 1.6; }
				code { background: #eee; padding: 2px 6px; border-radius: 4px; font-family: monospace; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>¡Conectado exitosamente con Mercado Libre!</h1>
				<p>Hemos intercambiado el código de autorización por tus tokens de acceso.</p>
				
				<h3>Nuevo ML_ACCESS_TOKEN:</h3>
				<div class="token-box">%s</div>
				
				<h3>Nuevo ML_REFRESH_TOKEN:</h3>
				<div class="token-box">%s</div>
				
				<p class="instruction">
					<strong>¿Qué hacer ahora?</strong><br>
					1. Copia el <code>ML_ACCESS_TOKEN</code> completo de arriba.<br>
					2. Ve a la configuración de tu aplicación en Render (o a tu archivo <code>.env</code> local).<br>
					3. Actualiza el valor de la variable <code>ML_ACCESS_TOKEN</code> con este nuevo token.<br>
					4. Si lo deseas, guarda también el <code>ML_REFRESH_TOKEN</code>.<br>
					5. Guarda los cambios y despliega de nuevo.
				</p>
			</div>
		</body>
		</html>`, accessToken, refreshToken)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "running",
		"message":     "Vision Price Bot Recolector is active.",
		"endpoints":   []string{"/health", "/sync"},
		"description": "Trigger manual synchronization by visiting /sync?category=your_category",
	})
}

func (c *BotController) HandleHealth(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func (c *BotController) HandleSync(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost && req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	category := req.URL.Query().Get("category")
	if category == "" {
		category = "materiales de construccion"
	}

	ctx := req.Context()
	err := c.useCase.Execute(ctx, category)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "sync triggered successfully for category: " + category,
	})
}
