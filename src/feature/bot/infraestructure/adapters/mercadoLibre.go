package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/domain/entities"
)

type MLProductFetcherAdapter struct {
	client *http.Client
	siteID string
}

func NewMLProductFetcherAdapter(siteID string) *MLProductFetcherAdapter {
	if siteID == "" {
		siteID = "MLM" // Default to Mexico
	}
	return &MLProductFetcherAdapter{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		siteID: siteID,
	}
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

func (a *MLProductFetcherAdapter) FetchByCategory(ctx context.Context, category string) ([]entities.Product, error) {
	apiURL := fmt.Sprintf("https://api.mercadolibre.com/sites/%s/search", a.siteID)

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	// If the category parameter starts with our siteID (like MLM or MLA), set it as category ID.
	// Otherwise, treat it as a query parameter search query 'q'.
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

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from ML API: %d", resp.StatusCode)
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
	apiURL := fmt.Sprintf("https://api.mercadolibre.com/items/%s/description", mlID)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", nil
		}
		return "", fmt.Errorf("unexpected status code from ML Description API: %d", resp.StatusCode)
	}

	var descRes mlDescriptionResult
	if err := json.NewDecoder(resp.Body).Decode(&descRes); err != nil {
		return "", err
	}

	return descRes.PlainText, nil
}
