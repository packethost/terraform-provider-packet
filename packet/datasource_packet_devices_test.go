package packet

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func testAccDatasourceDevicesConfig(projSuffix string) string {
	return fmt.Sprintf(`
resource "packet_project" "test" {
    name = "TerraformTestProject-%s"
}

resource "packet_device" "d1" {
  hostname         = "d1"
  plan             = "t1.small.x86"
  facility         = "sjc1"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
}

resource "packet_device" "d2" {
  hostname         = "d2"
  plan             = "t1.small.x86"
  facility         = "sjc1"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
}

resource "packet_device" "d3" {
  hostname         = "d3"
  plan             = "t1.small.x86"
  facility         = "sjc1"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
}

data "packet_devices" "testdevs" {
  names = ["d1", "d2"]
  depends_on = ["packet_device.d1", "packet_device.d2", "packet_device.d3]
}

`, projSuffix)
}

func TestAccDataSourcePacketDevices_Basic(t *testing.T) {
	rs := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: testAccDataSourceDevicesConfig(rs),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.packet_devices", "id", "coreos_alpha"),
				),
			},
		},
	})
}
