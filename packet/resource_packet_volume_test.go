package packet

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/packethost/packngo"
)

func init() {
	resource.AddTestSweepers("packet_volume", &resource.Sweeper{
		Name: "packet_volume",
		F:    testSweepVolumes,
	})
}

func testSweepVolumes(region string) error {
	log.Printf("[DEBUG] Sweeping volumes")
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client for sweeping volumes: %s", err)
	}
	client := meta.(*packngo.Client)

	ps, _, err := client.Projects.List(nil)
	if err != nil {
		return fmt.Errorf("Error getting project list for sweepeing volumes: %s", err)
	}
	pids := []string{}
	for _, p := range ps {
		if strings.HasPrefix(p.Name, "tfacc-") {
			pids = append(pids, p.ID)
		}
	}
	vids := []string{}
	for _, pid := range pids {
		vs, _, err := client.Volumes.List(pid, nil)
		if err != nil {
			return fmt.Errorf("Error listing volumes to sweep: %s", err)
		}
		for _, v := range vs {
			vids = append(vids, v.ID)
		}
	}

	for _, vid := range vids {
		log.Printf("Removing volume %s", vid)
		_, err := client.Volumes.Delete(vid)
		if err != nil {
			return fmt.Errorf("Error deleting volume %s", err)
		}
	}
	return nil
}
