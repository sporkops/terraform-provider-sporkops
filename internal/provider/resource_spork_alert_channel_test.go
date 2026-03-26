//go:build acceptance

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAlertChannelResource_email(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_alert_channel" "test" {
  name = "Test Email Channel"
  type = "email"
  config = {
    to = "test@example.com"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("spork_alert_channel.test", "id"),
					resource.TestCheckResourceAttr("spork_alert_channel.test", "type", "email"),
					resource.TestCheckResourceAttr("spork_alert_channel.test", "name", "Test Email Channel"),
					resource.TestCheckResourceAttr("spork_alert_channel.test", "config.to", "test@example.com"),
				),
			},
			{
				ResourceName:      "spork_alert_channel.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAlertChannelResource_webhook(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_alert_channel" "test" {
  name = "Test Webhook"
  type = "webhook"
  config = {
    url = "https://hooks.example.com/alert"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("spork_alert_channel.test", "id"),
					resource.TestCheckResourceAttr("spork_alert_channel.test", "type", "webhook"),
					resource.TestCheckResourceAttr("spork_alert_channel.test", "config.url", "https://hooks.example.com/alert"),
				),
			},
		},
	})
}

func TestAccAlertChannelResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_alert_channel" "test" {
  name = "Original Channel"
  type = "email"
  config = {
    to = "original@example.com"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("spork_alert_channel.test", "name", "Original Channel"),
					resource.TestCheckResourceAttr("spork_alert_channel.test", "config.to", "original@example.com"),
				),
			},
			{
				Config: `
resource "spork_alert_channel" "test" {
  name = "Updated Channel"
  type = "email"
  config = {
    to = "updated@example.com"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("spork_alert_channel.test", "name", "Updated Channel"),
					resource.TestCheckResourceAttr("spork_alert_channel.test", "config.to", "updated@example.com"),
				),
			},
		},
	})
}
