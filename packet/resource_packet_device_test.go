package packet

import (
	"fmt"
	"log"
	"net"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/packethost/packngo"
)

func init() {
	resource.AddTestSweepers("packet_device", &resource.Sweeper{
		Name:         "packet_device",
		F:            testSweepDevices,
		Dependencies: []string{"packet_volume"},
	})
}

func testSweepDevices(region string) error {
	log.Printf("[DEBUG] Sweeping devices")
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client for sweeping devices: %s", err)
	}
	client := meta.(*packngo.Client)

	ps, _, err := client.Projects.List(nil)
	if err != nil {
		return fmt.Errorf("Error getting project list for sweepeing devices: %s", err)
	}
	pids := []string{}
	for _, p := range ps {
		if strings.HasPrefix(p.Name, "tfacc-") {
			pids = append(pids, p.ID)
		}
	}
	dids := []string{}
	for _, pid := range pids {
		ds, _, err := client.Devices.List(pid, nil)
		if err != nil {
			return fmt.Errorf("Error listing devices to sweep: %s", err)
		}
		for _, d := range ds {
			dids = append(dids, d.ID)
		}
	}

	for _, did := range dids {
		log.Printf("Removing device %s", did)
		_, err := client.Devices.Delete(did, true)
		if err != nil {
			return fmt.Errorf("Error deleting device %s", err)
		}
	}
	return nil
}

func testAccCheckPacketDeviceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*packngo.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "packet_device" {
			continue
		}
		if _, _, err := client.Devices.Get(rs.Primary.ID, nil); err == nil {
			return fmt.Errorf("Device still exists")
		}
	}
	return nil
}

func testAccCheckPacketDeviceNetwork(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var ip net.IP
		var k, v string
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		k = "access_public_ipv6"
		v = rs.Primary.Attributes[k]
		ip = net.ParseIP(v)
		if ip == nil {
			return fmt.Errorf("\"%s\" is not a valid IP address: %s",
				k, v)
		}

		k = "access_public_ipv4"
		v = rs.Primary.Attributes[k]
		ip = net.ParseIP(v)
		if ip == nil {
			return fmt.Errorf("\"%s\" is not a valid IP address: %s",
				k, v)
		}

		k = "access_private_ipv4"
		v = rs.Primary.Attributes[k]
		ip = net.ParseIP(v)
		if ip == nil {
			return fmt.Errorf("\"%s\" is not a valid IP address: %s",
				k, v)
		}

		return nil
	}
}

func testAccCheckPacketDeviceConfig_basic(projSuffix string) string {
	return fmt.Sprintf(`
resource "packet_project" "test" {
    name = "tfacc-device-%s"
}

resource "packet_device" "test" {
  hostname         = "tfacc-test-device"
  plan             = "t1.small.x86"
  facilities       = ["sjc1"]
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
}`, projSuffix)
}

func TestAccPacketDevice_Basic(t *testing.T) {
	var device, deviceWithUserData packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test"
	testUD := `#cloud-config
runcmd:
 - [ ls, -l, / ]
`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPacketDeviceConfig_basic(rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					testAccCheckPacketDeviceAttributes(&device),
					resource.TestCheckResourceAttr(
						r, "public_ipv4_subnet_size", "31"),
					resource.TestCheckResourceAttr(
						r, "network_type", "layer3"),
					resource.TestCheckResourceAttr(
						r, "ipxe_script_url", ""),
					resource.TestCheckResourceAttr(
						r, "always_pxe", "false"),
					resource.TestCheckResourceAttrSet(
						r, "root_password"),
					resource.TestCheckResourceAttrPair(
						r, "deployed_facility", r, "facilities.0"),
				),
			},
			{
				Config: testAccCheckPacketDeviceConfig_basicUserData(rs, testUD),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &deviceWithUserData),
					resource.TestCheckResourceAttr(r, "user_data", testUD),
					testAccCheckPacketSameDevice(t, &device, &deviceWithUserData),
				),
			},
		},
	})
}

func testAccCheckPacketSameDevice(t *testing.T, before, after *packngo.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID != after.ID {
			t.Fatalf("Expected device to be the same, but it was recreated: %s -> %s", before.ID, after.ID)
		}
		return nil
	}
}

func testAccCheckPacketDeviceConfig_basicUserData(projSuffix, ud string) string {
	return fmt.Sprintf(`
resource "packet_project" "test" {
    name = "tfacc-device-%s"
}

locals {
	ud = <<EOS
%sEOS
}

resource "packet_device" "test" {
  hostname         = "tfacc-test-device"
  plan             = "t1.small.x86"
  facilities       = ["sjc1"]
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  user_data        = "${local.ud}"
}`, projSuffix, ud)
}

func testAccCheckPacketDeviceAttributes(device *packngo.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if device.Hostname != "tfacc-test-device" {
			return fmt.Errorf("Bad name: %s", device.Hostname)
		}
		if device.State != "active" {
			return fmt.Errorf("Device should be 'active', not '%s'", device.State)
		}

		return nil
	}
}

func testAccCheckPacketDeviceExists(n string, device *packngo.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*packngo.Client)

		foundDevice, _, err := client.Devices.Get(rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if foundDevice.ID != rs.Primary.ID {
			return fmt.Errorf("Record not found: %v - %v", rs.Primary.ID, foundDevice)
		}

		*device = *foundDevice

		return nil
	}
}
