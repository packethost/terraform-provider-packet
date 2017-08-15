package packet

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/packethost/packngo"
)

func TestAccPacketDevice_Basic(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_basic, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists("packet_device.terraform_test_device", &device),
					testAccCheckPacketDeviceAttributes(&device),
					testAccCheckPacketDevicePublicIPv4Cidr(&device, 31),
				),
			},
		},
	})
}

func TestAccPacketDevice_RequestSubnet(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_request_subnet, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists("packet_device.terraform_test_device_subnet_29", &device),
					testAccCheckPacketDevicePublicIPv4Cidr(&device, 29),
				),
			},
		},
	})
}

func testAccCheckPacketDeviceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*packngo.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "packet_device" {
			continue
		}
		if _, _, err := client.Devices.Get(rs.Primary.ID); err == nil {
			return fmt.Errorf("Device still exists")
		}
	}
	return nil
}

func getPublicIPv4Cidr(device *packngo.Device) (int, error) {
	for _, ipa := range device.Network {
		if ipa.AddressFamily == 4 && ipa.Public {
			return ipa.Cidr, nil
		}
	}
	return 0, fmt.Errorf("device %s does not have a public IPv4", device.ID)
}

func testAccCheckPacketDevicePublicIPv4Cidr(device *packngo.Device, cidr int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		devCidr, err := getPublicIPv4Cidr(device)
		if err != nil {
			return err
		}
		if devCidr != cidr {
			return fmt.Errorf("The CIDR prefix of device %s is %d, but is %d instead", device.ID, devCidr, cidr)
		}
		return nil
	}
}

func testAccCheckPacketDeviceAttributes(device *packngo.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if device.Hostname != "terraform-test-device" {
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

		foundDevice, _, err := client.Devices.Get(rs.Primary.ID)
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

var testAccCheckPacketDeviceConfig_basic = `
resource "packet_project" "terraform_test_project" {
    name = "TerraformTestProject-%s"
}

resource "packet_device" "terraform_test_device" {
  hostname         = "terraform-test-device"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.terraform_test_project.id}"
}`

var testAccCheckPacketDeviceConfig_request_subnet = `
resource "packet_project" "terraform_test_project" {
    name = "TerraformTestProject-%s"
}

resource "packet_device" "terraform_test_device_subnet_29" {
  hostname         = "terraform-test-device-subnet-29"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.terraform_test_project.id}"
  public_ipv4_subnet_size = 29
}`
