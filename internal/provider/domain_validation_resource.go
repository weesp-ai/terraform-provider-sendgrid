package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/weesp-ai/terraform-provider-sendgrid/internal/client"
)

const (
	validationTimeout = 5 * time.Minute
	validationPoll    = 10 * time.Second
)

var (
	_ resource.Resource                = (*domainValidationResource)(nil)
	_ resource.ResourceWithImportState = (*domainValidationResource)(nil)
	_ resource.ResourceWithConfigure   = (*domainValidationResource)(nil)
)

type domainValidationResource struct {
	client *client.Client
}

type domainValidationModel struct {
	ID       types.String `tfsdk:"id"`
	DomainID types.Int64  `tfsdk:"domain_id"`
	Valid    types.Bool   `tfsdk:"valid"`
}

// NewDomainValidationResource is the constructor referenced by Provider.Resources.
func NewDomainValidationResource() resource.Resource {
	return &domainValidationResource{}
}

func (r *domainValidationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_validation"
}

func (r *domainValidationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *domainValidationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Polls SendGrid until an Authenticated Domain passes DNS validation. Place this resource downstream of the DNS records that satisfy the domain's CNAME requirements.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "ID of the SendGrid Authenticated Domain to validate.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"valid": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "True after SendGrid confirms the domain's DNS records.",
			},
		},
	}
}

func (r *domainValidationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainValidationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID := plan.DomainID.ValueInt64()
	domain, err := waitForValid(ctx, r.client, domainID)
	if err != nil {
		resp.Diagnostics.AddError("Validate domain failed", err.Error())
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(domainID, 10))
	plan.Valid = types.BoolValue(domain.Valid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainValidationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainValidationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.GetDomain(ctx, state.DomainID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Read domain failed", err.Error())
		return
	}
	if domain == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Valid = types.BoolValue(domain.Valid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *domainValidationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// domain_id carries RequiresReplace; Update is unreachable.
}

func (r *domainValidationResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Validation is a logical resource — there is nothing on SendGrid's side to delete.
}

func (r *domainValidationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "expected an integer SendGrid domain ID, got: "+req.ID)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_id"), id)...)
}

func waitForValid(ctx context.Context, c *client.Client, id int64) (*client.Domain, error) {
	deadline := time.Now().Add(validationTimeout)
	for {
		d, err := c.ValidateDomain(ctx, id)
		if err == nil && d.Valid {
			return d, nil
		}
		if time.Now().After(deadline) {
			if err != nil {
				return nil, fmt.Errorf("timed out after %s; last error: %w", validationTimeout, err)
			}
			return nil, fmt.Errorf("timed out after %s; domain still not valid", validationTimeout)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(validationPoll):
		}
	}
}
