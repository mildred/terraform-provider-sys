package sys

import (
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSymlink() *schema.Resource {
	return &schema.Resource{
		Create: resourceSymlinkCreate,
		Read:   resourceSymlinkRead,
		Delete: resourceSymlinkDelete,
		Exists: resourceSymlinkExists,

		Description: "Creates a symlink",

		Schema: map[string]*schema.Schema{
			"source": {
				Type:        schema.TypeString,
				Description: "Symlink source path",
				Optional:    true,
				ForceNew:    true,
			},
			"path": {
				Type:        schema.TypeString,
				Description: "Path to the output file",
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceSymlinkRead(d *schema.ResourceData, _ interface{}) error {
	path := d.Get("path").(string)

	target, err := os.Readlink(path)
	if err == nil {
		d.Set("source", target)
		d.SetId(path)
	} else {
		d.SetId("")
	}

	return nil
}

func resourceSymlinkExists(d *schema.ResourceData, _ interface{}) (bool, error) {
	path := d.Get("path").(string)

	target, err := os.Readlink(path)
	if err == nil {
		d.Set("source", target)
		return true, nil
	} else {
		return false, nil
	}
}

func resourceSymlinkCreate(d *schema.ResourceData, _ interface{}) error {
	source := d.Get("source").(string)
	destination := d.Get("path").(string)

	err := os.Symlink(source, destination)
	if err != nil {
		return err
	}

	return resourceSymlinkRead(d, nil)
}

func resourceSymlinkDelete(d *schema.ResourceData, _ interface{}) error {
	return os.Remove(d.Get("path").(string))
}
