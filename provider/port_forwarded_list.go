package provider

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type datasourcePortForwardedListType struct{}

func (c datasourcePortForwardedListType) GetSchema(_ context.Context) (tfsdk.Schema,
	diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"ports": {
				// When Computed is true, the provider will set value --
				// the user cannot define the value
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Type:     types.StringType,
						Computed: true,
						Required: false,
					},
					"index": {
						Type:     types.Int64Type,
						Computed: true,
						Required: false,
					},
					"enabled": {
						Type:     types.BoolType,
						Computed: true,
						Required: false,
					},
					"name": {
						Type:     types.StringType,
						Computed: true,
						Required: false,
					},
					"protocol": {
						Type:     types.StringType,
						Computed: true,
						Required: false,
					},
					"wan_port": {
						Type:     types.Int64Type,
						Computed: true,
						Required: false,
					},
					"lan_port": {
						Type:     types.Int64Type,
						Computed: true,
						Required: false,
					},
					"lan_ip": {
						Type:     types.StringType,
						Computed: true,
						Required: false,
					},
					"lan_mac": {
						Type:     types.StringType,
						Computed: true,
						Required: false,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
		},
	}, nil
}

func (c datasourcePortForwardedListType) NewDataSource(_ context.Context,
	p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return datasourcePortForwardedList{
		p: *(p.(*provider)),
	}, nil
}

type datasourcePortForwardedList struct {
	p provider
}

func (r datasourcePortForwardedList) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var resourceState struct {
		Ports []PortForwarded `tfsdk:"ports"`
	}

	diags := req.Config.Get(ctx, &resourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err, ports = r.p.router.GetAllPortForwarded()

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get port forwarded list",
			"Failed to get port forwarded list from router",
		)
		return
	}

	log.Printf("[INFO] Found %d forwarded ports", len(ports))

	for _, port := range ports {
		resourceState.Ports = append(resourceState.Ports, PortForwarded{
			// ID:       types.String(port.Data.ID),
			ID:       types.String{Value: computeID(port.Data.Name, port.Data.WanPort, port.Data.Protocol)},
			Index:    types.Int64{Value: int64(port.Index)},
			Enabled:  types.Bool{Value: port.Data.Enabled},
			Name:     types.String{Value: port.Data.Name},
			Protocol: types.String{Value: port.Data.Protocol},
			WanPort:  types.Int64{Value: int64(port.Data.WanPort)},
			LanPort:  types.Int64{Value: int64(port.Data.LanPort)},
			LanIp:    types.String{Value: port.Data.LanIp},
			LanMac:   types.String{Value: port.Data.LanMac},
		})
	}

	diags = resp.State.Set(ctx, &resourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
