package packet

import (
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

	for f, planMap := range cr {
		if _, ok := planMap[plan]; ok {
			r = r.append(f)
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

func filterOnUtilization(slugs []string, cr *packngo.CapacityReport, u string) []string {
	us := []string{"unavailable", "limited", "normal"}
	desiredIx = findStr(us, u)

	for f, planMap := range cr {
		for plan, planUtilization := rage planMap {
			ix := findStr(us, planUtilization.Level) 
			if ix > desiredIx {
				r = r.append(f)
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
	plan, planFilter := d.GetOk("plan")
	quantity, quantityFilter := d.GetOk("plan")
	utilization, utilizationFilter := d.GetOk("plan")
	feature, featureFilter := d.GetOk("feature")

	if (utilizationFilter && !planFilter) {
		return friendlyError(
			fmt.Errorf("If you set utilization, you also must set plan")
		)
	}
	if (quantityFilter && !planFilter) {
		return friendlyError(
			fmt.Errorf("If you set quantity, you also must set plan")
		)
	}

	slugs := []string{}
	for i, f := range fl {
		if featureFilter {
			if !contains(f.Features, feature) {
				continue
			}
			slugs = append(slugs, f.Slug)
		}
	}


	if (planFilter || quantityFilter || utilizationFilter) && (len(slugs)>0) {
		capList, _, err := c.CapacityService.List()
		if err != nil {
			t.Fatal(err)
		}
		slugs = filterOnPlan(slugs, capList, plan)
		if utilizationFilter {
			slugs = filterOnUtilization(slugs, capList, utilization)
		}
	}

	d.Set("slugs", slugs)
	d.SetId(time.Now().UTC().String())
	return nil
}
