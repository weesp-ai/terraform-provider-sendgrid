package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := New("test-key")
	c.BaseURL = srv.URL
	return c
}

func TestListDomains(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/whitelabel/domains" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("missing auth header: %q", got)
		}
		_ = json.NewEncoder(w).Encode([]Domain{
			{ID: 1, Domain: "a.example.com"},
			{ID: 2, Domain: "b.example.com"},
		})
	})

	got, err := c.ListDomains(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 domains, got %d", len(got))
	}
}

func TestFindDomainByHostname(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]Domain{{ID: 1, Domain: "match.example.com"}})
	})

	d, err := c.FindDomainByHostname(context.Background(), "match.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if d == nil || d.ID != 1 {
		t.Fatalf("want match, got %+v", d)
	}

	miss, err := c.FindDomainByHostname(context.Background(), "no-such.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if miss != nil {
		t.Fatalf("want nil, got %+v", miss)
	}
}

func TestGetDomain404(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	d, err := c.GetDomain(context.Background(), 42)
	if err != nil {
		t.Fatalf("404 should not be an error, got %v", err)
	}
	if d != nil {
		t.Fatalf("want nil, got %+v", d)
	}
}

func TestCreateDomain(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("want POST, got %s", r.Method)
		}
		var got CreateDomainRequest
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		if got.Domain != "new.example.com" {
			t.Errorf("want new.example.com, got %s", got.Domain)
		}
		_ = json.NewEncoder(w).Encode(Domain{
			ID: 99, Domain: got.Domain,
			DNS: DNS{MailCname: DNSRecord{Host: "em.new", Data: "u.sendgrid.net"}},
		})
	})

	d, err := c.CreateDomain(context.Background(), CreateDomainRequest{
		Domain: "new.example.com", AutomaticSecurity: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != 99 || d.DNS.MailCname.Host != "em.new" {
		t.Fatalf("unexpected response: %+v", d)
	}
}

func TestInboundParseRuleCRUD(t *testing.T) {
	var lastMethod, lastPath string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		lastMethod, lastPath = r.Method, r.URL.Path
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/v3/user/webhooks/parse/settings/missing.example.com" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(InboundParseRule{
				Hostname: "x.example.com", URL: "https://api/x", SendRaw: true,
			})
		default:
			w.WriteHeader(http.StatusOK)
		}
	})

	if err := c.CreateInboundParseRule(context.Background(), InboundParseRule{
		Hostname: "x.example.com", URL: "https://api/x", SendRaw: true,
	}); err != nil {
		t.Fatal(err)
	}
	if lastMethod != http.MethodPost || lastPath != "/v3/user/webhooks/parse/settings" {
		t.Errorf("create: bad path %s %s", lastMethod, lastPath)
	}

	got, err := c.GetInboundParseRule(context.Background(), "x.example.com")
	if err != nil || got == nil || got.Hostname != "x.example.com" {
		t.Fatalf("get: err=%v got=%+v", err, got)
	}

	missing, err := c.GetInboundParseRule(context.Background(), "missing.example.com")
	if err != nil {
		t.Fatalf("404 should not be error: %v", err)
	}
	if missing != nil {
		t.Fatalf("want nil for missing, got %+v", missing)
	}

	if err := c.UpdateInboundParseRule(context.Background(), InboundParseRule{
		Hostname: "x.example.com", URL: "https://api/x2", SendRaw: true,
	}); err != nil {
		t.Fatal(err)
	}
	if lastMethod != http.MethodPatch {
		t.Errorf("update: want PATCH, got %s", lastMethod)
	}

	if err := c.DeleteInboundParseRule(context.Background(), "x.example.com"); err != nil {
		t.Fatal(err)
	}
	if lastMethod != http.MethodDelete {
		t.Errorf("delete: want DELETE, got %s", lastMethod)
	}
}
