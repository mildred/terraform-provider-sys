package sys

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"regexp"
	"os/exec"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var uname_all_regexp *regexp.Regexp
func init(){
	uname_all_regexp = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(.+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)$`)
}

func dataSourceUname() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUnameRead,
		Description: "Return values from the uname executable",

		Schema: map[string]*schema.Schema{
			"flag": {
				Type:        schema.TypeString,
				Default:     "a",
				Optional:    true,
				ForceNew:    true,
				Description: `Uname flag without the dash: a: all, s: kernel name, n: nodename, r: kernel release, v: kernel version, m: machine, p: processor, i: hardware platform, o: operating system`,
			},
			"output": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `Output from uname command`,
			},
			"kernel_name": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `uname -s`,
			},
			"nodename": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `uname -n`,
			},
			"kernel_release": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `uname -r`,
			},
			"kernel_version": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `uname -v`,
			},
			"machine": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `uname -m`,
			},
			"processor": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `uname -p`,
			},
			"hardware_platform": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `uname -i`,
			},
			"operating_system": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `uname -o`,
			},
		},
	}
}

func dataSourceUnameRead(d *schema.ResourceData, _ interface{}) error {
	flag := d.Get("flag").(string)
	if len(flag) == 0 {
		flag = "-a"
	} else if flag[0] != '-' {
		if len(flag) > 1 {
			flag = "--" + flag
		} else {
			flag = "-" + flag
		}
	}

	out, err := exec.Command("uname", flag).Output()
	if err != nil {
		return err
	}

	out_line := strings.Trim(string(out), "\n\r\t ")

	d.Set("output", out_line)

	switch flag {
	case "-a", "--all":
		parts := uname_all_regexp.FindStringSubmatch(out_line)
		d.Set("kernel_name",       parts[1])
		d.Set("nodename",          parts[2])
		d.Set("kernel_release",    parts[3])
		d.Set("kernel_version",    parts[4])
		d.Set("machine",           parts[5])
		d.Set("processor",         parts[6])
		d.Set("hardware_platform", parts[7])
		d.Set("operating_system",  parts[8])
	case "-s", "--kernel-name":
		d.Set("kernel_name",       out_line)
	case "-n", "--nodename":
		d.Set("nodename",          out_line)
	case "-r", "--kernel-release":
		d.Set("kernel_release",    out_line)
	case "-v", "--kernel-version":
		d.Set("kernel_version",    out_line)
	case "-m", "--machine":
		d.Set("machine",           out_line)
	case "-p", "--processor":
		d.Set("processor",         out_line)
	case "-i", "--hardware-platform":
		d.Set("hardware_platform", out_line)
	case "-o", "--operating-system":
		d.Set("operating_system",  out_line)
	}

	checksum := sha1.Sum([]byte(flag + "\n" + out_line))
	d.SetId(hex.EncodeToString(checksum[:]))

	return nil
}
