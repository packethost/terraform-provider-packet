package packet

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccPacketCapacityCheck_Basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testPacketCapacityCheck_Basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.packet_capacity_check.test", "available", "true"),
				),
			},
		},
	})
}

func testPacketCapacityCheck_Basic() string {
	return `
data "packet_capacity_check" "test" {
    facility         = "ewr1"
    plan             = "baremetal_0"
    quantity         = 1
}`
}

//					resource.TestCheckResourceAttrSet(
//						"data.packet_capacity_check.test", "available"),
