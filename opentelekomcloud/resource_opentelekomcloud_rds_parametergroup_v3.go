package opentelekomcloud

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/rds/v3/configurations"
)

func resourceRdsConfigurationV3() *schema.Resource {
	return &schema.Resource{
		Create: resourceRdsConfigurationV3Create,
		Read:   resourceRdsConfigurationV3Read,
		Update: resourceRdsConfigurationV3Update,
		Delete: resourceRdsConfigurationV3Delete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"values": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"datastore": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							DiffSuppressFunc: suppressCaseInsensitive,
						},
						"version": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"configuration_parameters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"restart_required": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"readonly": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"value_range": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func getValues(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("values").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func getDatastore(d *schema.ResourceData) configurations.DataStore {
	datastoreRaw := d.Get("datastore").([]interface{})
	rawMap := datastoreRaw[0].(map[string]interface{})

	datastore := configurations.DataStore{
		Type:    rawMap["type"].(string),
		Version: rawMap["version"].(string),
	}

	log.Printf("[DEBUG] getDatastore: %#v", datastore)
	return datastore
}

func resourceRdsConfigurationV3Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	rdsClient, err := config.rdsV3Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud RDSv3 client: %s", err)
	}

	createOpts := configurations.CreateOpts{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Values:      getValues(d),
		DataStore:   getDatastore(d),
	}
	log.Printf("[DEBUG] CreateOpts: %#v", createOpts)

	configuration, err := configurations.Create(rdsClient, createOpts).Extract()
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud RDSv3 configuration: %s", err)
	}

	log.Printf("[DEBUG] RDSv3 configuration created: %#v", configuration)
	d.SetId(configuration.ID)

	return resourceRdsConfigurationV3Read(d, meta)
}

func resourceRdsConfigurationV3Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client, err := config.rdsV3Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud RDSv3 client: %s", err)
	}
	configuration, err := configurations.Get(client, d.Id()).Extract()

	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error retrieving OpenTelekomCloud RDSv3 configuration: %s", err)
	}
	mErr := multierror.Append(nil,
		d.Set("name", configuration.Name),
		d.Set("description", configuration.Description),
		d.Set("created", configuration.Created),
		d.Set("updated", configuration.Updated),
	)
	if mErr.ErrorOrNil() != nil {
		return mErr
	}

	datastore := []map[string]string{
		{
			"type":    configuration.DatastoreName,
			"version": configuration.DatastoreVersionName,
		},
	}
	if err := d.Set("datastore", datastore); err != nil {
		return err
	}

	parameters := make([]map[string]interface{}, len(configuration.Parameters))
	for i, parameter := range configuration.Parameters {
		parameters[i] = make(map[string]interface{})
		parameters[i]["name"] = parameter.Name
		parameters[i]["value"] = parameter.Value
		parameters[i]["restart_required"] = parameter.RestartRequired
		parameters[i]["readonly"] = parameter.ReadOnly
		parameters[i]["value_range"] = parameter.ValueRange
		parameters[i]["type"] = parameter.Type
		parameters[i]["description"] = parameter.Description
	}
	if err := d.Set("configuration_parameters", parameters); err != nil {
		return err
	}

	return nil
}

func resourceRdsConfigurationV3Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	rdsClient, err := config.rdsV3Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud RDSv3 Client: %s", err)
	}
	var updateOpts configurations.UpdateOpts

	if d.HasChange("name") {
		updateOpts.Name = d.Get("name").(string)
	}
	if d.HasChange("description") {
		updateOpts.Description = d.Get("description").(string)
	}
	if d.HasChange("values") {
		updateOpts.Values = getValues(d)
	}
	log.Printf("[DEBUG] updateOpts: %#v", updateOpts)

	err = configurations.Update(rdsClient, d.Id(), updateOpts).ExtractErr()
	if err != nil {
		return fmt.Errorf("error updating OpenTelekomCloud RDSv3 configuration: %s", err)
	}
	return resourceRdsConfigurationV3Read(d, meta)
}

func resourceRdsConfigurationV3Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	rdsClient, err := config.rdsV3Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud RDSv3 client: %s", err)
	}

	err = configurations.Delete(rdsClient, d.Id()).ExtractErr()
	if err != nil {
		return fmt.Errorf("error deleting OpenTelekomCloud RDSv3 configuration: %s", err)
	}

	d.SetId("")
	return nil
}