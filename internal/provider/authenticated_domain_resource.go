package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/weesp-ai/terraform-provider-sendgrid/internal/client"
)

var (
	_ resource.Resource                = (*authenticatedDomainResource)(nil)
	_ resource.ResourceWithImportState = (*authenticatedDomainResource)(nil)
	_ resource.ResourceWithConfigure   = (*authenticatedDomainResource)(nil)
)

type authenticatedDomainResource struct {
	client *client.Client
}

type authenticatedDomainModel struct {
	ID                types.String `tfsdk:"id"`
	Domain            types.String `tfsdk:"domain"`
	AutomaticSecurity types.Bool   `tfsdk:"automatic_security"`
	CustomSPF         types.Bool   `tfsdk:"custom_spf"`
	IsDefault         types.Bool   `tfsdk:"is_default"`
	DomainID          types.Int64  `tfsdk:"domain_id"`
	Valid             types.Bool   `tfsdk:"valid"`
	MailCnameHost     types.String `tfsdk:"mail_cname_host"`
	MailCnameData     types.String `tfsdk:"mail_cname_data"`
	DKIM1Host         types.String `tfsdk:"dkim1_host"`
	DKIM1Data         types.String `tfsdk:"dkim1_data"`
	DKIM2Host         types.String `tfsdk:"dkim2_host"`
	DKIM2Data         types.String `tfsdk:"dkim2_data"`
}

// NewAuthenticatedDomainResource is the constructor referenced by Provider.Resources.
func NewAuthenticatedDomainResource() resource.Resource {
	return &authenticatedDomainResource{}
}

func (r *authenticatedDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authenticated_domain"
}

func (r *authenticatedDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *authenticatedDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A SendGrid Authenticated Domain (whitelabel) with DKIM/CNAME DNS attributes. SendGrid domains are immutable; changes to any input force a replacement. The provider adopts an existing domain with the same hostname instead of failing on conflict.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The SendGrid domain ID, as a string.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Fully-qualified hostname to authenticate (e.g. `mail.example.com`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"automatic_security": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether SendGrid manages the DKIM keys and CNAMEs. Defaults to `true`.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"custom_spf": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether SPF records are managed by the user. Defaults to `false`.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"is_default": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether this is the SendGrid account's default domain. Defaults to `false`.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"domain_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The SendGrid domain ID as an integer.",
				PlanModifiers:       []planmodifier.Int64{},
			},
			"valid":           computedBool("Whether SendGrid considers this domain DNS-validated."),
			"mail_cname_host": computedString("DNS host for the SendGrid `em.*` mail CNAME."),
			"mail_cname_data": computedString("DNS target for the mail CNAME."),
			"dkim1_host":      computedString("DNS host for the first DKIM CNAME."),
			"dkim1_data":      computedString("DNS target for the first DKIM CNAME."),
			"dkim2_host":      computedString("DNS host for the second DKIM CNAME."),
			"dkim2_data":      computedString("DNS target for the second DKIM CNAME."),
		},
	}
}

func computedString(desc string) schema.StringAttribute {
	return schema.StringAttribute{Computed: true, MarkdownDescription: desc}
}

func computedBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{Computed: true, MarkdownDescription: desc}
}

func (r *authenticatedDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan authenticatedDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Idempotent: if a domain with this hostname already exists in SG, adopt it.
	domain, err := r.client.FindDomainByHostname(ctx, plan.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Lookup existing domain failed", err.Error())
		return
	}
	if domain == nil {
		domain, err = r.client.CreateDomain(ctx, client.CreateDomainRequest{
			Domain:            plan.Domain.ValueString(),
			AutomaticSecurity: plan.AutomaticSecurity.ValueBool(),
			CustomSPF:         plan.CustomSPF.ValueBool(),
			Default:           plan.IsDefault.ValueBool(),
		})
		if err != nil {
			resp.Diagnostics.AddError("Create domain failed", err.Error())
			return
		}
	}

	applyDomainToModel(domain, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *authenticatedDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state authenticatedDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid stored domain ID", err.Error())
		return
	}

	domain, err := r.client.GetDomain(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Read domain failed", err.Error())
		return
	}
	if domain == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	applyDomainToModel(domain, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *authenticatedDomainResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All inputs carry RequiresReplace; the framework will never call Update.
	resp.Diagnostics.AddError("Update not supported", "All authenticated_domain inputs are immutable; changes force a replacement.")
}

func (r *authenticatedDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state authenticatedDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid stored domain ID", err.Error())
		return
	}
	if err := r.client.DeleteDomain(ctx, id); err != nil {
		resp.Diagnostics.AddError("Delete domain failed", err.Error())
	}
}

func (r *authenticatedDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func applyDomainToModel(d *client.Domain, m *authenticatedDomainModel) {
	m.ID = types.StringValue(strconv.FormatInt(d.ID, 10))
	m.DomainID = types.Int64Value(d.ID)
	m.Domain = types.StringValue(d.Domain)
	m.Valid = types.BoolValue(d.Valid)
	m.MailCnameHost = types.StringValue(d.DNS.MailCname.Host)
	m.MailCnameData = types.StringValue(d.DNS.MailCname.Data)
	m.DKIM1Host = types.StringValue(d.DNS.DKIM1.Host)
	m.DKIM1Data = types.StringValue(d.DNS.DKIM1.Data)
	m.DKIM2Host = types.StringValue(d.DNS.DKIM2.Host)
	m.DKIM2Data = types.StringValue(d.DNS.DKIM2.Data)
}
