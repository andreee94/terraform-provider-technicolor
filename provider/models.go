package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PortForwarded struct {
	ID       types.String `tfsdk:"id"`
	Index    types.Int64  `tfsdk:"index"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Name     types.String `tfsdk:"name"`
	Protocol types.String `tfsdk:"protocol"`
	WanPort  types.Int64  `tfsdk:"wan_port"`
	LanPort  types.Int64  `tfsdk:"lan_port"`
	LanIp    types.String `tfsdk:"lan_ip"`
	LanMac   types.String `tfsdk:"lan_mac"`
}
