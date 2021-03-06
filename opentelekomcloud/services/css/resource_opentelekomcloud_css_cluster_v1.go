package css

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/opentelekomcloud/gophertelekomcloud"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func ResourceCssClusterV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceCssClusterV1Create,
		Read:   resourceCssClusterV1Read,
		Update: resourceCssClusterV1Update,
		Delete: resourceCssClusterV1Delete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"node_config": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flavor": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"network_info": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_id": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"security_group_id": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"vpc_id": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
						"volume": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
									},
									"volume_type": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"encryption_key": {
										Type:     schema.TypeString,
										Computed: true,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"enable_https": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"expect_node_num": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},

			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"datastore": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCssClusterV1UserInputParams(d *schema.ResourceData) map[string]interface{} {
	return map[string]interface{}{
		"terraform_resource_data": d,
		"enable_https":            d.Get("enable_https"),
		"expect_node_num":         d.Get("expect_node_num"),
		"name":                    d.Get("name"),
		"node_config":             d.Get("node_config"),
	}
}

func resourceCssClusterV1Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating sdk client, err=%s", err)
	}

	opts := resourceCssClusterV1UserInputParams(d)

	arrayIndex := map[string]int{
		"node_config.network_info": 0,
		"node_config.volume":       0,
		"node_config":              0,
	}

	params, err := buildCssClusterV1CreateParameters(opts, arrayIndex)
	if err != nil {
		return fmt.Errorf("error building the request body of api(create)")
	}
	r, err := sendCssClusterV1CreateRequest(d, params, client)
	if err != nil {
		return fmt.Errorf("error creating CssClusterV1: %s", err)
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	obj, err := asyncWaitCssClusterV1Create(d, config, r, client, timeout)
	if err != nil {
		return err
	}
	id, err := common.NavigateValue(obj, []string{"id"}, nil)
	if err != nil {
		return fmt.Errorf("error constructing id: %s", err)
	}
	d.SetId(id.(string))

	return resourceCssClusterV1Read(d, meta)
}

func resourceCssClusterV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating sdk client, err=%s", err)
	}

	res := make(map[string]interface{})

	v, err := sendCssClusterV1ReadRequest(d, client)
	if err != nil {
		return err
	}
	res["read"] = v

	return setCssClusterV1Properties(d, res)
}

func resourceCssClusterV1Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating sdk client, err=%s", err)
	}

	opts := resourceCssClusterV1UserInputParams(d)

	arrayIndex := map[string]int{
		"node_config.network_info": 0,
		"node_config.volume":       0,
		"node_config":              0,
	}
	timeout := d.Timeout(schema.TimeoutUpdate)

	params, err := buildCssClusterV1ExtendClusterParameters(opts, arrayIndex)
	if err != nil {
		return fmt.Errorf("error building the request body of api(extend_cluster)")
	}
	if e, _ := common.IsEmptyValue(reflect.ValueOf(params)); !e {
		r, err := sendCssClusterV1ExtendClusterRequest(d, params, client)
		if err != nil {
			return err
		}
		_, err = asyncWaitCssClusterV1ExtendCluster(d, config, r, client, timeout)
		if err != nil {
			return err
		}
	}

	return resourceCssClusterV1Read(d, meta)
}

func resourceCssClusterV1Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating sdk client, err=%s", err)
	}

	url, err := common.ReplaceVars(d, "clusters/{id}", nil)
	if err != nil {
		return err
	}
	url = client.ServiceURL(url)

	log.Printf("[DEBUG] Deleting Cluster %q", d.Id())
	r := golangsdk.Result{}
	_, r.Err = client.Delete(url, &golangsdk.RequestOpts{
		OkCodes:      common.SuccessHTTPCodes,
		JSONBody:     nil,
		JSONResponse: nil,
		MoreHeaders:  map[string]string{"Content-Type": "application/json"},
	})
	if r.Err != nil {
		return fmt.Errorf("error deleting Cluster %q: %s", d.Id(), r.Err)
	}

	_, err = common.WaitToFinish(
		[]string{"Done"}, []string{"Pending"},
		d.Timeout(schema.TimeoutCreate),
		1*time.Second,
		func() (interface{}, string, error) {
			_, err := client.Get(url, nil, &golangsdk.RequestOpts{
				MoreHeaders: map[string]string{"Content-Type": "application/json"}})
			if err != nil {
				if _, ok := err.(golangsdk.ErrDefault404); ok {
					return true, "Done", nil
				}
				return nil, "", nil
			}
			return true, "Pending", nil
		},
	)
	return err
}

func buildCssClusterV1CreateParameters(opts map[string]interface{}, arrayIndex map[string]int) (interface{}, error) {
	params := make(map[string]interface{})

	v, err := expandCssClusterV1CreateDiskEncryption(opts, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		params["diskEncryption"] = v
	}

	v, err = expandCssClusterV1CreateHttpsEnable(opts, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		params["httpsEnable"] = v
	}

	v, err = expandCssClusterV1CreateInstance(opts, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		params["instance"] = v
	}

	v, err = common.NavigateValue(opts, []string{"expect_node_num"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		params["instanceNum"] = v
	}

	v, err = common.NavigateValue(opts, []string{"name"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		params["name"] = v
	}

	if len(params) == 0 {
		return params, nil
	}

	params = map[string]interface{}{"cluster": params}

	return params, nil
}

func expandCssClusterV1CreateDiskEncryption(d interface{}, arrayIndex map[string]int) (interface{}, error) {
	req := make(map[string]interface{})

	v, err := common.NavigateValue(d, []string{"node_config", "volume", "encryption_key"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		req["systemCmkid"] = v
	}

	v, err = expandCssClusterV1CreateDiskEncryptionSystemEncrypted(d, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		req["systemEncrypted"] = v
	}

	return req, nil
}

func expandCssClusterV1CreateDiskEncryptionSystemEncrypted(d interface{}, arrayIndex map[string]int) (interface{}, error) {
	v, err := common.NavigateValue(d, []string{"node_config", "volume", "encryption_key"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if v1, ok := v.(string); ok && v1 != "" {
		return "1", nil
	}
	return "0", nil
}

func expandCssClusterV1CreateHttpsEnable(d interface{}, arrayIndex map[string]int) (interface{}, error) {
	v, err := common.NavigateValue(d, []string{"enable_https"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if v1, ok := v.(bool); ok && v1 {
		return "true", nil
	}
	return "false", nil
}

func expandCssClusterV1CreateInstance(d interface{}, arrayIndex map[string]int) (interface{}, error) {
	req := make(map[string]interface{})

	v, err := common.NavigateValue(d, []string{"node_config", "availability_zone"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		req["availability_zone"] = v
	}

	v, err = common.NavigateValue(d, []string{"node_config", "flavor"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		req["flavorRef"] = v
	}

	v, err = expandCssClusterV1CreateInstanceNics(d, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		req["nics"] = v
	}

	v, err = expandCssClusterV1CreateInstanceVolume(d, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		req["volume"] = v
	}

	return req, nil
}

func expandCssClusterV1CreateInstanceNics(d interface{}, arrayIndex map[string]int) (interface{}, error) {
	req := make(map[string]interface{})

	v, err := common.NavigateValue(d, []string{"node_config", "network_info", "network_id"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		req["netId"] = v
	}

	v, err = common.NavigateValue(d, []string{"node_config", "network_info", "security_group_id"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		req["securityGroupId"] = v
	}

	v, err = common.NavigateValue(d, []string{"node_config", "network_info", "vpc_id"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		req["vpcId"] = v
	}

	return req, nil
}

func expandCssClusterV1CreateInstanceVolume(d interface{}, arrayIndex map[string]int) (interface{}, error) {
	req := make(map[string]interface{})

	v, err := common.NavigateValue(d, []string{"node_config", "volume", "size"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		req["size"] = v
	}

	v, err = common.NavigateValue(d, []string{"node_config", "volume", "volume_type"}, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		req["volume_type"] = v
	}

	return req, nil
}

func sendCssClusterV1CreateRequest(_ *schema.ResourceData, params interface{},
	client *golangsdk.ServiceClient) (interface{}, error) {
	url := client.ServiceURL("clusters")

	r := golangsdk.Result{}
	_, r.Err = client.Post(url, params, &r.Body, &golangsdk.RequestOpts{
		OkCodes: common.SuccessHTTPCodes,
	})
	if r.Err != nil {
		return nil, fmt.Errorf("error running api(create): %s", r.Err)
	}
	return r.Body, nil
}

func asyncWaitCssClusterV1Create(d *schema.ResourceData, _ *cfg.Config, result interface{},
	client *golangsdk.ServiceClient, timeout time.Duration) (interface{}, error) {

	data := make(map[string]string)
	pathParameters := map[string][]string{
		"id": {"cluster", "id"},
	}
	for key, path := range pathParameters {
		value, err := common.NavigateValue(result, path, nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving async operation path parameter: %s", err)
		}
		data[key] = value.(string)
	}

	url, err := common.ReplaceVars(d, "clusters/{id}", data)
	if err != nil {
		return nil, err
	}
	url = client.ServiceURL(url)

	return common.WaitToFinish(
		[]string{"200"},
		[]string{"100"},
		timeout, 1*time.Second,
		func() (interface{}, string, error) {
			r := golangsdk.Result{}
			_, r.Err = client.Get(url, &r.Body, &golangsdk.RequestOpts{
				MoreHeaders: map[string]string{"Content-Type": "application/json"}})
			if r.Err != nil {
				return nil, "", nil
			}

			status, err := common.NavigateValue(r.Body, []string{"status"}, nil)
			if err != nil {
				return nil, "", nil
			}
			return r.Body, status.(string), nil
		},
	)
}

func buildCssClusterV1ExtendClusterParameters(opts map[string]interface{}, arrayIndex map[string]int) (interface{}, error) {
	params := make(map[string]interface{})

	v, err := expandCssClusterV1ExtendClusterNodeNum(opts, arrayIndex)
	if err != nil {
		return nil, err
	}
	if e, err := common.IsEmptyValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	} else if !e {
		params["modifySize"] = v
	}

	if len(params) == 0 {
		return params, nil
	}

	params = map[string]interface{}{"grow": params}

	return params, nil
}

func sendCssClusterV1ExtendClusterRequest(d *schema.ResourceData, params interface{},
	client *golangsdk.ServiceClient) (interface{}, error) {
	url, err := common.ReplaceVars(d, "clusters/{id}/extend", nil)
	if err != nil {
		return nil, err
	}
	url = client.ServiceURL(url)

	r := golangsdk.Result{}
	_, r.Err = client.Post(url, params, &r.Body, &golangsdk.RequestOpts{
		OkCodes: common.SuccessHTTPCodes,
	})
	if r.Err != nil {
		return nil, fmt.Errorf("error running api(extend_cluster): %s", r.Err)
	}
	return r.Body, nil
}

func asyncWaitCssClusterV1ExtendCluster(d *schema.ResourceData, _ *cfg.Config, _ interface{}, client *golangsdk.ServiceClient, timeout time.Duration) (interface{}, error) {

	url, err := common.ReplaceVars(d, "clusters/{id}", nil)
	if err != nil {
		return nil, err
	}
	url = client.ServiceURL(url)

	return common.WaitToFinish(
		[]string{"Done"}, []string{"Pending"}, timeout, 1*time.Second,
		func() (interface{}, string, error) {
			r := golangsdk.Result{}
			_, r.Err = client.Get(url, &r.Body, &golangsdk.RequestOpts{
				MoreHeaders: map[string]string{"Content-Type": "application/json"}})
			if r.Err != nil {
				return nil, "", nil
			}

			if checkCssClusterV1ExtendClusterFinished(r.Body) {
				return r.Body, "Done", nil
			}
			return r.Body, "Pending", nil
		},
	)
}

func sendCssClusterV1ReadRequest(d *schema.ResourceData, client *golangsdk.ServiceClient) (interface{}, error) {
	url, err := common.ReplaceVars(d, "clusters/{id}", nil)
	if err != nil {
		return nil, err
	}
	url = client.ServiceURL(url)

	r := golangsdk.Result{}
	_, r.Err = client.Get(url, &r.Body, &golangsdk.RequestOpts{
		MoreHeaders: map[string]string{"Content-Type": "application/json"}})
	if r.Err != nil {
		return nil, fmt.Errorf("error running api(read) for resource(CssClusterV1), error: %s", r.Err)
	}

	return r.Body, nil
}

func setCssClusterV1Properties(d *schema.ResourceData, response map[string]interface{}) error {
	opts := resourceCssClusterV1UserInputParams(d)

	v, err := common.NavigateValue(response, []string{"read", "created"}, nil)
	if err != nil {
		return fmt.Errorf("error reading Cluster:created, err: %s", err)
	}
	if err = d.Set("created", v); err != nil {
		return fmt.Errorf("error setting Cluster:created, err: %s", err)
	}

	v, _ = opts["datastore"]
	v, err = flattenCssClusterV1Datastore(response, nil, v)
	if err != nil {
		return fmt.Errorf("error reading Cluster:datastore, err: %s", err)
	}
	if err = d.Set("datastore", v); err != nil {
		return fmt.Errorf("error setting Cluster:datastore, err: %s", err)
	}

	v, err = common.NavigateValue(response, []string{"read", "httpsEnable"}, nil)
	if err != nil {
		return fmt.Errorf("error reading Cluster:enable_https, err: %s", err)
	}
	if err = d.Set("enable_https", v); err != nil {
		return fmt.Errorf("error setting Cluster:enable_https, err: %s", err)
	}

	v, err = common.NavigateValue(response, []string{"read", "endpoint"}, nil)
	if err != nil {
		return fmt.Errorf("error reading Cluster:endpoint, err: %s", err)
	}
	if err = d.Set("endpoint", v); err != nil {
		return fmt.Errorf("error setting Cluster:endpoint, err: %s", err)
	}

	v, err = common.NavigateValue(response, []string{"read", "name"}, nil)
	if err != nil {
		return fmt.Errorf("error reading Cluster:name, err: %s", err)
	}
	if err = d.Set("name", v); err != nil {
		return fmt.Errorf("error setting Cluster:name, err: %s", err)
	}

	v, _ = opts["node_config"]
	v, err = flattenCssClusterV1NodeConfig(response, nil, v)
	if err != nil {
		return fmt.Errorf("error reading Cluster:node_config, err: %s", err)
	}
	if err = d.Set("node_config", v); err != nil {
		return fmt.Errorf("error setting Cluster:node_config, err: %s", err)
	}

	v, _ = opts["nodes"]
	v, err = flattenCssClusterV1Nodes(response, nil, v)
	if err != nil {
		return fmt.Errorf("error reading Cluster:nodes, err: %s", err)
	}
	if err = d.Set("nodes", v); err != nil {
		return fmt.Errorf("error setting Cluster:nodes, err: %s", err)
	}

	v, err = common.NavigateValue(response, []string{"read", "updated"}, nil)
	if err != nil {
		return fmt.Errorf("error reading Cluster:updated, err: %s", err)
	}
	if err = d.Set("updated", v); err != nil {
		return fmt.Errorf("error setting Cluster:updated, err: %s", err)
	}

	return nil
}

func flattenCssClusterV1Datastore(d interface{}, arrayIndex map[string]int, currentValue interface{}) (interface{}, error) {
	result, ok := currentValue.([]interface{})
	if !ok || len(result) == 0 {
		result = make([]interface{}, 1, 1)
	}
	if result[0] == nil {
		result[0] = make(map[string]interface{})
	}
	r := result[0].(map[string]interface{})

	v, err := common.NavigateValue(d, []string{"read", "datastore", "type"}, arrayIndex)
	if err != nil {
		return nil, fmt.Errorf("error reading Cluster:type, err: %s", err)
	}
	r["type"] = v

	v, err = common.NavigateValue(d, []string{"read", "datastore", "version"}, arrayIndex)
	if err != nil {
		return nil, fmt.Errorf("error reading Cluster:version, err: %s", err)
	}
	r["version"] = v

	return result, nil
}

func flattenCssClusterV1NodeConfig(d interface{}, arrayIndex map[string]int, currentValue interface{}) (interface{}, error) {
	result, ok := currentValue.([]interface{})
	if !ok || len(result) == 0 {
		result = make([]interface{}, 1, 1)
	}
	if result[0] == nil {
		result[0] = make(map[string]interface{})
	}
	r := result[0].(map[string]interface{})

	v, _ := r["network_info"]
	v, err := flattenCssClusterV1NodeConfigNetworkInfo(d, arrayIndex, v)
	if err != nil {
		return nil, fmt.Errorf("error reading Cluster:network_info, err: %s", err)
	}
	r["network_info"] = v

	v, _ = r["volume"]
	v, err = flattenCssClusterV1NodeConfigVolume(d, arrayIndex, v)
	if err != nil {
		return nil, fmt.Errorf("error reading Cluster:volume, err: %s", err)
	}
	r["volume"] = v

	return result, nil
}

func flattenCssClusterV1NodeConfigNetworkInfo(d interface{}, arrayIndex map[string]int, currentValue interface{}) (interface{}, error) {
	result, ok := currentValue.([]interface{})
	if !ok || len(result) == 0 {
		result = make([]interface{}, 1, 1)
	}
	if result[0] == nil {
		result[0] = make(map[string]interface{})
	}
	r := result[0].(map[string]interface{})

	v, err := common.NavigateValue(d, []string{"read", "subnetId"}, arrayIndex)
	if err != nil {
		return nil, fmt.Errorf("error reading Cluster:network_id, err: %s", err)
	}
	r["network_id"] = v

	v, err = common.NavigateValue(d, []string{"read", "securityGroupId"}, arrayIndex)
	if err != nil {
		return nil, fmt.Errorf("error reading Cluster:security_group_id, err: %s", err)
	}
	r["security_group_id"] = v

	return result, nil
}

func flattenCssClusterV1NodeConfigVolume(d interface{}, arrayIndex map[string]int, currentValue interface{}) (interface{}, error) {
	result, ok := currentValue.([]interface{})
	if !ok || len(result) == 0 {
		result = make([]interface{}, 1, 1)
	}
	if result[0] == nil {
		result[0] = make(map[string]interface{})
	}
	r := result[0].(map[string]interface{})

	v, err := common.NavigateValue(d, []string{"read", "cmkId"}, arrayIndex)
	if err != nil {
		v = ""
	}
	r["encryption_key"] = v

	return result, nil
}

func flattenCssClusterV1Nodes(d interface{}, arrayIndex map[string]int, currentValue interface{}) (interface{}, error) {
	result, ok := currentValue.([]interface{})
	if !ok || len(result) == 0 {
		v, err := common.NavigateValue(d, []string{"read", "instances"}, arrayIndex)
		if err != nil {
			return nil, err
		}
		n := len(v.([]interface{}))
		result = make([]interface{}, n, n)
	}

	newArrayIndex := make(map[string]int)
	if arrayIndex != nil {
		for k, v := range arrayIndex {
			newArrayIndex[k] = v
		}
	}

	for i := 0; i < len(result); i++ {
		newArrayIndex["read.instances"] = i
		if result[i] == nil {
			result[i] = make(map[string]interface{})
		}
		r := result[i].(map[string]interface{})

		v, err := common.NavigateValue(d, []string{"read", "instances", "id"}, newArrayIndex)
		if err != nil {
			return nil, fmt.Errorf("error reading Cluster:id, err: %s", err)
		}
		r["id"] = v

		v, err = common.NavigateValue(d, []string{"read", "instances", "name"}, newArrayIndex)
		if err != nil {
			return nil, fmt.Errorf("error reading Cluster:name, err: %s", err)
		}
		r["name"] = v

		v, err = common.NavigateValue(d, []string{"read", "instances", "type"}, newArrayIndex)
		if err != nil {
			return nil, fmt.Errorf("error reading Cluster:type, err: %s", err)
		}
		r["type"] = v
	}

	return result, nil
}
