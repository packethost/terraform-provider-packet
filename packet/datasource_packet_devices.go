package packet

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/packethost/packngo"
)

func dataSourcePacketDevices() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePacketDevicesRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"spot_market_request_id"},
			},
			"spot_market_request_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"project_id"},
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"facilities": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"plans": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"operating_systems": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"public_ipv4s": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"private_ipv4s": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"public_ipv6s": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func getComputedListsFromDeviceList(devices []packngo.Device) (ids, pub4, pri4, pub6 []string) {
	for _, d := range devices {
		ids = append(ids, d.ID)
		pu4, pr4, pu6 := getPub4Pri4Pub6(d.Network)
		pub4 = append(pub4, pu4)
		pri4 = append(pri4, pr4)
		pub6 = append(pub6, pu6)
	}
	return ids, pub4, pri4, pub6
}

func getPub4Pri4Pub6(ns []*packngo.IPAddressAssignment) (pub4, pri4, pub6 string) {
	for _, ip := range ns {
		if ip.Management {
			if ip.AddressFamily == 4 {
				if ip.Public {
					pub4 = ip.Address
				} else {
					pri4 = ip.Address
				}
			} else {
				pub6 = ip.Address
			}
		}
	}
	return pub4, pri4, pub6

}

type deviceFilter = func([]string, packngo.Device) bool

func tagFilter(f []string, d packngo.Device) bool {
	return stringSlicesIntersect(f, d.Tags)
}

func facilityFilter(f []string, d packngo.Device) bool {
	return contains(f, d.Facility.Code)
}

func planFilter(f []string, d packngo.Device) bool {
	return contains(f, d.Plan.Slug) || contains(f, d.Plan.Name)
}

func osFilter(f []string, d packngo.Device) bool {
	return contains(f, d.OS.Slug)
}

func nameFilter(f []string, d packngo.Device) bool {
	return contains(f, d.Hostname)
}

func dataSourcePacketDevicesRead(d *schema.ResourceData, meta interface{}) error {
	var ids, pub4, pri4, pub6 []string
	var ds []packngo.Device
	var err error
	client := meta.(*packngo.Client)
	smrIdRaw, spotOK := d.GetOk("spot_market_request_id")
	pIdRaw, projOK := d.GetOk("project_id")
	if !projOK && !spotOK {
		return fmt.Errorf("You have to set either project_id or spot_market_request_id")
	}

	if spotOK {
		opts := packngo.GetOptions{Includes: []string{"devices"}}
		smr, _, err := client.SpotMarketRequests.Get(smrIdRaw.(string), &opts)
		if err != nil {
			return friendlyError(err)
		}
		ds = smr.Devices
	} else {
		pid := pIdRaw.(string)
		ds, _, err = client.Devices.List(pid, nil)
		if err != nil {
			return friendlyError(err)
		}
	}
	log.Printf("+++++++++++++++++++++++++++++++++++ ds len %d", len(ds))
	// tags, facilities, plans, operating_systems
	fields := []string{"tags", "names", "facilities", "plans", "operating_systems"}
	filters := []deviceFilter{tagFilter, nameFilter, facilityFilter, planFilter, osFilter}

	for i, field := range fields {
		n := d.Get(field + ".#").(int)
		if n > 0 {
			newDs := []packngo.Device{}
			values := convertStringArr(d.Get(field).([]interface{}))
			for _, de := range ds {
				if filters[i](values, de) {
					newDs = append(newDs, de)
				}
			}
			ds = newDs
		}
	}
	log.Printf("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@ ds len %d", len(ds))

	ids, pub4, pri4, pub6 = getComputedListsFromDeviceList(ds)
	d.Set("ids", ids)
	d.Set("public_ipv4s", pub4)
	d.Set("private_ipv4s", pri4)
	d.Set("public_ipv6s", pub6)

	return nil

}
