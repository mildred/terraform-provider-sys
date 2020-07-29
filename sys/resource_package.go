package sys

import (
	"bytes"
	"os"
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"context"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

var debLock sync.Mutex

func resourcePackage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePackageCreate,
		ReadContext:   resourcePackageRead,
		DeleteContext: resourcePackageDelete,

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
			"installed_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func debRead(d *schema.ResourceData, m *providerConfiguration, lock bool) diag.Diagnostics {
	if lock {
		debLock.Lock()
		defer debLock.Unlock()
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := exec.Command("dpkg", "-s", d.Get("name").(string))
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		if er := err.(*exec.ExitError); er != nil {
			d.SetId("")
		} else {
			return diag.Errorf("cannot query dpkg -s: %v\n%s", err, stderr.String())
		}
	} else {
		var name, version string

		r := bufio.NewScanner(stdout)
		for r.Scan() {
			line := r.Text()
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				continue
			}
			value := strings.TrimSpace(parts[1])

			switch key := parts[0]; key {
			case "Package":
				d.Set("name", value)
				name = value
			case "Version":
				d.Set("installed_version", value)
				version = value
			}
		}
		if err := r.Err(); err != nil {
			return diag.Errorf("cannot parse dpkg -s results: %v\n%s", err, stdout.String())
		} else if name == "" || version == "" {
			return diag.Errorf("cannot parse dpkg -s results (name: %s, version: %s):\n%s", name, version, stdout.String())
		}
		d.SetId(fmt.Sprintf("%s_%s", name, version))
	}
	return nil
}

func resourcePackageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	switch t := d.Get("type").(string); t {
	case "deb":
		return debRead(d, m.(*providerConfiguration), true)
	default:
		return diag.Errorf("Unknown package type %s", t)
	}
}

func debCreate(d *schema.ResourceData, m *providerConfiguration) diag.Diagnostics {
	debLock.Lock()
	defer debLock.Unlock()

	if ! m.debUpdated {
		stderr := new(bytes.Buffer)
		cmd := exec.Command("apt-get", "update")
		cmd.Stderr = stderr
		//cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return diag.Errorf("Error running apt-get update: %s\n%s", err.Error(), stderr.String())
		}
		m.debUpdated = true
	}

	pkgspec := d.Get("name").(string)
	version := d.Get("version")
	if version != nil && version.(string) != "" {
		pkgspec = fmt.Sprintf("%s=%s", pkgspec, version)
	}

	stderr := new(bytes.Buffer)
	cmd := exec.Command("apt-get", "install", "-y", pkgspec)
	cmd.Stderr = stderr
	//cmd.Stdout = os.Stdout
	cmd.Env = append(os.Environ(),
		"DEBIAN_FRONTEND=noninteractive",
	)
	err := cmd.Run()
	if err != nil {
		return diag.Errorf("Error running apt-get install %s: %s\n%s", pkgspec, err.Error(), stderr.String())
	}

	return debRead(d, m, false)
}

func resourcePackageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	switch t := d.Get("type").(string); t {
	case "deb":
		return debCreate(d, m.(*providerConfiguration))
	default:
		return diag.Errorf("Unknown package type %s", t)
	}
}

func debDelete(d *schema.ResourceData, m *providerConfiguration) diag.Diagnostics {
	debLock.Lock()
	defer debLock.Unlock()

	stderr := new(bytes.Buffer)
	pkgspec := d.Get("name").(string)

	cmd := exec.Command("apt-get", "remove", "-y")
	cmd.Stderr = stderr
	//cmd.Stdout = os.Stdout
	cmd.Env = append(os.Environ(),
		"DEBIAN_FRONTEND=noninteractive",
	)
	err := cmd.Run()
	if err != nil {
		return diag.Errorf("Error running apt-get remove %s: %s\n%s", pkgspec, err.Error(), stderr.String())
	}

	return debRead(d, m, false)
}

func resourcePackageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	switch t := d.Get("type").(string); t{
	case "deb":
		return debDelete(d, m.(*providerConfiguration))
	default:
		return diag.Errorf("Unknown package type %s", t)
	}
}

