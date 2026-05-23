// Package client is a minimal HTTP client for the SendGrid v3 endpoints used by this provider:
// Authenticated Domains and Inbound Parse Settings.
//
// The JSON shapes match SendGrid's API surface verbatim; if SendGrid changes either,
// this package must change in lockstep.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
)

const defaultBaseURL = "https://api.sendgrid.com"

// Client wraps an *http.Client with a SendGrid API key.
type Client struct {
	APIKey  string
	BaseURL string
	HTTP    *http.Client
}

// New returns a Client with a 30s timeout against the public SendGrid API.
func New(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: defaultBaseURL,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// --- Authenticated Domains ---------------------------------------------------

// DNSRecord is a single DNS record SendGrid expects to be present for a whitelabel domain.
type DNSRecord struct {
	Host string `json:"host"`
	Type string `json:"type"`
	Data string `json:"data"`
	TTL  int    `json:"ttl,omitempty"`
}

// DNS bundles the three DNS records SendGrid generates for an authenticated domain.
type DNS struct {
	MailCname DNSRecord `json:"mail_cname"`
	DKIM1     DNSRecord `json:"dkim1"`
	DKIM2     DNSRecord `json:"dkim2"`
}

// Domain mirrors SendGrid's whitelabel domain object.
type Domain struct {
	ID        int64  `json:"id"`
	Domain    string `json:"domain"`
	Subdomain string `json:"subdomain"`
	DNS       DNS    `json:"dns"`
	Valid     bool   `json:"valid"`
}

// CreateDomainRequest is the body for POST /v3/whitelabel/domains.
type CreateDomainRequest struct {
	Domain            string `json:"domain"`
	Subdomain         string `json:"subdomain,omitempty"`
	AutomaticSecurity bool   `json:"automatic_security"`
	CustomSPF         bool   `json:"custom_spf"`
	Default           bool   `json:"default"`
}

// ListDomains returns up to 500 SendGrid Authenticated Domains in one call.
// SG provides no direct lookup-by-name, so callers list and filter client-side.
func (c *Client) ListDomains(ctx context.Context) ([]Domain, error) {
	var out []Domain
	if err := c.do(ctx, http.MethodGet, "/v3/whitelabel/domains?limit=500", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// FindDomainByHostname is a convenience wrapper over ListDomains.
func (c *Client) FindDomainByHostname(ctx context.Context, hostname string) (*Domain, error) {
	domains, err := c.ListDomains(ctx)
	if err != nil {
		return nil, err
	}
	for i := range domains {
		if domains[i].Domain == hostname {
			return &domains[i], nil
		}
	}
	return nil, nil
}

// CreateDomain creates a SendGrid Authenticated Domain. The returned Domain
// includes the SG-generated DKIM/CNAME records that must be published in DNS
// before validation will succeed.
func (c *Client) CreateDomain(ctx context.Context, req CreateDomainRequest) (*Domain, error) {
	var out Domain
	if err := c.do(ctx, http.MethodPost, "/v3/whitelabel/domains", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDomain fetches a domain by id. Returns (nil, nil) on 404.
func (c *Client) GetDomain(ctx context.Context, id int64) (*Domain, error) {
	var out Domain
	found, err := c.doMaybe404(ctx, http.MethodGet, fmt.Sprintf("/v3/whitelabel/domains/%d", id), nil, &out)
	if err != nil || !found {
		return nil, err
	}
	return &out, nil
}

// ValidateDomain triggers SG's DNS-record validation for the given domain id.
// Returns the post-validation Domain (check .Valid).
func (c *Client) ValidateDomain(ctx context.Context, id int64) (*Domain, error) {
	var out Domain
	if err := c.do(ctx, http.MethodPost, fmt.Sprintf("/v3/whitelabel/domains/%d/validate", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteDomain removes the domain from SendGrid. 404 is treated as success.
func (c *Client) DeleteDomain(ctx context.Context, id int64) error {
	_, err := c.doMaybe404(ctx, http.MethodDelete, fmt.Sprintf("/v3/whitelabel/domains/%d", id), nil, nil)
	return err
}

// --- Inbound Parse Rules -----------------------------------------------------

// InboundParseRule mirrors the SG /v3/user/webhooks/parse/settings object.
type InboundParseRule struct {
	Hostname  string `json:"hostname"`
	URL       string `json:"url"`
	SpamCheck bool   `json:"spam_check"`
	SendRaw   bool   `json:"send_raw"`
}

// CreateInboundParseRule registers a new parse rule for the given hostname.
func (c *Client) CreateInboundParseRule(ctx context.Context, rule InboundParseRule) error {
	return c.do(ctx, http.MethodPost, "/v3/user/webhooks/parse/settings", rule, nil)
}

// GetInboundParseRule looks up the rule by hostname. Returns (nil, nil) on 404.
func (c *Client) GetInboundParseRule(ctx context.Context, hostname string) (*InboundParseRule, error) {
	var out InboundParseRule
	found, err := c.doMaybe404(ctx, http.MethodGet, fmt.Sprintf("/v3/user/webhooks/parse/settings/%s", hostname), nil, &out)
	if err != nil || !found {
		return nil, err
	}
	return &out, nil
}

// inboundParseRuleUpdate is the PATCH body for UpdateInboundParseRule.
// The SendGrid API rejects "hostname" as an unknown field when it appears in
// the body (it is a URL path parameter only), so we use a separate struct.
type inboundParseRuleUpdate struct {
	URL       string `json:"url"`
	SpamCheck bool   `json:"spam_check"`
	SendRaw   bool   `json:"send_raw"`
}

// UpdateInboundParseRule patches the rule for an existing hostname.
func (c *Client) UpdateInboundParseRule(ctx context.Context, rule InboundParseRule) error {
	body := inboundParseRuleUpdate{
		URL:       rule.URL,
		SpamCheck: rule.SpamCheck,
		SendRaw:   rule.SendRaw,
	}
	return c.do(ctx, http.MethodPatch, fmt.Sprintf("/v3/user/webhooks/parse/settings/%s", rule.Hostname), body, nil)
}

// DeleteInboundParseRule removes the parse rule. 404 is treated as success.
func (c *Client) DeleteInboundParseRule(ctx context.Context, hostname string) error {
	_, err := c.doMaybe404(ctx, http.MethodDelete, fmt.Sprintf("/v3/user/webhooks/parse/settings/%s", hostname), nil, nil)
	return err
}

// --- internal HTTP plumbing --------------------------------------------------

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	_, err := c.doMaybe404(ctx, method, path, body, out)
	return err
}

// doMaybe404 performs an HTTP request and treats 404 as "not found".
// Returns (true, nil) on 2xx, (false, nil) on 404, (false, err) otherwise.
// Retries on 429/503 with exponential backoff (3s → 30s, capped at 5 minutes total).
func (c *Client) doMaybe404(ctx context.Context, method, path string, body, out any) (bool, error) {
	var bodyBuf []byte
	if body != nil {
		var err error
		bodyBuf, err = json.Marshal(body)
		if err != nil {
			return false, fmt.Errorf("marshal %s %s: %w", method, path, err)
		}
	}

	var (
		found    bool
		respOut  []byte
		lastCode int
	)

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 3 * time.Second
	b.MaxInterval = 30 * time.Second
	b.MaxElapsedTime = 5 * time.Minute

	err := backoff.Retry(func() error {
		var bodyReader io.Reader
		if bodyBuf != nil {
			bodyReader = bytes.NewReader(bodyBuf)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
		if err != nil {
			return backoff.Permanent(fmt.Errorf("build %s %s: %w", method, path, err))
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.HTTP.Do(req)
		if err != nil {
			return backoff.Permanent(fmt.Errorf("%s %s: %w", method, path, err))
		}

		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		lastCode = resp.StatusCode

		if resp.StatusCode == http.StatusNotFound {
			found = false
			return nil
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			return fmt.Errorf("%s %s: status=%d body=%s", method, path, resp.StatusCode, string(respBody))
		}

		if resp.StatusCode >= 400 {
			return backoff.Permanent(fmt.Errorf("%s %s: status=%d body=%s", method, path, resp.StatusCode, string(respBody)))
		}

		found = true
		respOut = respBody
		return nil
	}, backoff.WithContext(b, ctx))

	if err != nil {
		return false, err
	}
	if lastCode == http.StatusNotFound {
		return false, nil
	}

	if out != nil && len(respOut) > 0 {
		if err := json.Unmarshal(respOut, out); err != nil {
			return false, fmt.Errorf("unmarshal %s %s: %w", method, path, err)
		}
	}
	return found, nil
}
