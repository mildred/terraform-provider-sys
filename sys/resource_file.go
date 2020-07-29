package sys

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/mildred/terraform-provider-sys/sys/utils"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFile() *schema.Resource {
	return &schema.Resource{
		Create: resourceFileCreate,
		Read:   resourceFileRead,
		Delete: resourceFileDelete,
		Update: resourceFileUpdate,

		Schema: map[string]*schema.Schema{
			"content": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"sensitive_content", "content_base64", "source"},
			},
			"sensitive_content": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"content", "content_base64", "source"},
			},
			"content_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"sensitive_content", "content", "source"},
			},
			"source": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"content", "sensitive_content", "content_base64"},
			},
			"filename": {
				Type:        schema.TypeString,
				Description: "Path to the output file",
				Required:    true,
				ForceNew:    true,
			},
			"file_permission": {
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
			"force_overwrite": {
				Type:        schema.TypeBool,
				Description: "Force overwrite an existing file",
				Optional:    true,
				Default:     false,
			},
			"systemd": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unit": {
							Type:        schema.TypeString,
							Description: "Name of the unit",
							Required:    true,
						},
						"enable": {
							Type:        schema.TypeBool,
							Description: "Enable the unit",
							Optional:    true,
						},
						"start": {
							Type:        schema.TypeBool,
							Description: "Start the unit",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

type resourceFileSystemd struct {
	unit       string
	enable     bool
	has_enable bool
	start      bool
	has_start  bool
	old_enable bool
	had_enable bool
	old_start  bool
	had_start  bool
}

func resourceFileReadSystemd(d *schema.ResourceData) map[string]resourceFileSystemd {
	oldVal, newVal := d.GetChange("systemd")
	res := map[string]resourceFileSystemd{}

	if oldVal != nil {
		for _, raw := range oldVal.([]interface{}) {
			sd := raw.(map[string]interface{})
			var r resourceFileSystemd
			r.unit = sd["unit"].(string)
			r.had_enable = sd["enable"] != nil
			if r.had_enable {
				r.old_enable = sd["enable"].(bool)
			}
			r.had_start = sd["start"] != nil
			if r.had_start {
				r.old_start = sd["start"].(bool)
			}
			res[r.unit] = r
		}
	}

	if newVal != nil {
		for _, raw := range newVal.([]interface{}) {
			sd := raw.(map[string]interface{})
			unit := sd["unit"].(string)
			r := res[unit]
			r.has_enable = sd["enable"] != nil
			if r.has_enable {
				r.enable = sd["enable"].(bool)
			}
			r.has_start = sd["start"] != nil
			if r.has_start {
				r.start = sd["start"].(bool)
			}
			res[r.unit] = r
		}
	}

	return res
}

func resourceFileRead(d *schema.ResourceData, _ interface{}) error {
	// If the output file doesn't exist, mark the resource for creation.
	outputPath := d.Get("filename").(string)
	st, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		d.SetId("")
		return nil
	}

	same, err := utils.FileModeSame(d.Get("file_permission").(string), st.Mode(), utils.Umask)
	if err != nil {
		return err
	}
	if ! same {
		d.Set("file_permission", st.Mode().String())
	}

	// Verify that the content of the destination file matches the content we
	// expect. Otherwise, the file might have been modified externally and we
	// must reconcile.
	outputContent, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("Cannot read file, %v", err)
	}

	outputChecksum := sha1.Sum([]byte(outputContent))
	if hex.EncodeToString(outputChecksum[:]) != d.Id() {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceFileContent(d *schema.ResourceData) ([]byte, bool, error) {
	if content, sensitiveSpecified := d.GetOk("sensitive_content"); sensitiveSpecified {
		return []byte(content.(string)), true, nil
	}
	if b64Content, b64Specified := d.GetOk("content_base64"); b64Specified {
		res, err := base64.StdEncoding.DecodeString(b64Content.(string))
		return res, true, err
	}
	if content, contentSpecified := d.GetOk("content"); contentSpecified {
		return []byte(content.(string)), true, nil
	}
	return nil, false, nil
}

func resourceFileUpdate(d *schema.ResourceData, _ interface{}) error {
	destination := d.Get("path").(string)

	if d.HasChange("file_permission") {
		perm := d.Get("file_permission").(string)
		modeInt, _ := strconv.ParseInt(perm, 8, 64)
		mode := os.FileMode(modeInt)

		err := os.Chmod(destination, mode)
		if err != nil {
			return fmt.Errorf("cannot chmod %s, %s", mode, err)
		}
	}

	/*
	reload := false
	for unit, sd := range resourceFileReadSystemd(d) {
		if !reload {
			err := systemdDaemonReload()
			if err != nil {
				return err
			}
			reload = true
		}
		systemdUpdnStartEnable(unit, true, sd.enable, sd.start)
	}
	*/

	return nil
}

func resourceFileCreate(d *schema.ResourceData, _ interface{}) error {
	forceOverwrite := d.Get("force_overwrite").(bool)
	source, sourceSpecified := d.GetOk("source")
	content, contentSpecified, err := resourceFileContent(d)
	if err != nil {
		return fmt.Errorf("content error, %v", err)
	}

	destination := d.Get("filename").(string)

	destinationDir := path.Dir(destination)
	if _, err := os.Stat(destinationDir); err != nil {
		dirPerm := d.Get("directory_permission").(string)
		dirMode, _ := strconv.ParseInt(dirPerm, 8, 64)
		if err := os.MkdirAll(destinationDir, os.FileMode(dirMode)); err != nil {
			return fmt.Errorf("cannot create parent directories, %v", err)
		}
	}

	filePerm := d.Get("file_permission").(string)

	fileMode, _ := strconv.ParseInt(filePerm, 8, 64)

	if sourceSpecified {
		if !forceOverwrite {
			if _, err := os.Lstat(destination); err == nil || !os.IsNotExist(err) {
				return fmt.Errorf("destination file exists at %v", destination)
			}
		}
		err = getter.GetFile(destination, source.(string))
		if err != nil {
			return fmt.Errorf("cannot fetch source %v, %v", source, err)
		}
	}

	if contentSpecified {
		data := []byte(content)
		flags := os.O_WRONLY | os.O_CREATE
		if forceOverwrite {
			flags = flags | os.O_EXCL
		} else {
			flags = flags | os.O_TRUNC
		}
		f, err := os.OpenFile(destination, flags, os.FileMode(fileMode))
		if err != nil {
			return fmt.Errorf("cannot write file, %v", err)
		}
		n, err := f.Write(data)
		if err == nil && n < len(data) {
			err = io.ErrShortWrite
		}
		if err1 := f.Close(); err == nil {
			err = err1
		}
		if err != nil {
			return fmt.Errorf("cannot write file, %v", err)
		}

		checksum := sha1.Sum([]byte(content))
		d.SetId(hex.EncodeToString(checksum[:]))
	} else {
		err = os.Chmod(destination, os.FileMode(fileMode))
		if err != nil {
			return fmt.Errorf("cannot chmod %s, %v", filePerm, err)
		}
		h := sha1.New()
		f, err := os.Open(destination)
		if err != nil {
			return fmt.Errorf("cannot open file, %v", err)
		}
		defer f.Close()
		_, err = io.Copy(h, f)
		if err != nil {
			return fmt.Errorf("cannot checksum file, %v", err)
		}
		checksum := h.Sum(nil)
		d.SetId(hex.EncodeToString(checksum[:]))
	}

	/*
	reload := false
	for unit, sd := range resourceFileReadSystemd(d) {
		if !reload {
			err := systemdDaemonReload()
			if err != nil {
				return err
			}
			reload = true
		}
		systemdUpdnStartEnable(unit, true, sd.enable, sd.start)
	}
	*/

	return nil
}

func resourceFileDelete(d *schema.ResourceData, _ interface{}) error {
	err := os.Remove(d.Get("filename").(string))
	if err != nil {
		return fmt.Errorf("cannot delete file, %v", err)
	}

	/*
	reload := false
	for unit, sd := range resourceFileReadSystemd(d) {
		if !reload {
			err := systemdDaemonReload()
			if err != nil {
				return err
			}
			reload = true
		}
		systemdUpdnStartEnable(unit, false, sd.old_enable, sd.old_start)
	}
	*/

	return nil
}
