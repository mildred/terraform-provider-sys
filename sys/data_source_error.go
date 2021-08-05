package sys

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceError() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceErrorRead,

		Schema: map[string]*schema.Schema{
			"error": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "An error occurred",
			},
		},
	}
}

func dataSourceErrorRead(d *schema.ResourceData, _ interface{}) error {
	is_error := d.Get("error").(bool)
	message := d.Get("message").(string)

	if is_error {
		return errors.New(message)
	} else {
		d.SetId("")
		return nil
	}
}
