package packet

import (
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/packethost/packngo"
)

const (
	terminationTimeRoundVal = time.Second * 10
)

type timeParserFunc func(s string) (time.Time, error)

var terminationTimeParsers = []timeParserFunc{
	timeFromRFC3339,
	timeAfterDuration,
}

func resourcePacketDevice() *schema.Resource {
	return &schema.Resource{
		Create: resourcePacketDeviceCreate,
		Read:   resourcePacketDeviceRead,
		Update: resourcePacketDeviceUpdate,
		Delete: resourcePacketDeviceDelete,

		Schema: map[string]*schema.Schema{
			"project_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"operating_system": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"facility": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"plan": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"billing_cycle": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"root_password": &schema.Schema{
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"locked": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},

			"access_public_ipv6": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"access_public_ipv4": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"access_private_ipv4": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"network": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"gateway": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"family": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},

						"cidr": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},

						"public": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"created": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"user_data": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"public_ipv4_subnet_size": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"ipxe_script_url": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_data"},
			},

			"always_pxe": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"hardware_reservation_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "next-available" && len(old) > 0 {
						return true
					}
					return false
				},
			},

			"spot_instance": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"spot_price_max": &schema.Schema{
				Type:         schema.TypeFloat,
				Optional:     true,
				ValidateFunc: floatAtLeast(0.0),
			},

			"termination_time": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: stringTimeParsibleBy(&terminationTimeParsers),
			},

			"termination_time_remaining": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"termination_timestamp": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

//
// schema.Resource CRUD functions
//

func resourcePacketDeviceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*packngo.Client)

	createRequest := &packngo.DeviceCreateRequest{
		Hostname:             d.Get("hostname").(string),
		Plan:                 d.Get("plan").(string),
		Facility:             d.Get("facility").(string),
		OS:                   d.Get("operating_system").(string),
		BillingCycle:         d.Get("billing_cycle").(string),
		ProjectID:            d.Get("project_id").(string),
		PublicIPv4SubnetSize: d.Get("public_ipv4_subnet_size").(int),
	}

	if attr, ok := d.GetOk("user_data"); ok {
		createRequest.UserData = attr.(string)
	}

	if attr, ok := d.GetOk("ipxe_script_url"); ok {
		createRequest.IPXEScriptURL = attr.(string)
	}

	if attr, ok := d.GetOk("hardware_reservation_id"); ok {
		createRequest.HardwareReservationID = attr.(string)
	}

	if createRequest.OS == "custom_ipxe" {
		if createRequest.IPXEScriptURL == "" && createRequest.UserData == "" {
			return friendlyError(errors.New("\"ipxe_script_url\" or \"user_data\"" +
				" must be provided when \"custom_ipxe\" OS is selected."))
		}
	}

	if createRequest.OS != "custom_ipxe" && createRequest.IPXEScriptURL != "" {
		return friendlyError(errors.New("\"ipxe_script_url\" argument provided, but" +
			" OS is not \"custom_ipxe\". Please verify and fix device arguments."))
	}

	if attr, ok := d.GetOk("always_pxe"); ok {
		createRequest.AlwaysPXE = attr.(bool)
	}

	if s, ok := d.GetOk("spot_instance"); ok && s.(bool) {
		createRequest.SpotInstance = s.(bool)

		if sp, ok := d.GetOk("spot_price_max"); ok {
			createRequest.SpotPriceMax = sp.(float64)
		} else {
			return friendlyError(errors.New("\"spot_price_max\" must be " +
				"provided when \"spot_instance\" is true."))
		}

		if tt, ok := d.GetOk("termination_time"); ok {
			t, errs := timeFromParsers(tt.(string), &terminationTimeParsers)
			t = t.Round(terminationTimeRoundVal)
			if errs != nil {
				return fmt.Errorf("%v", errs)
			}
			createRequest.TerminationTime = &packngo.Timestamp{Time: t}
		}
	} else {
		if _, ok := d.GetOk("spot_price_max"); ok {
			return errShouldOnlyBeProvidedWhen("spot_price_max", "spot_instance")
		}

		if _, ok := d.GetOk("termination_time"); ok {
			return errShouldOnlyBeProvidedWhen("termination_time", "spot_instance")
		}
	}

	tags := d.Get("tags.#").(int)
	if tags > 0 {
		createRequest.Tags = make([]string, 0, tags)
		for i := 0; i < tags; i++ {
			key := fmt.Sprintf("tags.%d", i)
			createRequest.Tags = append(createRequest.Tags, d.Get(key).(string))
		}
	}

	newDevice, _, err := client.Devices.Create(createRequest)
	if err != nil {
		return friendlyError(err)
	}

	d.SetId(newDevice.ID)

	// Wait for the device so we can get the networking attributes that show up after a while.
	_, err = waitForDeviceAttribute(d, "active", []string{"queued", "provisioning"}, "state", meta)
	if err != nil {
		if isForbidden(err) {
			// If the device doesn't get to the active state, we can't recover it from here.
			d.SetId("")

			return errors.New("provisioning time limit exceeded; the Packet team will investigate")
		}
		return err
	}

	return resourcePacketDeviceRead(d, meta)
}

func resourcePacketDeviceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*packngo.Client)
	var device *packngo.Device

	pID := d.Get("project_id")
	devices, _, err := client.Devices.List(pID.(string))
	if err != nil {
		return friendlyError(err)
	}

	for _, pDevice := range devices {
		if pDevice.ID == d.Id() {
			device = &pDevice
			break
		}
	}

	// If device isn't in the project's device list, mark as
	// successfully gone.
	if device == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", device.Hostname)
	d.Set("plan", device.Plan.Slug)
	d.Set("facility", device.Facility.Code)
	d.Set("operating_system", device.OS.Slug)
	d.Set("state", device.State)
	d.Set("billing_cycle", device.BillingCycle)
	d.Set("locked", device.Locked)
	d.Set("created", device.Created)
	d.Set("updated", device.Updated)
	d.Set("ipxe_script_url", device.IPXEScriptURL)
	d.Set("always_pxe", device.AlwaysPXE)
	d.Set("root_password", device.RootPassword)

	if len(device.HardwareReservation.Href) > 0 {
		d.Set("hardware_reservation_id", path.Base(device.HardwareReservation.Href))
	}

	tags := make([]string, 0, len(device.Tags))
	for _, tag := range device.Tags {
		tags = append(tags, tag)
	}
	d.Set("tags", tags)

	var (
		ipv4SubnetSize int
		host           string
		networks       = make([]map[string]interface{}, 0, 1)
	)
	for _, ip := range device.Network {
		network := map[string]interface{}{
			"address": ip.Address,
			"gateway": ip.Gateway,
			"family":  ip.AddressFamily,
			"cidr":    ip.CIDR,
			"public":  ip.Public,
		}
		networks = append(networks, network)

		// Initial device IPs are fixed and marked as "Management"
		if ip.Management {
			if ip.AddressFamily == 4 {
				if ip.Public {
					host = ip.Address
					ipv4SubnetSize = ip.CIDR
					d.Set("access_public_ipv4", ip.Address)
				} else {
					d.Set("access_private_ipv4", ip.Address)
				}
			} else {
				d.Set("access_public_ipv6", ip.Address)
			}
		}
	}
	d.Set("network", networks)
	d.Set("public_ipv4_subnet_size", ipv4SubnetSize)
	d.Set("spot_instance", device.SpotInstance)
	d.Set("spot_price_max", device.SpotPriceMax)

	if device.TerminationTime != nil && !device.TerminationTime.Time.IsZero() {
		t := device.TerminationTime.Time.Local()
		rfc3339 := t.Round(terminationTimeRoundVal).Format(time.RFC3339)
		d.Set("termination_timestamp", rfc3339)

		// Round to 10 second intervals
		remaining := time.Until(device.TerminationTime.Time).Round(terminationTimeRoundVal)
		d.Set("termination_time_remaining", remaining.String())
	} else {
		d.Set("termination_time", "")
		d.Set("termination_timestamp", "")
		d.Set("termination_time_remaining", "")
	}

	if host != "" {
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": host,
		})
	}

	return nil
}

func resourcePacketDeviceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*packngo.Client)

	if d.HasChange("locked") {
		var action func(string) (*packngo.Response, error)
		if d.Get("locked").(bool) {
			action = client.Devices.Lock
		} else {
			action = client.Devices.Unlock
		}
		if _, err := action(d.Id()); err != nil {
			return friendlyError(err)
		}
	}

	return resourcePacketDeviceRead(d, meta)
}

func resourcePacketDeviceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*packngo.Client)

	if _, err := client.Devices.Delete(d.Id()); err != nil {
		return friendlyError(err)
	}

	return nil
}

//
// helpers
//

func waitForDeviceAttribute(d *schema.ResourceData, target string, pending []string, attribute string, meta interface{}) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     []string{target},
		Refresh:    newDeviceStateRefreshFunc(d, attribute, meta),
		Timeout:    60 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	return stateConf.WaitForState()
}

func newDeviceStateRefreshFunc(d *schema.ResourceData, attribute string, meta interface{}) resource.StateRefreshFunc {
	client := meta.(*packngo.Client)

	return func() (interface{}, string, error) {
		if err := resourcePacketDeviceRead(d, meta); err != nil {
			return nil, "", err
		}

		if attr, ok := d.GetOk(attribute); ok {
			device, _, err := client.Devices.Get(d.Id())
			if err != nil {
				return nil, "", friendlyError(err)
			}
			return &device, attr.(string), nil
		}

		return nil, "", nil
	}
}

// powerOnAndWait Powers on the device and waits for it to be active.
func powerOnAndWait(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*packngo.Client)
	_, err := client.Devices.PowerOn(d.Id())
	if err != nil {
		return friendlyError(err)
	}

	_, err = waitForDeviceAttribute(d, "active", []string{"off"}, "state", client)
	return err
}

func timeValueExample(layout string) string {
	exampleTime := time.Now().Add(time.Hour + time.Minute).Local()

	return fmt.Sprintf("One hour and one minute from now: \"%s\"",
		exampleTime.Format(layout))
}

func timeDurationExample() string {
	return fmt.Sprint("One hour and one minute from now: \"1h1m\"")
}

func errShouldOnlyBeProvidedWhen(provided, when string) error {
	return friendlyError(fmt.Errorf("\"%s\" should "+
		"only be provided when \"%s\" is true.", provided, when))
}

func timeValueErr(s, layoutName, layout string) error {
	return fmt.Errorf("\"%s\" is not a valid value for time in"+
		" %s format. Below is an example of a valid value.\n%s",
		s, layoutName, timeValueExample(layout))
}

func timeDurationErr(s string) error {
	return fmt.Errorf("\"%s\" is not a valid value for time in"+
		" Duration format. Below is an example of a valid value.\n%s",
		s, timeDurationExample())
}

func timeFromRFC3339(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, timeValueErr(s, "RFC3339", time.RFC3339)
	}

	return t, nil
}

func timeAfterDuration(s string) (time.Time, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return time.Time{}, timeDurationErr(s)
	}

	t := time.Now().Add(d)
	return t, nil
}

func timeFromParsers(v string, tpfs *[]timeParserFunc) (t time.Time, errs []error) {
	for _, tpf := range *tpfs {
		t, err := tpf(v)
		if err == nil {
			return t.Local(), nil
		} else {
			errs = append(errs, err)
		}
	}
	return
}

//
// schema.SchemaDefaultFunc functions
//

func floatAtLeast(min float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warns []string, errs []error) {
		v, ok := i.(float64)
		if !ok {
			errs = append(errs, fmt.Errorf("expected type of %s to be int", k))
			return
		}

		if v < min {
			errs = append(errs, fmt.Errorf("expected %s to be at least (%f), got %f", k, min, v))
			return
		}

		return
	}
}

func stringTimeParsibleBy(parsers *[]timeParserFunc) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warns []string, errs []error) {
		v, ok := i.(string)
		if !ok {
			errs = append(errs, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		// No value is OK. We take that to mean the optional config arg
		// is not set.
		if v == "" {
			return
		}

		_, errs = timeFromParsers(v, parsers)
		return
	}
}
