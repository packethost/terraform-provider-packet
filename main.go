package main

import (
	"github.com/equinix/terraform-provider-equinix-metal/metal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: metal.Provider})
}
