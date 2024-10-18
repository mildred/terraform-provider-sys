package sys

import (
	"os"
	"path"
	"strconv"
	"fmt"

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
			"directory_permission": {
				Description:  "(default: \"0777\") The permission to set for any directories created. Expects a string.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "0777",
				ValidateFunc: validateMode,
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

	destinationDir := path.Dir(destination)
	if _, err := os.Stat(destinationDir); err != nil {
		dirPerm := d.Get("directory_permission").(string)
		dirMode, _ := strconv.ParseInt(dirPerm, 8, 64)

		if err := os.MkdirAll(destinationDir, os.FileMode(dirMode)); err != nil {
			return fmt.Errorf("cannot create parent directories, %v", err)
		}
	}

	err := os.Symlink(source, destination)
	if err != nil {
		return err
	}

	return resourceSymlinkRead(d, nil)
}

func resourceSymlinkDelete(d *schema.ResourceData, _ interface{}) error {
	return os.Remove(d.Get("path").(string))
}
