package packet

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccPacketFacility_Basic(t *testing.T) {
	totalFacs := new(int)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: `resource "packet_facility" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFacility("packet_facility.test", totalFacs),
				),
			},
			resource.TestStep{
				Config: `resource "packet_facility" "test2" { features = ["storage", "global_ipv4"] }`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFacilityLessThan("packet_facility.test2", totalFacs),
				),
			},
		},
	})
}

func checkFacilitiesAndGetCount(s *terraform.State, res string) (error, int) {
	rs, ok := s.RootModule().Resources[res]
	if !ok {
		return fmt.Errorf("Can't find facility resource: %s", res), 0
	}

	if rs.Primary.ID == "" {
		return errors.New("facilities resource ID not set."), 0
	}

	countStr, ok := rs.Primary.Attributes["slugs.#"]
	if !ok {
		return fmt.Errorf("can't find 'slugs' attribute"), 0
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return errors.New("failed to read number of facilities"), 0
	}
	if count == 0 {
		return fmt.Errorf("expected some facilities listed, this is most likely a bug"), 0
	}
	for i := 0; i < count; i++ {
		idx := "slugs." + strconv.Itoa(i)
		v, ok := rs.Primary.Attributes[idx]
		if !ok {
			return fmt.Errorf("facilities list is corrupt (%q not found), this is definitely a bug", idx), 0
		}
		if len(v) < 1 {
			return fmt.Errorf("Empty facility slug (%q), this is definitely a bug", idx), 0
		}
	}
	return nil, count

}

func testAccCheckFacilityLessThan(res string, total *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err, count := checkFacilitiesAndGetCount(s, res)
		if err != nil {
			return err
		}
		if count >= *total {
			return fmt.Errorf("%q should filter out some facilities, and the total should be less than %d, but is %d", res, *total, count)
		}
		return nil
	}
}

func testAccCheckFacility(res string, total *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err, count := checkFacilitiesAndGetCount(s, res)
		if err != nil {
			return err
		}
		*total = count
		return nil
	}
}
