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
				Description: "Working directory where to run the script",
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"shell": {
				Description: "Shell to use",
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "/bin/sh",
			},
			"make": {
				Description: "Script to construct the resource (does not read the value)",
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"create": {
				Description: "Script to construct the resource, in addition to `make`. Must output on the standard output the resource id (used to determine if the resource needs to be reconstructed).",
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"read": {
				Description: "Script that reads the resource id on the standard output",
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"delete": {
				Description: "Script to delete the resource",
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"filename": {
				Description: "Filename created by the resource, can be used to avoid implementing `read`. The file is removed on resource deletion.",
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
	cwd := d.Get("working_directory")
	shell := d.Get("shell").(string)
	if ok {
		id, err := resourceShellScriptRun(cwd, shell, script.(string), true)
		d.SetId(id)
		if err != nil {
			return fmt.Errorf("cannot execute read script, %v", err)
		}
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
	script_make, okm := d.GetOk("make")
	script, okc := d.GetOk("create")
	filename, okf := d.GetOk("filename")
	script_read, okr := d.GetOk("read")
	cwd := d.Get("working_directory")
	shell := d.Get("shell").(string)

	scriptname := "create"
	if script.(string) == "" {
		script = script_read
		scriptname = "read"
	}

	if okm {
		_, err := resourceShellScriptRun(cwd, shell, script_make.(string), false)
		if err != nil {
			return fmt.Errorf("cannot execute make script, %v", err)
		}
	}

	if okc && okr {
		id, err := resourceShellScriptRun(cwd, shell, script.(string), true)
		d.SetId(id)
		if err != nil {
			return fmt.Errorf("cannot execute %s script, %v", scriptname, err)
		}
	} else if okc && okf {
		_, err := resourceShellScriptRun(cwd, shell, script.(string), false)
		if err != nil {
			return fmt.Errorf("cannot execute %s script, %v", scriptname, err)
		}
		id, err := checksumFile(filename.(string))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot checksum file, %v", err)
		}
		d.SetId(id)
	} else if okc {
		_, err := resourceShellScriptRun(cwd, shell, script.(string), false)
		if err != nil {
			return fmt.Errorf("cannot execute %s script, %v", scriptname, err)
		}
		d.SetId("1")
	} else {
		d.SetId("1")
	}
	return nil
}

func resourceShellScriptDelete(d *schema.ResourceData, _ interface{}) error {
	script, ok := d.GetOk("delete")
	cwd := d.Get("working_directory")
	shell := d.Get("shell").(string)
	if ok {
		_, err := resourceShellScriptRun(cwd, shell, script.(string), false)
		return err
	} else if filename, ok := d.GetOk("filename"); ok {
		return os.RemoveAll(filename.(string))
	}
	return nil
}

func resourceShellScriptRun(cwd interface{}, shell, script string, collectId bool) (string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := exec.Command(shell)
	cmd.Stdin = bytes.NewReader([]byte(script))
	if collectId {
	    cmd.Stdout = stdout
	} else {
	    cmd.Stdout = stderr
	}
	cmd.Stderr = stderr
	if cwd != nil {
		cmd.Dir = cwd.(string)
	}

	err := cmd.Run()

	if err != nil {
		if er, ok := err.(*exec.ExitError); ok && er != nil {
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
