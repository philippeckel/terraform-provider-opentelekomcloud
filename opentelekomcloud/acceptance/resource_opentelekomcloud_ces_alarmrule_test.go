package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cloudeyeservice/alarmrule"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func TestCESAlarmRule_basic(t *testing.T) {
	var ar alarmrule.AlarmRule

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCESAlarmRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCESAlarmRule_basic,
				Check: resource.ComposeTestCheckFunc(
					testCESAlarmRuleExists("opentelekomcloud_ces_alarmrule.alarmrule_1", &ar),
				),
			},
			{
				Config: testCESAlarmRule_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"opentelekomcloud_ces_alarmrule.alarmrule_1", "alarm_enabled", "false"),
				),
			},
		},
	})
}

func testCESAlarmRuleDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*cfg.Config)
	networkingClient, err := config.CesV1Client(OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("Error creating OpenTelekomCloud ces client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_ces_alarmrule" {
			continue
		}

		id := rs.Primary.ID
		_, err := alarmrule.Get(networkingClient, id).Extract()
		if err == nil {
			return fmt.Errorf("Alarm rule still exists")
		}
	}

	return nil
}

func testCESAlarmRuleExists(n string, ar *alarmrule.AlarmRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*cfg.Config)
		networkingClient, err := config.CesV1Client(OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating OpenTelekomCloud ces client: %s", err)
		}

		id := rs.Primary.ID
		found, err := alarmrule.Get(networkingClient, id).Extract()
		if err != nil {
			return err
		}

		*ar = *found

		return nil
	}
}

var testCESAlarmRule_basic = fmt.Sprintf(`
resource "opentelekomcloud_compute_instance_v2" "vm_1" {
  name = "instance_1"
  network {
    uuid = "%s"
  }
}

resource "opentelekomcloud_smn_topic_v2" "topic_1" {
  name		  = "topic_1"
  display_name    = "The display name of topic_1"
}

resource "opentelekomcloud_ces_alarmrule" "alarmrule_1" {
  alarm_name = "alarm_rule1"

  metric {
    namespace = "SYS.ECS"
    metric_name = "network_outgoing_bytes_rate_inband"
    dimensions {
        name = "instance_id"
        value = opentelekomcloud_compute_instance_v2.vm_1.id
    }
  }
  condition  {
    period = 300
    filter = "average"
    comparison_operator = ">"
    value = 6
    unit = "B/s"
    count = 1
  }
  alarm_action_enabled = false

  alarm_actions {
    type = "notification"
    notification_list = [
      opentelekomcloud_smn_topic_v2.topic_1.topic_urn
    ]
  }
}
`, OS_NETWORK_ID)

var testCESAlarmRule_update = fmt.Sprintf(`
resource "opentelekomcloud_compute_instance_v2" "vm_1" {
  name = "instance_1"
  network {
    uuid = "%s"
  }
}

resource "opentelekomcloud_smn_topic_v2" "topic_1" {
  name		  = "topic_1"
  display_name    = "The display name of topic_1"
}

resource "opentelekomcloud_ces_alarmrule" "alarmrule_1" {
  alarm_name = "alarm_rule1"

  metric {
    namespace = "SYS.ECS"
    metric_name = "network_outgoing_bytes_rate_inband"
    dimensions {
        name = "instance_id"
        value = opentelekomcloud_compute_instance_v2.vm_1.id
    }
  }
  condition  {
    period = 300
    filter = "average"
    comparison_operator = ">"
    value = 6
    unit = "B/s"
    count = 1
  }
  alarm_action_enabled = false
  alarm_enabled = false

  alarm_actions {
    type = "notification"
    notification_list = [
      opentelekomcloud_smn_topic_v2.topic_1.topic_urn
    ]
  }
}
`, OS_NETWORK_ID)
