package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	maxRetries     = 3
	baseDelay      = 500 * time.Millisecond
	maxRetryAfter  = 60
	maxResponseBody = 1 << 20 // 1 MB
	maxErrorBodyLen = 200
)

var ErrNotFound = errors.New("resource not found")

type SporkClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
}

func NewSporkClient(baseURL, apiKey, version string) *SporkClient {
	parsedBase, _ := url.Parse(baseURL)
	return &SporkClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				// Strip Authorization header on cross-origin redirects to
				// prevent credential leakage to external domains.
				if parsedBase != nil && req.URL.Host != parsedBase.Host {
					req.Header.Del("Authorization")
				}
				return nil
			},
		},
		UserAgent: fmt.Sprintf("spork-terraform/%s", version),
	}
}

// API model structs aligned with the REST API

// Monitor matches the JSON shape returned by the REST API.
type Monitor struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	Target          string            `json:"target"`
	Method          string            `json:"method,omitempty"`
	ExpectedStatus  int               `json:"expected_status,omitempty"`
	Interval        int               `json:"interval,omitempty"`
	Timeout         int               `json:"timeout,omitempty"`
	Regions         []string          `json:"regions,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	Body            string            `json:"body,omitempty"`
	Keyword         string            `json:"keyword,omitempty"`
	KeywordType     string            `json:"keyword_type,omitempty"`
	SSLWarnDays     int               `json:"ssl_warn_days,omitempty"`
	AlertChannelIDs []string          `json:"alert_channel_ids,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	Paused          bool              `json:"paused"`
	Status          string            `json:"status,omitempty"`
	CreatedAt       string            `json:"created_at,omitempty"`
	UpdatedAt       string            `json:"updated_at,omitempty"`
}

// AlertChannel matches the JSON shape returned by the REST API.
type AlertChannel struct {
	ID        string            `json:"id,omitempty"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Config    map[string]string `json:"config"`
	Verified  bool              `json:"verified,omitempty"`
	Secret    string            `json:"secret,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
	UpdatedAt string            `json:"updated_at,omitempty"`
}

// StatusPage matches the JSON shape returned by the REST API.
type StatusPage struct {
	ID                      string            `json:"id,omitempty"`
	Name                    string            `json:"name"`
	Slug                    string            `json:"slug"`
	Components              []StatusComponent `json:"components,omitempty"`
	ComponentGroups         []ComponentGroup  `json:"component_groups,omitempty"`
	CustomDomain            string            `json:"custom_domain,omitempty"`
	DomainStatus            string            `json:"domain_status,omitempty"`
	Theme                   string            `json:"theme,omitempty"`
	AccentColor             string            `json:"accent_color,omitempty"`
	FontFamily              string            `json:"font_family,omitempty"`
	HeaderStyle             string            `json:"header_style,omitempty"`
	LogoURL                 string            `json:"logo_url,omitempty"`
	WebhookURL              string            `json:"webhook_url,omitempty"`
	EmailSubscribersEnabled bool              `json:"email_subscribers_enabled"`
	IsPublic                bool              `json:"is_public"`
	Password                string            `json:"password,omitempty"`
	CreatedAt               string            `json:"created_at,omitempty"`
	UpdatedAt               string            `json:"updated_at,omitempty"`
}

// ComponentGroup organizes components into named sections on the status page.
type ComponentGroup struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Order       int    `json:"order"`
}

// StatusComponent maps a monitor to a display name on a status page.
type StatusComponent struct {
	ID          string `json:"id,omitempty"`
	MonitorID   string `json:"monitor_id"`
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
	GroupID     string `json:"group_id,omitempty"`
	Order       int    `json:"order"`
}

// dataEnvelope is the standard API response wrapper: {"data": ...}
type dataEnvelope struct {
	Data json.RawMessage `json:"data"`
}

// listEnvelope is the standard API list response wrapper: {"data": [...], "meta": {...}}
type listEnvelope struct {
	Data json.RawMessage `json:"data"`
	Meta struct {
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	} `json:"meta"`
}

// apiErrorEnvelope matches the REST API error format: {"error": {"code": ..., "message": ...}}
type apiErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *SporkClient) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var jsonBytes []byte
	if body != nil {
		var err error
		jsonBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	url := c.BaseURL + path

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		var reqBody io.Reader
		if jsonBytes != nil {
			reqBody = bytes.NewReader(jsonBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", c.UserAgent)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue // retry on network errors
		}

		respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Retry on transient server errors and rate limiting
		if resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout {
			// Respect Retry-After header if present, capped to prevent unbounded sleep.
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
					if seconds > maxRetryAfter {
						seconds = maxRetryAfter
					}
					if seconds > 0 {
						select {
						case <-ctx.Done():
							return ctx.Err()
						case <-time.After(time.Duration(seconds) * time.Second):
						}
					}
				}
			}
			lastErr = fmt.Errorf("API error (HTTP %d): transient error, retrying", resp.StatusCode)
			continue
		}

		return c.handleResponse(resp.StatusCode, respBody, result)
	}

	return fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

func (c *SporkClient) handleResponse(statusCode int, respBody []byte, result interface{}) error {
	switch statusCode {
	case http.StatusOK, http.StatusCreated:
		// success — unwrap the {data: ...} envelope
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized:
		return fmt.Errorf(
			"Spork API key is invalid or missing.\n\n" +
				"Set your API key to authenticate with Spork:\n\n" +
				"  export SPORK_API_KEY=\"your-api-key\"\n\n" +
				"Don't have an account? Sign up free:\n" +
				"  https://sporkops.com/signup?ref=terraform\n\n" +
				"Generate an API key in the dashboard:\n" +
				"  https://sporkops.com/settings/api-keys\n\n" +
				"Docs: https://sporkops.com/docs")
	case http.StatusPaymentRequired:
		return fmt.Errorf(
			"Subscription required.\n\n" +
				"Subscribe to a plan to get started:\n" +
				"  https://sporkops.com/billing?ref=terraform\n\n" +
				"Plans start at $4/mo.")
	case http.StatusForbidden:
		return fmt.Errorf(
			"Subscription inactive.\n\n" +
				"Subscribe or update your billing to continue:\n" +
				"  https://sporkops.com/billing\n\n" +
				"Plans start at $4/mo.")
	default:
		// Parse structured error: {"error": {"code": "...", "message": "..."}}
		var apiErr apiErrorEnvelope
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("API error (HTTP %d): %s", statusCode, apiErr.Error.Message)
		}
		body := string(respBody)
		if len(body) > maxErrorBodyLen {
			body = body[:maxErrorBodyLen] + "…"
		}
		return fmt.Errorf("API error (HTTP %d): %s", statusCode, body)
	}

	// Unwrap the {"data": ...} envelope before unmarshalling into result.
	if result != nil && len(respBody) > 0 {
		var envelope dataEnvelope
		if err := json.Unmarshal(respBody, &envelope); err != nil {
			return fmt.Errorf("failed to unmarshal response envelope: %w", err)
		}
		if err := json.Unmarshal(envelope.Data, result); err != nil {
			return fmt.Errorf("failed to unmarshal response data: %w", err)
		}
	}

	return nil
}

// Monitor CRUD

func (c *SporkClient) CreateMonitor(ctx context.Context, monitor Monitor) (*Monitor, error) {
	var result Monitor
	err := c.doRequest(ctx, http.MethodPost, "/monitors", monitor, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *SporkClient) GetMonitor(ctx context.Context, id string) (*Monitor, error) {
	var result Monitor
	err := c.doRequest(ctx, http.MethodGet, "/monitors/"+url.PathEscape(id), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *SporkClient) UpdateMonitor(ctx context.Context, id string, monitor Monitor) (*Monitor, error) {
	var result Monitor
	err := c.doRequest(ctx, http.MethodPatch, "/monitors/"+url.PathEscape(id), monitor, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *SporkClient) DeleteMonitor(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, "/monitors/"+url.PathEscape(id), nil, nil)
}

// AlertChannel CRUD — uses /alert-channels endpoint

func (c *SporkClient) CreateAlertChannel(ctx context.Context, channel AlertChannel) (*AlertChannel, error) {
	var result AlertChannel
	err := c.doRequest(ctx, http.MethodPost, "/alert-channels", channel, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *SporkClient) GetAlertChannel(ctx context.Context, id string) (*AlertChannel, error) {
	var result AlertChannel
	err := c.doRequest(ctx, http.MethodGet, "/alert-channels/"+url.PathEscape(id), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *SporkClient) UpdateAlertChannel(ctx context.Context, id string, channel AlertChannel) (*AlertChannel, error) {
	var result AlertChannel
	err := c.doRequest(ctx, http.MethodPut, "/alert-channels/"+url.PathEscape(id), channel, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *SporkClient) DeleteAlertChannel(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, "/alert-channels/"+url.PathEscape(id), nil, nil)
}

// doListRequest performs a GET request and unwraps the list envelope {"data": [...], "meta": {...}}.
func (c *SporkClient) doListRequest(ctx context.Context, path string, result interface{}) error {
	var jsonBytes []byte
	url := c.BaseURL + path
	_ = jsonBytes // suppress unused warning

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", c.UserAgent)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
					if seconds > maxRetryAfter {
						seconds = maxRetryAfter
					}
					if seconds > 0 {
						select {
						case <-ctx.Done():
							return ctx.Err()
						case <-time.After(time.Duration(seconds) * time.Second):
						}
					}
				}
			}
			lastErr = fmt.Errorf("API error (HTTP %d): transient error, retrying", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			return c.handleResponse(resp.StatusCode, respBody, nil)
		}

		// Parse list envelope
		if result != nil && len(respBody) > 0 {
			var envelope listEnvelope
			if err := json.Unmarshal(respBody, &envelope); err != nil {
				return fmt.Errorf("failed to unmarshal response envelope: %w", err)
			}
			if err := json.Unmarshal(envelope.Data, result); err != nil {
				return fmt.Errorf("failed to unmarshal response data: %w", err)
			}
		}
		return nil
	}
	return fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

// ListMonitors returns all monitors for the authenticated user.
func (c *SporkClient) ListMonitors(ctx context.Context) ([]Monitor, error) {
	var result []Monitor
	err := c.doListRequest(ctx, "/monitors?per_page=100", &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListAlertChannels returns all alert channels for the authenticated user.
func (c *SporkClient) ListAlertChannels(ctx context.Context) ([]AlertChannel, error) {
	var result []AlertChannel
	err := c.doListRequest(ctx, "/alert-channels?per_page=100", &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// StatusPage CRUD

func (c *SporkClient) CreateStatusPage(ctx context.Context, page StatusPage) (*StatusPage, error) {
	var result StatusPage
	err := c.doRequest(ctx, http.MethodPost, "/status-pages", page, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *SporkClient) GetStatusPage(ctx context.Context, id string) (*StatusPage, error) {
	var result StatusPage
	err := c.doRequest(ctx, http.MethodGet, "/status-pages/"+url.PathEscape(id), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *SporkClient) UpdateStatusPage(ctx context.Context, id string, page StatusPage) (*StatusPage, error) {
	var result StatusPage
	err := c.doRequest(ctx, http.MethodPut, "/status-pages/"+url.PathEscape(id), page, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *SporkClient) DeleteStatusPage(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, "/status-pages/"+url.PathEscape(id), nil, nil)
}

func (c *SporkClient) ListStatusPages(ctx context.Context) ([]StatusPage, error) {
	var result []StatusPage
	err := c.doListRequest(ctx, "/status-pages?per_page=100", &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Custom domain management

func (c *SporkClient) SetCustomDomain(ctx context.Context, statusPageID, domain string) error {
	body := map[string]string{"domain": domain}
	return c.doRequest(ctx, http.MethodPost, "/status-pages/"+url.PathEscape(statusPageID)+"/custom-domain", body, nil)
}

func (c *SporkClient) RemoveCustomDomain(ctx context.Context, statusPageID string) error {
	return c.doRequest(ctx, http.MethodDelete, "/status-pages/"+url.PathEscape(statusPageID)+"/custom-domain", nil, nil)
}
