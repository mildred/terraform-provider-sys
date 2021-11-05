package sys

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOsRelease() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOsReleaseRead,

		Schema: map[string]*schema.Schema{
			"filename": {
				Type:        schema.TypeString,
				Description: "Path to the os-release file",
				Optional:    true,
				ForceNew:    true,
				Default:     "/etc/os-release",
			},
			"result": {
				Type:     schema.TypeMap,
				Description: "Map of the variables contained in the file",
				Computed: true,
			},
			"raw_content": {
				Type:     schema.TypeString,
				Description: "Raw content of the file",
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A string identifying the operating system, without a version component, and suitable for presentation to the user. If not set, a default of "NAME=Linux" may be used.`,
			},
			"os_id": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A lower-case string (no spaces or other characters outside of 0–9, a–z, ".", "_" and "-") identifying the operating system, excluding any version information and suitable for processing by scripts or usage in generated filenames. If not set, a default of "ID=linux" may be used.`,
			},
			"id_like": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A space-separated list of operating system identifiers in the same syntax as the ID= setting. It should list identifiers of operating systems that are closely related to the local operating system in regards to packaging and programming interfaces, for example listing one or more OS identifiers the local OS is a derivative from. An OS should generally only list other OS identifiers it itself is a derivative of, and not any OSes that are derived from it, though symmetric relationships are possible. Build scripts and similar should check this variable if they need to identify the local operating system and the value of ID= is not recognized. Operating systems should be listed in order of how closely the local operating system relates to the listed ones, starting with the closest. This field is optional.`,
			},
			"pretty_name": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A pretty operating system name in a format suitable for presentation to the user. May or may not contain a release code name or OS version of some kind, as suitable. If not set, a default of "PRETTY_NAME="Linux"" may be used`,
			},
			"cpe_name": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A CPE name for the operating system, in URI binding syntax, following the Common Platform Enumeration Specification as proposed by the NIST. This field is optional.`,
			},
			"variant": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A string identifying a specific variant or edition of the operating system suitable for presentation to the user. This field may be used to inform the user that the configuration of this system is subject to a specific divergent set of rules or default configuration settings. This field is optional and may not be implemented on all systems.`,
			},
			"variant_id": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A lower-case string (no spaces or other characters outside of 0–9, a–z, ".", "_" and "-"), identifying a specific variant or edition of the operating system. This may be interpreted by other packages in order to determine a divergent default configuration. This field is optional and may not be implemented on all systems.`,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A string identifying the operating system version, excluding any OS name information, possibly including a release code name, and suitable for presentation to the user. This field is optional.`,
			},
			"version_id": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A lower-case string (mostly numeric, no spaces or other characters outside of 0–9, a–z, ".", "_" and "-") identifying the operating system version, excluding any OS name information or release code name, and suitable for processing by scripts or usage in generated filenames. This field is optional.`,
			},
			"version_codename": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A lower-case string (no spaces or other characters outside of 0–9, a–z, ".", "_" and "-") identifying the operating system release code name, excluding any OS name information or release version, and suitable for processing by scripts or usage in generated filenames. This field is optional and may not be implemented on all systems.`,
			},
			"build_id": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A string uniquely identifying the system image originally used as the installation base. In most cases, VERSION_ID or IMAGE_ID+IMAGE_VERSION are updated when the entire system image is replaced during an update. BUILD_ID may be used in distributions where the original installation image version is important: VERSION_ID would change during incremental system updates, but BUILD_ID would not. This field is optional.`,
			},
			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A lower-case string (no spaces or other characters outside of 0–9, a–z, ".", "_" and "-"), identifying a specific image of the operating system. This is supposed to be used for environments where OS images are prepared, built, shipped and updated as comprehensive, consistent OS images. This field is optional and may not be implemented on all systems, in particularly not on those that are not managed via images but put together and updated from individual packages and on the local system.`,
			},
			"image_version": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `A lower-case string (mostly numeric, no spaces or other characters outside of 0–9, a–z, ".", "_" and "-") identifying the OS image version. This is supposed to be used together with IMAGE_ID described above, to discern different versions of the same image.`,
			},
		},
	}
}

func dataSourceOsReleaseRead(d *schema.ResourceData, _ interface{}) error {
	path := d.Get("filename").(string)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	d.Set("raw_content", string(content))

	result := map[string]string{}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		key := ""
		val := ""
		in_key := true
		in_esc := false
		quote := ""
		for _, c := range line {
			char := string(c)
			if in_esc {
				in_esc = false
			} else if in_key && c == '=' {
				in_key = false
				char = ""
			} else {
				switch char {
				case quote:
					quote = ""
					char = ""
				case "\\":
					if quote != "'" {
						in_esc = true
						char = ""
					}
				case "'", "\"":
					quote = string(c)
					char = ""
				}
			}
			if in_key {
				key = key + char
			} else {
				val = val + char
			}
		}

		result[key] = val
	}

	d.Set("result", result)

	d.Set("name", result["NAME"])
	d.Set("os_id", result["ID"])
	d.Set("id_like", result["ID_LIKE"])
	d.Set("pretty_name", result["PRETTY_NAME"])
	d.Set("cpe_name", result["CPE_NAME"])
	d.Set("variant", result["VARIANT"])
	d.Set("variant_id", result["VARIANT_ID"])
	d.Set("version", result["VERSION"])
	d.Set("version_id", result["VERSION_ID"])
	d.Set("version_codename", result["VERSION_CODENAME"])
	d.Set("build_id", result["BUILD_ID"])
	d.Set("image_id", result["IMAGE_ID"])
	d.Set("image_version", result["IMAGE_VERSION"])

	checksum := sha1.Sum([]byte(content))
	d.SetId(hex.EncodeToString(checksum[:]))

	return nil
}
