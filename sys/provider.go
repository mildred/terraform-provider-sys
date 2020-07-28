package sys

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"sys_file":         resourceFile(),
			"sys_dir":          resourceDir(),
			"sys_shell_script": resourceShellScript(),
			"sys_symlink":      resourceSymlink(),
			"sys_null":         resourceNull(),
			"sys_package":      resourcePackage(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"sys_file": dataSourceFile(),
		},
	}
}
