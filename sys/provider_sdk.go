package sys

import (
	"context"
	"sync"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"log_level": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "info",
			},
		},
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
		        "sys_os_release":   dataSourceOsRelease(),
			"sys_file":         dataSourceFile(),
			"sys_shell_script": dataSourceShellScript(),
			"sys_error":        dataSourceError(),
			"uname":            dataSourceUname(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

type providerConfiguration struct {
	debUpdated bool
	Logger     hclog.Logger
	SdLocks    map[string]sync.Locker
	Lock       sync.Mutex
}

func providerConfigure(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	logger := hclog.New(&hclog.LoggerOptions{
		Level: hclog.LevelFromString(data.Get("log_level").(string)),
	})
	configuration := &providerConfiguration{
		Logger: logger,
	}
	return configuration, nil
}
