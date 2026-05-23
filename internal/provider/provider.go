// Package provider implements the Terraform SendGrid provider.
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/weesp-ai/terraform-provider-sendgrid/internal/client"
)

var _ provider.Provider = (*sendgridProvider)(nil)

type sendgridProvider struct {
	version string
}

type sendgridProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

// New returns a constructor for the SendGrid provider, suitable for passing to
// providerserver.Serve. The version is baked in by main.go from -ldflags.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &sendgridProvider{version: version}
	}
}

func (p *sendgridProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sendgrid"
	resp.Version = p.version
}

func (p *sendgridProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a focused slice of SendGrid (Authenticated Domains, their DNS validation, and Inbound Parse webhook rules).",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "SendGrid API key. Required scopes: `mail.send`, `whitelabel`, `inbound_parse`. Pass via a Terraform variable (e.g. `var.sendgrid_api_key`, set with `TF_VAR_sendgrid_api_key` in the environment) — the provider does not read API keys from arbitrary environment variables.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *sendgridProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data sendgridProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := client.New(data.APIKey.ValueString())
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *sendgridProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAuthenticatedDomainResource,
		NewDomainValidationResource,
		NewInboundParseRuleResource,
	}
}

func (p *sendgridProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
