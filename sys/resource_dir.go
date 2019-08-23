package sys

import (
	"os"
	"path"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDir() *schema.Resource {
	return &schema.Resource{
		Create: resourceDirCreate,
		Read:   resourceDirRead,
		Delete: resourceDirDelete,

		Schema: map[string]*schema.Schema{
			"path": {
				Type:        schema.TypeString,
				Description: "Path to the output file",
				Required:    true,
				ForceNew:    true,
			},
			"permission": {
				Type:         schema.TypeString,
				Description:  "Permissions to set for the output file",
				Optional:     true,
				ForceNew:     true,
				Default:      "0777",
				ValidateFunc: validateMode,
			},
			"directory_permission": {
				Type:         schema.TypeString,
				Description:  "Permissions to set for directories created",
				Optional:     true,
				ForceNew:     true,
				Default:      "0777",
				ValidateFunc: validateMode,
			},
		},
	}
}

func resourceDirRead(d *schema.ResourceData, _ interface{}) error {
	// If the output file doesn't exist, mark the resource for creation.
	outputPath := d.Get("path").(string)
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		d.SetId("")
		return nil
	}

	d.SetId(outputPath)
	return nil
}

func resourceDirCreate(d *schema.ResourceData, _ interface{}) error {
	destination := d.Get("path").(string)

	destinationDir := path.Dir(destination)
	if _, err := os.Stat(destinationDir); err != nil {
		dirPerm := d.Get("directory_permission").(string)
		dirMode, _ := strconv.ParseInt(dirPerm, 8, 64)
		if err := os.MkdirAll(destinationDir, os.FileMode(dirMode)); err != nil {
			return err
		}
	}

	dirPerm := d.Get("permission").(string)

	dirMode, _ := strconv.ParseInt(dirPerm, 8, 64)

	os.Mkdir(destination, os.FileMode(dirMode))

	d.SetId(destination)

	return nil
}

func resourceDirDelete(d *schema.ResourceData, _ interface{}) error {
	return os.Remove(d.Get("path").(string))
}
