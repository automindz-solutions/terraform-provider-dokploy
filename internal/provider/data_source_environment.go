package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &EnvironmentDataSource{}

func NewEnvironmentDataSource() datasource.DataSource {
	return &EnvironmentDataSource{}
}

type EnvironmentDataSource struct {
	client *client.DokployClient
}

type EnvironmentDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (d *EnvironmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *EnvironmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing Dokploy environment by project ID and name. Useful for referencing the auto-created 'production' environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Environment ID.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "Project ID to search within.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Environment name to look up (e.g., 'production').",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Environment description.",
			},
		},
	}
}

func (d *EnvironmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type", "Expected *client.DokployClient")
		return
	}
	d.client = c
}

func (d *EnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config EnvironmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := config.ProjectID.ValueString()
	name := config.Name.ValueString()

	envs, err := d.client.ListEnvironmentsByProject(projectID)
	if err != nil {
		resp.Diagnostics.AddError("Error listing environments", fmt.Sprintf("Could not list environments for project %s: %s", projectID, err.Error()))
		return
	}

	for _, env := range envs {
		if env.Name == name {
			config.ID = types.StringValue(env.ID)
			config.Description = types.StringValue(env.Description)
			resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Environment not found",
		fmt.Sprintf("No environment named %q found in project %s", name, projectID),
	)
}
