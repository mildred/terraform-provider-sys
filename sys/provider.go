package sys

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"sys_file":         resourceFile(),
			"sys_dir":          resourceDir(),
			"sys_shell_script": resourceShellScript(),
			"sys_symlink":      resourceSymlink(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"sys_file": dataSourceFile(),
		},
	}
}
