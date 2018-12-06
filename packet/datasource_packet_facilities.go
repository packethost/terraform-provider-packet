package packet

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/packethost/packngo"
)

// As of end of 2018, faciltiy features are
// "baremetal", "layer_2", "backend_transfer", "storage", "global_ipv4"
func dataSourceFacility() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePacketFacilityRead,
		Schema: map[string]*schema.Schema{
			"slugs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"feature": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"plan": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"quantity": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"utilization"},
			},
			"utilization": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"quantity"},
			},
		},
	}
}

func filterOnPlan(slugs []string, cr *packngo.CapacityReport, plan string) []string {
	r := []string{}

	for f, planMap := range *cr {
		if _, ok := planMap[plan]; ok {
			r = append(r, f)
			continue
		}
	}
	return r
}

func findStr(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

func getQuantityCheckInput(slugs []string, plan string, q int) *packngo.CapacityInput {
	si := make([]packngo.ServerInfo, len(slugs))
	for i, s := range slugs {
		si[i] = packngo.ServerInfo{Facility: s, Plan: plan, Quantity: q}
	}
	input := &packngo.CapacityInput{si}
	return input
}

func filterOnUtilization(slugs []string, cr *packngo.CapacityReport, plan, u string) []string {
	r := []string{}
	us := []string{"unavailable", "limited", "normal"}
	desiredIx := findStr(us, u)

	for f, planMap := range *cr {
		for p, planUtilization := range planMap {
			if p != plan {
				continue
			}
			ix := findStr(us, planUtilization.Level)
			if ix >= desiredIx {
				r = append(r, f)
				break
			}
		}
	}
	return r
}
func dataSourcePacketFacilityRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*packngo.Client)
	fl, _, err := client.Facilities.List()
	if err != nil {
		return friendlyError(err)
	}
	pIf, planFilter := d.GetOk("plan")
	plan := pIf.(string)
	qIf, quantityFilter := d.GetOk("quantity")
	quantity := qIf.(int)
	uIf, utilizationFilter := d.GetOk("utilization")
	utilization := uIf.(string)
	fIf, featureFilter := d.GetOk("feature")
	feature := fIf.(string)

	if utilizationFilter && !planFilter {
		return friendlyError(fmt.Errorf("If you set utilization, you also must set plan"))
	}
	if quantityFilter && !planFilter {
		return friendlyError(fmt.Errorf("If you set quantity, you also must set plan"))
	}

	slugs := []string{}
	for _, f := range fl {
		if featureFilter {
			if !contains(f.Features, feature) {
				continue
			}
			slugs = append(slugs, f.Code)
		}
	}

	if (quantityFilter || utilizationFilter) && (len(slugs) > 0) {
		capList, _, err := client.CapacityService.List()
		if err != nil {
			return friendlyError(err)
		}
		slugs = filterOnPlan(slugs, capList, plan)
		if utilizationFilter {
			slugs = filterOnUtilization(slugs, capList, plan, utilization)
		}
		if quantityFilter {
			input := getQuantityCheckInput(slugs, plan, quantity)
			caps, _, err := client.CapacityService.Check(input)
			if err == nil {
				return friendlyError(err)
			}
			slugs = []string{}
			for _, s := range caps.Servers {
				if s.Available {
					slugs = append(slugs, s.Facility)
				}
			}
		}
	}

	d.Set("slugs", slugs)
	d.SetId(time.Now().UTC().String())
	return nil
}
