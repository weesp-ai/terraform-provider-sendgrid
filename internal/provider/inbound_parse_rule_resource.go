package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/weesp-ai/terraform-provider-sendgrid/internal/client"
)

var (
	_ resource.Resource                = (*inboundParseRuleResource)(nil)
	_ resource.ResourceWithImportState = (*inboundParseRuleResource)(nil)
	_ resource.ResourceWithConfigure   = (*inboundParseRuleResource)(nil)
)

type inboundParseRuleResource struct {
	client *client.Client
}

type inboundParseRuleModel struct {
	ID        types.String `tfsdk:"id"`
	Hostname  types.String `tfsdk:"hostname"`
	URL       types.String `tfsdk:"url"`
	SpamCheck types.Bool   `tfsdk:"spam_check"`
	SendRaw   types.Bool   `tfsdk:"send_raw"`
}

// NewInboundParseRuleResource is the constructor referenced by Provider.Resources.
func NewInboundParseRuleResource() resource.Resource {
	return &inboundParseRuleResource{}
}

func (r *inboundParseRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inbound_parse_rule"
}

func (r *inboundParseRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *inboundParseRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A SendGrid Inbound Parse webhook rule that forwards mail addressed to a hostname to an HTTPS URL. SendGrid permits at most one rule per hostname; the provider adopts an existing rule with the same hostname instead of failing.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Same as `hostname`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"hostname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Mail-flow hostname (e.g. `mail.example.com`). Changing this forces a new resource.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "HTTPS URL SendGrid POSTs the inbound mail to.",
			},
			"spam_check": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether SendGrid runs SpamAssassin on inbound mail before forwarding. Defaults to `false`.",
			},
			"send_raw": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether SendGrid forwards the original RFC 822 message as `email`. Defaults to `true`.",
			},
		},
	}
}

func (r *inboundParseRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan inboundParseRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := modelToRule(plan)
	existing, err := r.client.GetInboundParseRule(ctx, rule.Hostname)
	if err != nil {
		resp.Diagnostics.AddError("Check existing parse rule failed", err.Error())
		return
	}
	if existing != nil {
		if !rulesEqual(*existing, rule) {
			if err := r.client.UpdateInboundParseRule(ctx, rule); err != nil {
				resp.Diagnostics.AddError("Adopt existing parse rule failed", err.Error())
				return
			}
		}
	} else {
		if err := r.client.CreateInboundParseRule(ctx, rule); err != nil {
			resp.Diagnostics.AddError("Create parse rule failed", err.Error())
			return
		}
	}

	plan.ID = plan.Hostname
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *inboundParseRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state inboundParseRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetInboundParseRule(ctx, state.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read parse rule failed", err.Error())
		return
	}
	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.URL = types.StringValue(rule.URL)
	state.SpamCheck = types.BoolValue(rule.SpamCheck)
	state.SendRaw = types.BoolValue(rule.SendRaw)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *inboundParseRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan inboundParseRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateInboundParseRule(ctx, modelToRule(plan)); err != nil {
		resp.Diagnostics.AddError("Update parse rule failed", err.Error())
		return
	}
	plan.ID = plan.Hostname
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *inboundParseRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state inboundParseRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteInboundParseRule(ctx, state.Hostname.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete parse rule failed", err.Error())
	}
}

func (r *inboundParseRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hostname"), req.ID)...)
}

func modelToRule(m inboundParseRuleModel) client.InboundParseRule {
	return client.InboundParseRule{
		Hostname:  m.Hostname.ValueString(),
		URL:       m.URL.ValueString(),
		SpamCheck: m.SpamCheck.ValueBool(),
		SendRaw:   m.SendRaw.ValueBool(),
	}
}

func rulesEqual(a, b client.InboundParseRule) bool {
	return a.Hostname == b.Hostname && a.URL == b.URL && a.SpamCheck == b.SpamCheck && a.SendRaw == b.SendRaw
}
