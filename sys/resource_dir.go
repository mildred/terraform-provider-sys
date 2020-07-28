package sys

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/mildred/terraform-provider-sys/sys/utils"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDir() *schema.Resource {
	return &schema.Resource{
		Create: resourceDirCreate,
		Read:   resourceDirRead,
		Delete: resourceDirDelete,
		Update: resourceDirUpdate,

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
				Default:      "0777",
				ValidateFunc: validateMode,
			},
			"parent_permission": {
				Type:         schema.TypeString,
				Description:  "Permissions to set for directories created",
				Optional:     true,
				ForceNew:     true,
				Default:      "0777",
				ValidateFunc: validateMode,
			},
			"allow_existing": {
				Type:          schema.TypeBool,
				Description:   "Allow directory to exist prior to running terraform",
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"force_remove"},
			},
			"force_remove": {
				Type:          schema.TypeBool,
				Description:   "Force removing the directory even if not empty",
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"allow_existing"},
			},
		},
	}
}

func resourceDirRead(d *schema.ResourceData, _ interface{}) error {
	// If the output file doesn't exist, mark the resource for creation.
	outputPath := d.Get("path").(string)
	st, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		d.SetId("")
		return nil
	}

	same, err := utils.FileModeSame(d.Get("permission").(string), st.Mode(), utils.Umask)
	if err != nil {
		return err
	}
	if ! same {
		d.Set("permission", st.Mode().String())
	}

	d.SetId(outputPath)
	return nil
}

func resourceDirUpdate(d *schema.ResourceData, _ interface{}) error {
	destination := d.Get("path").(string)

	if d.HasChange("permission") {
		perm := d.Get("permission").(string)
		modeInt, _ := strconv.ParseInt(perm, 8, 64)
		mode := os.FileMode(modeInt)

		err := os.Chmod(destination, mode)
		if err != nil {
			return fmt.Errorf("cannot chmod %s, %s", mode, err)
		}
	}
	return nil
}

func resourceDirCreate(d *schema.ResourceData, _ interface{}) error {
	destination := d.Get("path").(string)
	allowExisting := d.Get("allow_existing").(bool)

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

	err := os.Mkdir(destination, os.FileMode(dirMode))
	if allowExisting && os.IsExist(err) {
		err = nil
	} else if err != nil {
		return err
	}

	d.SetId(destination)

	return nil
}

func resourceDirDelete(d *schema.ResourceData, _ interface{}) error {
	var err error
	forceRemove := d.Get("force_remove").(bool)
	allowExisting := d.Get("allow_existing").(bool)
	path := d.Get("path").(string)

	if forceRemove {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
		if allowExisting && os.IsExist(err) {
			err = nil
		}
	}

	return err
}
