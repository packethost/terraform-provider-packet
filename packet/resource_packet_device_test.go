package packet

import (
	"fmt"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/packethost/packngo"
)

// Regexp vars for use with resource.ExpectError, resource.TestMatchResourceAttr, etc.
var matchErrConflictsWith = regexp.MustCompile(".* conflicts with .*")
var matchErrMustBeProvided = regexp.MustCompile(".* must be provided when .*")
var matchErrShouldOnlyBeProvided = regexp.MustCompile(".* should only be provided when .*")
var matchErrOutOfRange = regexp.MustCompile(".* is out of range .*")
var matchErrIsNotValid = regexp.MustCompile(".* is not a valid value for.*")
var matchAttrDuration = regexp.MustCompile(`^\dh\d{1,2}m\d{1,2}s$`)

func TestAccPacketDevice_Basic(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_basic, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					testAccCheckPacketDeviceAttributes(&device),
					resource.TestCheckResourceAttr(
						r, "public_ipv4_subnet_size", "31"),
					resource.TestCheckResourceAttr(
						r, "ipxe_script_url", ""),
					resource.TestCheckResourceAttr(
						r, "always_pxe", "false"),
					resource.TestCheckResourceAttrSet(
						r, "root_password"),
					resource.TestCheckResourceAttr(
						r, "spot_instance", "false"),
					resource.TestCheckResourceAttr(
						r, "spot_price_max", ""),
					resource.TestCheckResourceAttr(
						r, "termination_time", ""),
				),
			},
		},
	})
}

func TestAccPacketDevice_RequestSubnet(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_subnet_29"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_request_subnet, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					resource.TestCheckResourceAttr(
						r, "public_ipv4_subnet_size", "29"),
				),
			},
		},
	})
}

func TestAccPacketDevice_IPXEScriptUrl(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_ipxe_script_url"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_ipxe_script_url, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					resource.TestCheckResourceAttr(
						r, "ipxe_script_url", "https://boot.netboot.xyz"),
				),
			},
		},
	})
}

func TestAccPacketDevice_AlwaysPXE(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_always_pxe"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_always_pxe, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					resource.TestCheckResourceAttr(
						r, "always_pxe", "true"),
				),
			},
		},
	})
}

func TestAccPacketDevice_ConflictingFields(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_ipxe_conflict"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_ipxe_conflict, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
				),
				ExpectError: matchErrConflictsWith,
			},
		},
	})
}

func TestAccPacketDevice_IPXEConfigMissing(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_ipxe_config_missing"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_ipxe_missing, rs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
				),
				ExpectError: matchErrMustBeProvided,
			},
		},
	})
}

func TestAccPacketDevice_SpotInstance(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_spot_instance"
	si := "true"
	spm := "0.01"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_spot_instance,
					rs, si, spm, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					resource.TestCheckResourceAttr(
						r, "spot_instance", si),
					resource.TestCheckResourceAttr(
						r, "spot_price_max", spm),
				),
			},
		},
	})
}

func TestAccPacketDevice_SpotTermRFC3339(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_spot_instance"
	si := "true"
	spm := "0.01"
	ttd := time.Duration(time.Hour * 6).Round(terminationTimeRoundVal)
	tn := time.Now()
	tt := tn.Add(ttd).Round(terminationTimeRoundVal)
	ttRFC := tt.Format(time.RFC3339)

	// Test termination_time with RFC3339 format
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_spot_instance,
					rs, si, spm, ttRFC),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					resource.TestCheckResourceAttr(
						r, "spot_instance", si),
					resource.TestCheckResourceAttr(
						r, "spot_price_max", spm),
					resource.TestCheckResourceAttr(
						r, "termination_time", ttRFC),
					resource.TestCheckResourceAttr(
						r, "termination_timestamp", ttRFC),
					resource.TestMatchResourceAttr(
						r, "termination_time_remaining",
						matchAttrDuration),
				),
			},
		},
	})
}

func TestAccPacketDevice_SpotTermDuration(t *testing.T) {
	var device packngo.Device
	rs := acctest.RandString(10)
	r := "packet_device.test_spot_instance"
	si := "true"
	spm := "0.01"
	ttd := time.Duration(time.Hour * 6).Round(terminationTimeRoundVal)
	tn := time.Now()
	tt := tn.Add(ttd).Round(terminationTimeRoundVal)
	ttRFC := tt.Format(time.RFC3339)

	// Test termination_time with Duration format
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_spot_instance,
					rs, si, spm, ttd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPacketDeviceExists(r, &device),
					testAccCheckPacketDeviceNetwork(r),
					resource.TestCheckResourceAttr(
						r, "spot_instance", si),
					resource.TestCheckResourceAttr(
						r, "spot_price_max", spm),
					resource.TestCheckResourceAttr(
						r, "termination_time", ttd.String()),
					resource.TestCheckResourceAttr(
						r, "termination_timestamp", ttRFC),
					resource.TestMatchResourceAttr(
						r, "termination_time_remaining",
						matchAttrDuration),
				),
			},
		},
	})
}

func TestAccPacketDevice_SpotInstanceInvalid(t *testing.T) {
	rs := acctest.RandString(10)
	si := "true"
	ttd := time.Duration(time.Hour * 6)
	tt := time.Now().Add(ttd)
	ttRFC := tt.Format(time.RFC3339)
	spm := "0.01"

	// Invalid termination_time test: Wrong format
	badTime := tt.Format(time.Stamp)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_spot_instance,
					rs, si, spm, badTime),
				ExpectError: matchErrIsNotValid,
			},
		},
	})

	// spot_instance false, but other spot fields set
	si = "false"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_spot_instance,
					rs, si, spm, ttRFC),
				ExpectError: matchErrShouldOnlyBeProvided,
			},
		},
	})
	si = "true" // Reset

	// Missing spot_price_max
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPacketDeviceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckPacketDeviceConfig_spot_price_missing,
					rs, si, ttRFC),
				ExpectError: matchErrMustBeProvided,
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

func testAccCheckPacketDeviceAttributes(device *packngo.Device) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if device.Hostname != "test-device" {
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

var testAccCheckPacketDeviceConfig_basic = `
resource "packet_project" "test" {
    name = "TerraformTestProject-%s"
}

resource "packet_device" "test" {
  hostname         = "test-device"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
}`

var testAccCheckPacketDeviceConfig_request_subnet = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}

resource "packet_device" "test_subnet_29" {
  hostname         = "test-subnet-29"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  public_ipv4_subnet_size = 29
}`

var testAccCheckPacketDeviceConfig_ipxe_script_url = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}

resource "packet_device" "test_ipxe_script_url" {
  hostname         = "test-ipxe-script-url"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "custom_ipxe"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  ipxe_script_url  = "https://boot.netboot.xyz"
}`

var testAccCheckPacketDeviceConfig_always_pxe = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}

resource "packet_device" "test_always_pxe" {
  hostname         = "test-always-pxe"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "custom_ipxe"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  ipxe_script_url  = "https://boot.netboot.xyz"
  always_pxe       = true
}`

var testAccCheckPacketDeviceConfig_ipxe_conflict = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}

resource "packet_device" "test_ipxe_conflict" {
  hostname         = "test-ipxe-conflict"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "custom_ipxe"
  user_data        = "#!ipxe\nset conflict ipxe_script_url"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  ipxe_script_url  = "https://boot.netboot.xyz"
  always_pxe       = true
}`

var testAccCheckPacketDeviceConfig_ipxe_missing = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}

resource "packet_device" "test_ipxe_missing" {
  hostname         = "test-ipxe-missing"
  plan             = "baremetal_0"
  facility         = "sjc1"
  operating_system = "custom_ipxe"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  always_pxe       = true
}`

var testAccCheckPacketDeviceConfig_spot_instance = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}

resource "packet_device" "test_spot_instance" {
  hostname         = "test-spot-instance"
  plan             = "baremetal_0"
  facility         = "nrt1"
  operating_system = "coreos_stable"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  spot_instance    = %s
  spot_price_max   = %s
  termination_time = "%s"
}`

var testAccCheckPacketDeviceConfig_spot_price_missing = `
resource "packet_project" "test" {
  name = "TerraformTestProject-%s"
}

resource "packet_device" "test_spot_instance" {
  hostname         = "test-spot-instance"
  plan             = "baremetal_0"
  facility         = "nrt1"
  operating_system = "coreos_stable"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.test.id}"
  spot_instance    = %s
  termination_time = "%s"
}`
