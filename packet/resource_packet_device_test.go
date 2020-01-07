package packet

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/packethost/packngo"
)

func init() {
	resource.AddTestSweepers("packet_device", &resource.Sweeper{
		Name: "packet_device",
		F:    testSweepDevices,
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
