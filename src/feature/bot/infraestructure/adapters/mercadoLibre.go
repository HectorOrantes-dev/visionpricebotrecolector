package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/domain/entities"
)

type MLProductFetcherAdapter struct {
	client       *http.Client
	siteID       string
	clientID     string
	clientSecret string
	accessToken  string
}

func NewMLProductFetcherAdapter(siteID string) *MLProductFetcherAdapter {
	if siteID == "" {
		siteID = "MLM" // Default to Mexico
	}
	return &MLProductFetcherAdapter{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		siteID:       siteID,
		clientID:     os.Getenv("ML_CLIENT_ID"),
		clientSecret: os.Getenv("ML_CLIENT_SECRET"),
		accessToken:  os.Getenv("ML_ACCESS_TOKEN"),
	}
}

type mlTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type mlSearchResult struct {
	Results []struct {
		ID         string  `json:"id"`
		Title      string  `json:"title"`
		Price      float64 `json:"price"`
		CurrencyID string  `json:"currency_id"`
	} `json:"results"`
}

type mlDescriptionResult struct {
	PlainText string `json:"plain_text"`
}

func (a *MLProductFetcherAdapter) getOrFetchToken(ctx context.Context) (string, error) {
	if a.accessToken != "" {
		return a.accessToken, nil
	}

	if a.clientID == "" || a.clientSecret == "" {
		return "", fmt.Errorf("Mercado Libre authentication required. Please define ML_ACCESS_TOKEN or both ML_CLIENT_ID and ML_CLIENT_SECRET in your .env file")
	}

	tokenURL := "https://api.mercadolibre.com/oauth/token"
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", a.clientID)
	data.Set("client_secret", a.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request execution failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return "", fmt.Errorf("auth endpoint returned status %d: %s", resp.StatusCode, buf.String())
	}

	var tokenRes mlTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	a.accessToken = tokenRes.AccessToken
	return a.accessToken, nil
}

func (a *MLProductFetcherAdapter) FetchByCategory(ctx context.Context, category string) ([]entities.Product, error) {
	token, err := a.getOrFetchToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.mercadolibre.com/sites/%s/search", a.siteID)

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	if len(category) > 3 && category[:3] == a.siteID {
		q.Set("category", category)
	} else {
		q.Set("q", category)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf("search API returned status %d: %s", resp.StatusCode, buf.String())
	}

	var searchRes mlSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchRes); err != nil {
		return nil, err
	}

	products := make([]entities.Product, len(searchRes.Results))
	for i, r := range searchRes.Results {
		products[i] = entities.Product{
			MLID:      r.ID,
			Nombre:    r.Title,
			Precio:    r.Price,
			Moneda:    r.CurrencyID,
			Categoria: category,
			CreatedAt: time.Now(),
		}
	}

	return products, nil
}

func (a *MLProductFetcherAdapter) FetchItemDescription(ctx context.Context, mlID string) (string, error) {
	token, err := a.getOrFetchToken(ctx)
	if err != nil {
		return "", err
	}

	apiURL := fmt.Sprintf("https://api.mercadolibre.com/items/%s/description", mlID)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", nil
		}
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return "", fmt.Errorf("description API returned status %d: %s", resp.StatusCode, buf.String())
	}

	var descRes mlDescriptionResult
	if err := json.NewDecoder(resp.Body).Decode(&descRes); err != nil {
		return "", err
	}

	return descRes.PlainText, nil
}
