package fw

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/extensions/fwaas_v2/policies"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func ResourceFWPolicyV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceFWPolicyV2Create,
		Read:   resourceFWPolicyV2Read,
		Update: resourceFWPolicyV2Update,
		Delete: resourceFWPolicyV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"audited": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"shared": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"rules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"value_specs": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFWPolicyV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*cfg.Config)
	networkingClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("Error creating OpenTelekomCloud networking client: %s", err)
	}

	v := d.Get("rules").([]interface{})

	log.Printf("[DEBUG] Rules found : %#v", v)
	log.Printf("[DEBUG] Rules count : %d", len(v))

	rules := make([]string, len(v))
	for i, v := range v {
		rules[i] = v.(string)
	}

	audited := d.Get("audited").(bool)

	opts := PolicyCreateOpts{
		policies.CreateOpts{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Audited:     &audited,
			TenantID:    d.Get("tenant_id").(string),
			Rules:       rules,
		},
		common.MapValueSpecs(d),
	}

	if r, ok := d.GetOk("shared"); ok {
		shared := r.(bool)
		opts.Shared = &shared
	}

	log.Printf("[DEBUG] Create firewall policy: %#v", opts)

	policy, err := policies.Create(networkingClient, opts).Extract()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Firewall policy created: %#v", policy)

	d.SetId(policy.ID)

	return resourceFWPolicyV2Read(d, meta)
}

func resourceFWPolicyV2Read(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Retrieve information about firewall policy: %s", d.Id())

	config := meta.(*cfg.Config)
	networkingClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("Error creating OpenTelekomCloud networking client: %s", err)
	}

	policy, err := policies.Get(networkingClient, d.Id()).Extract()
	if err != nil {
		return common.CheckDeleted(d, err, "FW policy")
	}

	log.Printf("[DEBUG] Read OpenTelekomCloud Firewall Policy %s: %#v", d.Id(), policy)

	d.Set("name", policy.Name)
	d.Set("description", policy.Description)
	d.Set("shared", policy.Shared)
	d.Set("audited", policy.Audited)
	d.Set("tenant_id", policy.TenantID)
	if err := d.Set("rules", policy.Rules); err != nil {
		return fmt.Errorf("[DEBUG] Error saving rules to state for OpenTelekomCloud firewall policy (%s): %s", d.Id(), err)
	}
	d.Set("region", config.GetRegion(d))

	return nil
}

func resourceFWPolicyV2Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*cfg.Config)
	networkingClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("Error creating OpenTelekomCloud networking client: %s", err)
	}

	opts := policies.UpdateOpts{}

	if d.HasChange("name") {
		opts.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		opts.Description = d.Get("description").(string)
	}

	if d.HasChange("rules") {
		v := d.Get("rules").([]interface{})

		log.Printf("[DEBUG] Rules found : %#v", v)
		log.Printf("[DEBUG] Rules count : %d", len(v))

		rules := make([]string, len(v))
		for i, v := range v {
			rules[i] = v.(string)
		}
		opts.Rules = rules
	}

	log.Printf("[DEBUG] Updating firewall policy with id %s: %#v", d.Id(), opts)

	err = policies.Update(networkingClient, d.Id(), opts).Err
	if err != nil {
		return err
	}

	return resourceFWPolicyV2Read(d, meta)
}

func resourceFWPolicyV2Delete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Destroy firewall policy: %s", d.Id())

	config := meta.(*cfg.Config)
	networkingClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("Error creating OpenTelekomCloud networking client: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    waitForFirewallPolicyDeletion(networkingClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      0,
		MinTimeout: 2 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return err
	}

	return nil
}

func waitForFirewallPolicyDeletion(networkingClient *golangsdk.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := policies.Delete(networkingClient, id).Err
		if err == nil {
			return "", "DELETED", nil
		}

		if errCode, ok := err.(golangsdk.ErrUnexpectedResponseCode); ok {
			if errCode.Actual == 409 {
				// This error usually means that the policy is attached
				// to a firewall. At this point, the firewall is probably
				// being delete. So, we retry a few times.
				return nil, "ACTIVE", nil
			}
		}

		return nil, "ACTIVE", err
	}
}
