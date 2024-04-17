package data_uname

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"regexp"
	"os/exec"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var uname_all_regexp *regexp.Regexp
func init(){
	uname_all_regexp = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(.+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)$`)
}

type DataSource struct {}

type Data struct {
		Id               types.String `tfsdk:"id"`
		Flag             types.String `tfsdk:"flag"`
		Output           types.String `tfsdk:"output"`
		KernelName       types.String `tfsdk:"kernel_name"`
		Nodename         types.String `tfsdk:"nodename"`
		KernelRelease    types.String `tfsdk:"kernel_release"`
		KernelVersion    types.String `tfsdk:"kernel_version"`
		Machine          types.String `tfsdk:"machine"`
		Processor        types.String `tfsdk:"processor"`
		HardwarePlatform types.String `tfsdk:"hardware_platform"`
		OperatingSystem  types.String `tfsdk:"operating_system"`
	}
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_uname"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Return values from the uname executable",

		Attributes: map[string]schema.Attribute{
			"flag": schema.StringAttribute{
				Optional: true,
				Description: `Uname flag without the dash: a: all, s: kernel name, n: nodename, r: kernel release, v: kernel version, m: machine, p: processor, i: hardware platform, o: operating system`,
			},
			"output": schema.StringAttribute{
				Computed: true,
				Description: `Output from uname command`,
			},
			"kernel_name": schema.StringAttribute{
				Computed: true,
				Description: `uname -s`,
			},
			"nodename": schema.StringAttribute{
				Computed: true,
				Description: `uname -n`,
			},
			"kernel_release": schema.StringAttribute{
				Computed: true,
				Description: `uname -r`,
			},
			"kernel_version": schema.StringAttribute{
				Computed: true,
				Description: `uname -v`,
			},
			"machine": schema.StringAttribute{
				Computed: true,
				Description: `uname -m`,
			},
			"processor": schema.StringAttribute{
				Computed: true,
				Description: `uname -p`,
			},
			"hardware_platform": schema.StringAttribute{
				Computed: true,
				Description: `uname -i`,
			},
			"operating_system": schema.StringAttribute{
				Computed: true,
				Description: `uname -o`,
			},
		},
	}
}


func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data Data

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	flag := data.Flag.ValueString()
	if len(flag) == 0 {
		flag = "-a"
	} else if flag[0] != '-' {
		if len(flag) > 1 {
			flag = "--" + flag
		} else {
			flag = "-" + flag
		}
	}

	out, err := exec.Command("uname", flag).Output()
	if err != nil {
		resp.Diagnostics.AddError("could not run uname command", err.Error())
		return
	}

	out_line := strings.Trim(string(out), "\n\r\t ")

	data.Output = types.StringValue(out_line)

	switch flag {
	case "-a", "--all":
		parts := uname_all_regexp.FindStringSubmatch(out_line)
		data.KernelName =      types.StringValue(parts[1])
		data.Nodename =         types.StringValue(parts[2])
		data.KernelRelease =    types.StringValue(parts[3])
		data.KernelVersion =    types.StringValue(parts[4])
		data.Machine =          types.StringValue(parts[5])
		data.Processor =        types.StringValue(parts[6])
		data.HardwarePlatform = types.StringValue(parts[7])
		data.OperatingSystem =  types.StringValue(parts[8])
	case "-s", "--kernel-name":
		data.KernelName =       types.StringValue(out_line)
	case "-n", "--nodename":
		data.Nodename =         types.StringValue(out_line)
	case "-r", "--kernel-release":
		data.KernelRelease =    types.StringValue(out_line)
	case "-v", "--kernel-version":
		data.KernelVersion =    types.StringValue(out_line)
	case "-m", "--machine":
		data.Machine =          types.StringValue(out_line)
	case "-p", "--processor":
		data.Processor =        types.StringValue(out_line)
	case "-i", "--hardware-platform":
		data.HardwarePlatform = types.StringValue(out_line)
	case "-o", "--operating-system":
		data.OperatingSystem =  types.StringValue(out_line)
	}

	checksum := sha1.Sum([]byte(flag + "\n" + out_line))
	data.Id = types.StringValue(hex.EncodeToString(checksum[:]))

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
