package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// providerFactories is wired into TestCase.ProtoV6ProviderFactories by
// acceptance tests (TF_ACC=1). Acceptance tests are not in this commit;
// the factory is exposed so they can be added later without code churn.
func providerFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"sendgrid": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func TestProviderConstruct(t *testing.T) {
	p := New("test")()
	if p == nil {
		t.Fatal("New returned nil provider")
	}
	if got := p.Resources(context.Background()); len(got) != 3 {
		t.Errorf("want 3 resources, got %d", len(got))
	}
	_ = providerFactories()
}
