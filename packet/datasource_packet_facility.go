package packet

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/packethost/packngo"
)

func dataSourceFacility() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePacketFacilityRead,
		Schema: map[string]*schema.Schema{
			"slugs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"features": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MinItems: 1,
				Optional: true,
			},
			"plan": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(packngo.DevicePlans, false),
			},
			"utilization": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(packngo.UtilizationLevels, false),
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

func filterOnUtilization(slugs []string, cr *packngo.CapacityReport, plan, u string) []string {
	r := []string{}
	desiredIx := findStr(packngo.UtilizationLevels, u)

	for f, planMap := range *cr {
		for p, planUtilization := range planMap {
			if p != plan {
				continue
			}
			ix := findStr(packngo.UtilizationLevels, planUtilization.Level)
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
	pIf, planFilter := d.GetOk("plan")
	plan := pIf.(string)
	uIf, utilizationFilter := d.GetOk("utilization")
	utilization := uIf.(string)
	fIf, featuresFilter := d.GetOk("features")
	featureSet := fIf.(*schema.Set)
	featureSlice := convertStringArr(featureSet.List())

	if featuresFilter {
		for _, f := range featureSlice {
			if !contains(packngo.FacilityFeatures, f) {
				return fmt.Errorf("%q is not a valid Packet facility feature, only %+v are allowed", f, packngo.FacilityFeatures)
			}
		}
	}

	if utilizationFilter && !planFilter {
		return friendlyError(fmt.Errorf("If you set utilization, you also must set plan"))
	}

	slugs := []string{}

	fl, _, err := client.Facilities.List(nil)
	if err != nil {
		return friendlyError(err)
	}

	for _, f := range fl {
		if featuresFilter {
			currentFacFeatureSet := schema.NewSet(
				featureSet.F, convertInterfaceArr(f.Features))

			diff := featureSet.Difference(currentFacFeatureSet)

			if diff.Len() > 0 {
				continue
			}
		}
		slugs = append(slugs, f.Code)
	}

	if (utilizationFilter || planFilter) && (len(slugs) > 0) {
		capList, _, err := client.CapacityService.List()
		if err != nil {
			return friendlyError(err)
		}
		slugs = filterOnPlan(slugs, capList, plan)
		if utilizationFilter {
			slugs = filterOnUtilization(slugs, capList, plan, utilization)
		}
	}

	d.Set("slugs", slugs)
	d.SetId(time.Now().UTC().String())
	return nil
}
