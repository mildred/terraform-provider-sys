package sys

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
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
				Required: true,
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
				Required: true,
				ForceNew: true,
			},
			"read": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"delete": {
				Type:     schema.TypeString,
				Required: true,
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
	script := d.Get("read").(string)
	id, err := resourceShellScriptRun(d, script)
	d.SetId(id)
	return err
}

func resourceShellScriptCreate(d *schema.ResourceData, _ interface{}) error {
	script := d.Get("create").(string)
	id, err := resourceShellScriptRun(d, script)
	d.SetId(id)
	return err
}

func resourceShellScriptDelete(d *schema.ResourceData, _ interface{}) error {
	script := d.Get("delete").(string)
	_, err := resourceShellScriptRun(d, script)
	return err
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
