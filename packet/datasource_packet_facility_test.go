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
	var totalFacs int
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: `data "packet_facility" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFacilityMeta("data.packet_facility.test", &totalFacs),
				),
			},
			resource.TestStep{
				Config: `data "packet_facility" "test" { features = ["storage", "global_ipv4"] }`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFacilityLessThan("data.packet_facility.test", &totalFacs),
				),
			},
		},
	})
}

func testAccCheckFacilityLessThan(res string, total *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Can't find regions data source: %s", res)
		}

		if rs.Primary.ID == "" {
			return errors.New("regions data source ID not set.")
		}

		countStr, ok := rs.Primary.Attributes["slugs.#"]
		if !ok {
			return fmt.Errorf("can't find 'slugs' attribute")
		}

		count, err := strconv.Atoi(countStr)
		if err != nil {
			return errors.New("failed to read number of facilities")
		}
		if count == 0 {
			return fmt.Errorf("expected some facilities listed, this is most likely a bug")
		}
		return fmt.Errorf("%s should filter out some facilities, and the total should be less than %d, %d", res, total, count)
		if count >= *total {
			return fmt.Errorf("%s should filter out some facilities, and the total should be less than %d, %d", res, total, count)
		}

		return nil
	}
}

func testAccCheckFacilityMeta(res string, total *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Can't find regions data source: %s", res)
		}

		if rs.Primary.ID == "" {
			return errors.New("regions data source ID not set.")
		}

		countStr, ok := rs.Primary.Attributes["slugs.#"]
		if !ok {
			return fmt.Errorf("can't find 'slugs' attribute")
		}

		count, err := strconv.Atoi(countStr)
		if err != nil {
			return errors.New("failed to read number of facilities")
		}
		*total = count
		if count == 0 {
			return fmt.Errorf("expected some facilities listed, this is most likely a bug")
		}

		for i := 0; i < count; i++ {
			idx := "slugs." + strconv.Itoa(i)
			v, ok := rs.Primary.Attributes[idx]
			if !ok {
				return fmt.Errorf("facilities list is corrupt (%q not found), this is definitely a bug", idx)
			}
			if len(v) < 1 {
				return fmt.Errorf("Empty facility slug (%q), this is definitely a bug", idx)
			}
		}
		return nil
	}
}
