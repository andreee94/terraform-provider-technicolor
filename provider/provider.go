package provider

import (
	"context"
	"os"
	"strconv"
	"terraform-provider-technicolor/technicolor"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var stderr = os.Stderr

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

type provider struct {
	configured bool
	version    string
	router     *technicolor.TechnicolorRouter
}

// func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
// 	return tfsdk.Schema{
// 		Attributes: map[string]tfsdk.Attribute{
// 			"example": {
// 				MarkdownDescription: "Example provider attribute",
// 				Optional:            true,
// 				Type:                types.StringType,
// 			},
// 		},
// 	}, nil
// }

// GetSchema
func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"host": {
				Type:        types.StringType,
				Description: "The hostname of the Technicolor router",
				Optional:    false,
				Required:    true,
				Computed:    false,
			},
			"port": {
				Type:        types.Int64Type,
				Description: "The port of the Technicolor router (Default: 80)",
				Optional:    true,
				Required:    false,
			},
			"username": {
				Type:     types.StringType,
				Optional: false,
				Computed: false,
				Required: true,
			},
			"password": {
				Type:      types.StringType,
				Optional:  false,
				Computed:  false,
				Sensitive: true,
				Required:  true,
			},
		},
	}, nil
}

// Provider schema struct
type providerData struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var host string
	var port int
	var username string
	var password string

	if !checkForUnknowsInConfig(&config, resp) {
		return
	}

	if config.Host.Null {
		host = os.Getenv("TECHNICOLOR_HOST")
	} else {
		host = config.Host.Value
	}

	if config.Username.Null {
		username = os.Getenv("TECHNICOLOR_USERNAME")
	} else {
		username = config.Username.Value
	}

	if config.Port.Null {
		portString := os.Getenv("TECHNICOLOR_PORT")
		if portString == "" {
			port = 80
		} else {
			port, err = strconv.Atoi(portString)
			if err != nil {
				resp.Diagnostics.AddError(
					"Invalid port",
					"The port must be an integer",
				)
				return
			}
		}
	} else {
		port = int(config.Port.Value)
	}

	if config.Password.Null {
		password = os.Getenv("TECHNICOLOR_PASSWORD")
	} else {
		password = config.Password.Value
	}

	if username == "" {
		resp.Diagnostics.AddError(
			"Unable to find username",
			"Username cannot be an empty string",
		)
		return
	}

	p.router = technicolor.NewTechnicolorRouter(host, port)

	err, isAuthenticated := p.router.Login(username, password)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to login",
			err.Error(),
		)
		return
	}

	if !isAuthenticated {
		resp.Diagnostics.AddError(
			"Unable to login",
			"Username or password is incorrect",
		)
		return
	}

	p.configured = true
}

func checkForUnknowsInConfig(config *providerData, resp *tfsdk.ConfigureProviderResponse) bool {
	if config.Host.Unknown {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as host",
		)
		return false
	}

	if config.Username.Unknown {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as username",
		)
		return false
	}

	if config.Password.Unknown {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as password",
		)
		return false
	}
	return true
}

func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		// "freenom_dns_record": resourceFreenomDnsRecordType{},
	}, nil
}

func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"technicolor_port_forwarded_list": datasourcePortForwardedListType{},
	}, nil
}
