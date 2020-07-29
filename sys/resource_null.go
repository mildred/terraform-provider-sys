package sys

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func resourceNull() *schema.Resource {
	return &schema.Resource{
		Create: resourceNullCreate,
		Read:   resourceNullRead,
		Delete: resourceNullDelete,
		Update: resourceNullUpdate,

		Schema: map[string]*schema.Schema{
			"triggers": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"values": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},
			"inputs": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},
			"outputs": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceNullCreate(d *schema.ResourceData, meta interface{}) error {
	id := fmt.Sprintf("%d", rand.Int())
	d.SetId(id)
	d.Set("outputs", d.Get("inputs"))
	d.Set("triggers", d.Get("triggers"))
	return nil
}

func resourceNullRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceNullUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceNullDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}
