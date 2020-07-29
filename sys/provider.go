package sys

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"sys_file":         resourceFile(),
			"sys_dir":          resourceDir(),
			"sys_shell_script": resourceShellScript(),
			"sys_symlink":      resourceSymlink(),
			"sys_null":         resourceNull(),
			"sys_package":      resourcePackage(),
			"sys_systemd_unit": resourceSystemdUnit(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"sys_file": dataSourceFile(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

type providerConfiguration struct {
	debUpdated bool
}

func providerConfigure(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	configuration := &providerConfiguration{
	}
	return configuration, nil
}
