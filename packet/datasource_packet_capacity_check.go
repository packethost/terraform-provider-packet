package packet

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/packethost/packngo"
)

func dataSourceCapacityCheck() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePacketCapacityCheckRead,
		Schema: map[string]*schema.Schema{
			"facility": {
				Type:     schema.TypeString,
				Required: true,
			},
			"plan": {
				Type:     schema.TypeString,
				Required: true,
			},
			"quantity": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"available": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"keepers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func dataSourcePacketCapacityCheckRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*packngo.Client)
	si := packngo.ServerInfo{
		Facility: d.Get("facility").(string),
		Plan:     d.Get("plan").(string),
		Quantity: d.Get("quantity").(int),
	}

	ci := &packngo.CapacityInput{[]packngo.ServerInfo{si}}

	cap, _, err := client.CapacityService.Check(ci)
	if err != nil {
		return friendlyError(err)
	}
	cis := cap.Servers
	if len(cis) != 1 {
		return friendlyError(fmt.Errorf("only one CapacityInput should have been returned, was %+v", cis))
	}
	d.Set("available", cis[0].Available)
	d.SetId("none")
	return nil
}
