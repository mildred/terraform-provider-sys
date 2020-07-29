package sys

import (
	"crypto/sha1"
	"io/ioutil"
	"encoding/base64"
	"encoding/hex"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFileRead,

		Schema: map[string]*schema.Schema{
			"filename": {
				Type:        schema.TypeString,
				Description: "Path to the output file",
				Required:    true,
				ForceNew:    true,
			},
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_base64": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFileRead(d *schema.ResourceData, _ interface{}) error {
	path := d.Get("filename").(string)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	d.Set("content", string(content))
	d.Set("content_base64", base64.StdEncoding.EncodeToString(content))

	checksum := sha1.Sum([]byte(content))
	d.SetId(hex.EncodeToString(checksum[:]))

	return nil
}
