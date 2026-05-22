package provider

import (
	"context"
	"fmt"

	"github.com/ahmedali6/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ScheduleResource{}
var _ resource.ResourceWithImportState = &ScheduleResource{}

func NewScheduleResource() resource.Resource {
	return &ScheduleResource{}
}

type ScheduleResource struct {
	client *client.DokployClient
}

type ScheduleResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	CronExpression types.String `tfsdk:"cron_expression"`
	Command        types.String `tfsdk:"command"`
	ShellType      types.String `tfsdk:"shell_type"`
	ScheduleType   types.String `tfsdk:"schedule_type"`
	ApplicationID  types.String `tfsdk:"application_id"`
	ComposeID      types.String `tfsdk:"compose_id"`
	ServerID       types.String `tfsdk:"server_id"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Timezone       types.String `tfsdk:"timezone"`
}

func (r *ScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (r *ScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy scheduled task.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Schedule ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the scheduled task.",
			},
			"cron_expression": schema.StringAttribute{
				Required:    true,
				Description: "Cron expression (e.g., '0 2 * * 5').",
			},
			"command": schema.StringAttribute{
				Required:    true,
				Description: "Command to execute.",
			},
			"shell_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("bash"),
				Description: "Shell type: 'bash' or 'sh'.",
			},
			"schedule_type": schema.StringAttribute{
				Required:    true,
				Description: "Schedule type: 'application', 'compose', 'server', or 'dokploy-server'.",
			},
			"application_id": schema.StringAttribute{
				Optional:    true,
				Description: "Application ID (for schedule_type = 'application').",
			},
			"compose_id": schema.StringAttribute{
				Optional:    true,
				Description: "Compose ID (for schedule_type = 'compose').",
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Server ID (for schedule_type = 'server').",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the schedule is enabled.",
			},
			"timezone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("UTC"),
				Description: "Timezone for the cron expression.",
			},
		},
	}
}

func (r *ScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *client.DokployClient")
		return
	}
	r.client = c
}

func (r *ScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedule := client.Schedule{
		Name:           plan.Name.ValueString(),
		CronExpression: plan.CronExpression.ValueString(),
		Command:        plan.Command.ValueString(),
		ShellType:      plan.ShellType.ValueString(),
		ScheduleType:   plan.ScheduleType.ValueString(),
		ApplicationID:  plan.ApplicationID.ValueString(),
		ComposeID:      plan.ComposeID.ValueString(),
		ServerID:       plan.ServerID.ValueString(),
		Enabled:        plan.Enabled.ValueBool(),
		Timezone:       plan.Timezone.ValueString(),
	}

	result, err := r.client.CreateSchedule(schedule)
	if err != nil {
		resp.Diagnostics.AddError("Error creating schedule", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ScheduleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedule, err := r.client.GetSchedule(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading schedule", err.Error())
		return
	}

	state.Name = types.StringValue(schedule.Name)
	state.CronExpression = types.StringValue(schedule.CronExpression)
	state.Command = types.StringValue(schedule.Command)
	state.ShellType = types.StringValue(schedule.ShellType)
	state.ScheduleType = types.StringValue(schedule.ScheduleType)
	if schedule.ApplicationID != "" {
		state.ApplicationID = types.StringValue(schedule.ApplicationID)
	}
	if schedule.ComposeID != "" {
		state.ComposeID = types.StringValue(schedule.ComposeID)
	}
	if schedule.ServerID != "" {
		state.ServerID = types.StringValue(schedule.ServerID)
	}
	state.Enabled = types.BoolValue(schedule.Enabled)
	state.Timezone = types.StringValue(schedule.Timezone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedule := client.Schedule{
		ScheduleID:     plan.ID.ValueString(),
		Name:           plan.Name.ValueString(),
		CronExpression: plan.CronExpression.ValueString(),
		Command:        plan.Command.ValueString(),
		ShellType:      plan.ShellType.ValueString(),
		ScheduleType:   plan.ScheduleType.ValueString(),
		ApplicationID:  plan.ApplicationID.ValueString(),
		ComposeID:      plan.ComposeID.ValueString(),
		ServerID:       plan.ServerID.ValueString(),
		Enabled:        plan.Enabled.ValueBool(),
		Timezone:       plan.Timezone.ValueString(),
	}

	if err := r.client.UpdateSchedule(schedule); err != nil {
		resp.Diagnostics.AddError("Error updating schedule", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSchedule(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting schedule", err.Error())
		return
	}
}

func (r *ScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	schedule, err := r.client.GetSchedule(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing schedule", fmt.Sprintf("Could not import schedule %s: %s", req.ID, err.Error()))
		return
	}

	var state ScheduleResourceModel
	state.ID = types.StringValue(schedule.ScheduleID)
	state.Name = types.StringValue(schedule.Name)
	state.CronExpression = types.StringValue(schedule.CronExpression)
	state.Command = types.StringValue(schedule.Command)
	state.ShellType = types.StringValue(schedule.ShellType)
	state.ScheduleType = types.StringValue(schedule.ScheduleType)
	if schedule.ApplicationID != "" {
		state.ApplicationID = types.StringValue(schedule.ApplicationID)
	}
	if schedule.ComposeID != "" {
		state.ComposeID = types.StringValue(schedule.ComposeID)
	}
	if schedule.ServerID != "" {
		state.ServerID = types.StringValue(schedule.ServerID)
	}
	state.Enabled = types.BoolValue(schedule.Enabled)
	state.Timezone = types.StringValue(schedule.Timezone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
