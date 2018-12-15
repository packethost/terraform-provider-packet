---
layout: "packet"
page_title: "Packet: facility"
sidebar_current: "docs-packet-resource-facility"
description: |-
  Select Packet Facilities
---

# packet\_facility

Use this resource to select facilities in the Packet Host. You can filter them by several criteria: features, plan and utilization level.

This resource has the `keepers` param in the same manner as the (random resources)[https://www.terraform.io/docs/providers/random/index.html#resource-quot-keepers-quot-].

All the arguments are optional, if you don't set them, all the facilities will be returned in the `slugs` attribute.

## Example Usage

```hcl
# create baremetal_0 device in a facility where Global IPv4 is available.

locals {
    plan = "baremetal_0"
}

resource "packet_facility" "example_facility" {
  plan             = "${local.plan}"
  features         = ["global_ipv4"]
  keepers       {
    when = "2018-12-15"
  }
}

resource "packet_project" "cool_project" {
  name             = "cool-project"
}

resource "packet_device" "server" {
  hostname         = "tftest"
  plan             = "${local.plan}"
  facility         = "${packet_facility.example_facility.slugs.0}"
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.cool_project.id}"
}
```

## Argument Reference

 * `features` - (Optional) Only return facilities with features listed in this array. Possible items are `"baremetal", "layer_2", "backend_transfer", "storage", "global_ipv4"`.
 * `plan` - (Optional) Only return facilities where this plan is available. The plan slugs can be found out from (the API docs)[https://www.packet.com/developers/api/#plans]. Set your auth token in the top of the page and see JSON from the API response. 
 * `utilization` - (Optional) The highest utilization level for a plan (set in the other argument) that you accept. Possible values are `"unavailable", "critical", "limited", "normal"`. E.g. if you set this to `"critical"`, only facilities where the selected plan utilization is `"critical", "limited"` and `"normal"` will be returned.
 * `keepers` - (Optional) Key-value pair which should be selected so that they remain the same, until the API should be queried for the facilities again.

## Attributes Reference

 * `slugs` - Array of facilitiy slugs selected by criteria.

