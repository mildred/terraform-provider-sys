package sys

import (
	"context"
	"sync"

	hclog "github.com/hashicorp/go-hclog"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mildred/terraform-provider-sys/sys/data_uname"
)

type sysProvider struct {
	debUpdated bool
	Logger     hclog.Logger
	SdLocks    map[string]sync.Locker
	Lock       sync.Mutex
}

type sysProviderModel struct {
	LogLevel types.String `tfsdk:"log_level"`
}

func New() provider.Provider {
	return &sysProvider{}
}

func (p *sysProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sys"
}

func (p *sysProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"log_level": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (p *sysProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data sysProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	log_level := data.LogLevel.ValueString()
	if log_level == "" {
		log_level = "info"
	}

	p.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.LevelFromString(log_level),
	})
}

func (p *sysProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource {
	}
}

func (p *sysProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource {
		data_uname.NewDataSource,
	}
}
