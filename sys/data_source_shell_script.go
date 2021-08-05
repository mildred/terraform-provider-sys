package sys

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceShellScript() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceShellScriptRead,

		Schema: map[string]*schema.Schema{
			"working_directory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "",
			},
			"shell": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "/bin/sh",
			},
			"read": {
				Type:        schema.TypeString,
				Description: "Shell script to read the value",
				Required:    true,
				ForceNew:    true,
			},
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_base64": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceShellScriptRead(d *schema.ResourceData, _ interface{}) error {
	script := d.Get("read")
	cwd := d.Get("working_directory")
	shell := d.Get("shell").(string)
	content, err := resourceShellScriptRun(cwd, shell, script.(string))
	if err != nil {
		return err
	}

	d.Set("content", string(content))
	d.Set("content_base64", base64.StdEncoding.EncodeToString([]byte(content)))

	checksum := sha1.Sum([]byte(content))
	d.SetId(hex.EncodeToString(checksum[:]))

	return nil
}
