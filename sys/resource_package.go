package sys

import (
	"bytes"
	"io"
	"os"
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourcePackage() *schema.Resource {
	return &schema.Resource{
		Create: resourcePackageCreate,
		Read:   resourcePackageRead,
		Delete: resourcePackageDelete,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func debRead(d *schema.ResourceData, _ interface{}) error {
	stdout := new(bytes.Buffer)

	cmd := exec.Command("dpkg", "-s", d.Get("name").(string))
	cmd.Stdout = stdout

	err := cmd.Run()
	if err != nil {
		if er := err.(*exec.ExitError); er != nil {
			d.SetId("")
		} else {
			return err
		}
	} else {
		var name, version string

		r := bufio.NewReader(stdout)
		line, err := r.ReadString(':')
		for err != io.EOF {
			key := line

			line, err = r.ReadString('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			value := strings.TrimSpace(line)
			switch key {
			case "Package":
				d.Set("name", value)
				name = value
			case "Version":
				d.Set("version", value)
				version = value
			}

			line, err = r.ReadString(':')
		}
		if err != nil && err != io.EOF {
			return err
		}
		d.SetId(fmt.Sprintf("%s_%s", name, version))
	}
	return err
}

func resourcePackageRead(d *schema.ResourceData, m interface{}) error {
	switch t := d.Get("type").(string); t {
	case "deb":
		return debRead(d, m)
	default:
		return fmt.Errorf("Unknown package type %s", t)
	}
}

func debCreate(d *schema.ResourceData, m interface{}) error {
	stderr := new(bytes.Buffer)
	cmd := exec.Command("apt-get", "update")
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error running apt-get update: %e\n%s", err, stderr.String())
	}

	pkgspec := d.Get("name").(string)
	version := d.Get("version")
	if version != nil {
		pkgspec = fmt.Sprintf("%s=%s", pkgspec, version)
	}

	stderr = new(bytes.Buffer)
	cmd = exec.Command("apt-get", "install", "-y")
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(),
		"DEBIAN_FRONTEND=noninteractive",
	)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error running apt-get install %s: %e\n%s", pkgspec, err, stderr.String())
	}

	return debRead(d, m)
}

func resourcePackageCreate(d *schema.ResourceData, m interface{}) error {
	switch t := d.Get("type").(string); t {
	case "deb":
		return debCreate(d, m)
	default:
		return fmt.Errorf("Unknown package type %s", t)
	}
}

func debDelete(d *schema.ResourceData, m interface{}) error {
	stderr := new(bytes.Buffer)
	pkgspec := d.Get("name").(string)

	cmd := exec.Command("apt-get", "remove", "-y")
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(),
		"DEBIAN_FRONTEND=noninteractive",
	)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error running apt-get remove %s: %e\n%s", pkgspec, err, stderr.String())
	}

	return debRead(d, m)
}

func resourcePackageDelete(d *schema.ResourceData, m interface{}) error {
	switch t := d.Get("type").(string); t{
	case "deb":
		return debDelete(d, m)
	default:
		return fmt.Errorf("Unknown package type %s", t)
	}
}

