package sys

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceShellScript() *schema.Resource {
	return &schema.Resource{
		Create: resourceShellScriptCreate,
		Read:   resourceShellScriptRead,
		Delete: resourceShellScriptDelete,

		Schema: map[string]*schema.Schema{
			"working_directory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"shell": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "/bin/sh",
			},
			"create": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"read": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"delete": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"filename": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

type ExitError struct {
	exec.ExitError
}

func (err *ExitError) Error() string {
	return fmt.Sprintf("%v: %v", err.ExitError.Error(), string(err.Stderr))
}

func resourceShellScriptRead(d *schema.ResourceData, _ interface{}) error {
	script, ok := d.GetOk("read")
	if ok {
		id, err := resourceShellScriptRun(d, script.(string))
		d.SetId(id)
		return err
	} else if filename, ok := d.GetOk("filename"); ok {
		id, err := checksumFile(filename.(string))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot checksum file, %v", err)
		}
		d.SetId(id)
	}
	return nil
}

func resourceShellScriptCreate(d *schema.ResourceData, _ interface{}) error {
	script, ok := d.GetOk("create")
	filename, okf := d.GetOk("filename")
	_, okr := d.GetOk("read")
	if ok && okr {
		id, err := resourceShellScriptRun(d, script.(string))
		d.SetId(id)
		return err
	} else if ok && okf {
		_, err := resourceShellScriptRun(d, script.(string))
		if err != nil {
			return err
		}
		id, err := checksumFile(filename.(string))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot checksum file, %v", err)
		}
		d.SetId(id)
	} else if ok {
		_, err := resourceShellScriptRun(d, script.(string))
		d.SetId("1")
		return err
	} else {
		d.SetId("1")
	}
	return nil
}

func resourceShellScriptDelete(d *schema.ResourceData, _ interface{}) error {
	script, ok := d.GetOk("delete")
	if ok {
		_, err := resourceShellScriptRun(d, script.(string))
		return err
	} else if filename, ok := d.GetOk("filename"); ok {
		return os.RemoveAll(filename.(string))
	}
	return nil
}

func resourceShellScriptRun(d *schema.ResourceData, script string) (string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cwd := d.Get("working_directory").(string)
	cmd := exec.Command(d.Get("shell").(string))
	cmd.Stdin = bytes.NewReader([]byte(script))
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Dir = cwd

	err := cmd.Run()

	if err != nil {
		if er := err.(*exec.ExitError); er != nil {
			er.Stderr = stderr.Bytes()
			err = &ExitError{*er}
		}
	}

	id := strings.TrimRight(stdout.String(), "\n")

	if len(id) > 64 {
		checksum := sha1.Sum([]byte(id))
		id = hex.EncodeToString(checksum[:])
	}

	return id, err
}
