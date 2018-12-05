package packet

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccPacketFacilites_Basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: `data "packet_facilities" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFacilitiesMeta("data.packet_facilities.test"),
				),
			},
		},
	})
}

func testAccCheckFacilitiesMeta(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find regions data source: %s", n)
		}

		count := ""

		if rs.Primary.ID == "" {
			return errors.New("regions data source ID not set.")
		}

		for _, f := range []string{"slugs", "ids", "features"} {
			count, ok = rs.Primary.Attributes[fmt.Sprintf("%s.#", f)]
			if !ok {
				return fmt.Errorf("can't find '%s' attribute", f)
			}
		}

		noOfNames, err := strconv.Atoi(count)
		if err != nil {
			return errors.New("failed to read number of facilities")
		}
		if noOfNames < 16 {
			return fmt.Errorf("expected at least 16 facilities, received %d, this is most likely a bug",
				noOfNames)
		}

		for i := 0; i < noOfNames; i++ {
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
