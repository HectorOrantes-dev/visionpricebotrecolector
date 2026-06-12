package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/domain/entities"
)

type SupabaseRepositoryAdapter struct {
	client *http.Client
	apiURL string
	apiKey string
}

func NewSupabaseRepositoryAdapter(apiURL, apiKey string) *SupabaseRepositoryAdapter {
	// Trim trailing slashes and rest/v1 path if present
	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, "/rest/v1")
	apiURL = strings.TrimSuffix(apiURL, "/")

	return &SupabaseRepositoryAdapter{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiURL: apiURL,
		apiKey: apiKey,
	}
}

func (r *SupabaseRepositoryAdapter) setHeaders(req *http.Request) {
	req.Header.Set("apikey", r.apiKey)
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")
}

func (r *SupabaseRepositoryAdapter) Upsert(ctx context.Context, product *entities.Product) error {
	body, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("error marshaling product: %w", err)
	}

	// PostgREST upsert URL with on_conflict param
	reqURL := fmt.Sprintf("%s/rest/v1/products?on_conflict=ml_id", r.apiURL)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	r.setHeaders(req)
	req.Header.Set("Prefer", "resolution=merge-duplicates")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return fmt.Errorf("unexpected status code from Supabase REST API: %d, response: %s", resp.StatusCode, buf.String())
	}

	return nil
}

func (r *SupabaseRepositoryAdapter) SaveSnapshot(ctx context.Context, snapshot *entities.PriceSnapshot) error {
	body, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("error marshaling snapshot: %w", err)
	}

	reqURL := fmt.Sprintf("%s/rest/v1/price_snapshots", r.apiURL)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	r.setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return fmt.Errorf("unexpected status code from Supabase REST API: %d, response: %s", resp.StatusCode, buf.String())
	}

	return nil
}

func (r *SupabaseRepositoryAdapter) ListByCategory(ctx context.Context, category string) ([]entities.Product, error) {
	reqURL := fmt.Sprintf("%s/rest/v1/products?categoria=eq.%s", r.apiURL, url.QueryEscape(category))
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	r.setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf("unexpected status code from Supabase REST API: %d, response: %s", resp.StatusCode, buf.String())
	}

	var products []entities.Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return products, nil
}
