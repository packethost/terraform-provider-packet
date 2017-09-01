---
layout: "packet"
page_title: "Packet: packet_device"
sidebar_current: "docs-packet-resource-device"
description: |-
  Provides a Packet device resource. This can be used to create, modify, and delete devices.
---

# packet\_device

Provides a Packet device resource. This can be used to create,
modify, and delete devices.

~> **Note:** All arguments including the root_password and user_data will be stored in
 the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).


## Example Usage

```hcl
# Create a device and add it to cool_project
resource "packet_device" "web1" {
  hostname         = "tf.coreos2"
  plan             = "baremetal_1"
  facility         = "ewr1"
  operating_system = "coreos_stable"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.cool_project.id}"
}
```

```hcl
# Same as above, but boot via iPXE initially, using the Ignition Provider for provisioning
resource "packet_device" "pxe1" {
  hostname         = "tf.coreos2-pxe"
  plan             = "baremetal_1"
  facility         = "ewr1"
  operating_system = "custom_ipxe"
  billing_cycle    = "hourly"
  project_id       = "${packet_project.cool_project.id}"
  ipxe_script_url  = "https://rawgit.com/cloudnativelabs/pxe/master/packet/coreos-stable-packet.ipxe"
  always_pxe       = "false"
  user_data        = "${data.ignition_config.example.rendered}"
}
```

## Argument Reference

The following arguments are supported:

* `hostname` - (Required) The device name
* `project_id` - (Required) The id of the project in which to create the device
* `operating_system` - (Required) The operating system slug
* `facility` - (Required) The facility in which to create the device
* `plan` - (Required) The hardware config slug
* `billing_cycle` - (Required) monthly or hourly
* `user_data` (Optional) - A string of the desired User Data for the device.
* `public_ipv4_subnet_size` (Optional) - Size of allocated subnet, more
  information is in the
  [Custom Subnet Size](https://help.packet.net/technical/networking/custom-subnet-size) doc.
* `spot_instance` (Optional) - If true, create a preemptible device using
  `spot_price_max` as a bid. See the
  [documentation](https://help.packet.net/technical/deployment-options/spot-market)
  for more details
* `spot_price_max` (Optional) - Spot market bid price. Must be set when
  `spot_instance` is true.
* `termination_time` (Optional) - Set this to automatically terminate the device
  at a certain time. Accepts RFC3339 (e.g. `2018-09-21T19:20:01-05:00`) or
  Duration (e.g. `6h20m`) formats. Should only be provided if `spot_instance` is
  true.
* `ipxe_script_url` (Optional) - URL pointing to a hosted iPXE script. More
  information is in the
  [Custom iPXE](https://help.packet.net/technical/infrastructure/custom-ipxe)
  doc.
* `always_pxe` (Optional) - If true, a device with OS `custom_ipxe` will
  continue to boot via iPXE on reboots.
* `hardware_reservation_id` (Optional) - The id of hardware reservation where you want this device deployed, or `next-available` if you want to pick your next available reservation automatically.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the device
* `hostname`- The hostname of the device
* `project_id`- The ID of the project the device belongs to
* `facility` - The facility the device is in
* `plan` - The hardware config of the device
* `network` - The device's private and public IP (v4 and v6) network details
* `access_public_ipv6` - The ipv6 maintenance IP assigned to the device
* `access_public_ipv4` - The ipv4 maintenance IP assigned to the device
* `access_private_ipv4` - The ipv4 private IP assigned to the device
* `locked` - Whether the device is locked
* `billing_cycle` - The billing cycle of the device (monthly or hourly)
* `operating_system` - The operating system running on the device
* `spot_instance` - True if this device is a preemptible spot instance
* `spot_price_max` - User-provided bid price if device is a `spot_instance`
* `termination_time` - User-provided date or duration for terminating a
  `spot_instance`
* `termination_timestamp` - Convenience attributes that stores the RFC3339
  formatted date of the device. Useful if `termination_time` is set as a
  duration instead of a date.
* `termination_time_remaining` - Convenience attribute that stores the time
  remaining before `termination_timestamp` in duration format.
* `ipxe_script_url` - User-provided iPXE script URL
* `always_pxe` - True if device will always reboot with iPXE boot
* `state` - The status of the device
* `created` - The timestamp for when the device was created
* `updated` - The timestamp for the last time the device was updated
* `tags` - Tags attached to the device
* `hardware_reservation_id` - The id of hardware reservation which this device occupies
* `root_password` - Root password to the server (disabled after 24 hours)
