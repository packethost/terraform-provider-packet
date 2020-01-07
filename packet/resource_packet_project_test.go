package packet

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/packethost/packngo"
)

func init() {
	resource.AddTestSweepers("packet_project", &resource.Sweeper{
		Name:         "packet_project",
		Dependencies: []string{"packet_device"},
		F:            testSweepProjects,
	})
}

func testSweepProjects(region string) error {
	log.Printf("[DEBUG] Sweeping projects")
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client for sweeping projects: %s", err)
	}
	client := meta.(*packngo.Client)

	ps, _, err := client.Projects.List(nil)
	if err != nil {
		return fmt.Errorf("Error getting project list for sweepeing projects: %s", err)
	}
	pids := []string{}
	for _, p := range ps {
		if strings.HasPrefix(p.Name, "tfacc-") {
			pids = append(pids, p.ID)
		}
	}
	for _, pid := range pids {
		log.Printf("Removing project %s", pid)
		_, err := client.Projects.Delete(pid)
		if err != nil {
			return fmt.Errorf("Error deleting project %s", err)
		}
	}
	return nil
}
