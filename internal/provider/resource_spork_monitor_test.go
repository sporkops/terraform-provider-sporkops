//go:build acceptance

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"spork": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("SPORK_API_KEY"); v == "" {
		t.Skip("SPORK_API_KEY must be set for acceptance tests")
	}
}

func TestAccMonitorResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_monitor" "test" {
  target = "https://httpbin.org/get"
  name   = "Test Monitor"
  type   = "http"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("spork_monitor.test", "id"),
					resource.TestCheckResourceAttr("spork_monitor.test", "target", "https://httpbin.org/get"),
					resource.TestCheckResourceAttr("spork_monitor.test", "name", "Test Monitor"),
					resource.TestCheckResourceAttr("spork_monitor.test", "type", "http"),
					resource.TestCheckResourceAttr("spork_monitor.test", "method", "GET"),
					resource.TestCheckResourceAttr("spork_monitor.test", "expected_status", "200"),
					resource.TestCheckResourceAttr("spork_monitor.test", "interval", "60"),
					resource.TestCheckResourceAttr("spork_monitor.test", "paused", "false"),
					resource.TestCheckResourceAttrSet("spork_monitor.test", "status"),
				),
			},
			{
				ResourceName:      "spork_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMonitorResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_monitor" "test" {
  target   = "https://httpbin.org/get"
  name     = "Test Monitor"
  type     = "http"
  interval = 60
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("spork_monitor.test", "name", "Test Monitor"),
					resource.TestCheckResourceAttr("spork_monitor.test", "interval", "60"),
				),
			},
			{
				Config: `
resource "spork_monitor" "test" {
  target   = "https://httpbin.org/get"
  name     = "Updated Monitor"
  type     = "http"
  interval = 300
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("spork_monitor.test", "name", "Updated Monitor"),
					resource.TestCheckResourceAttr("spork_monitor.test", "interval", "300"),
				),
			},
		},
	})
}

func TestAccMonitorResource_allAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_monitor" "test" {
  target          = "https://httpbin.org/status/201"
  name            = "Full Config Monitor"
  type            = "http"
  method          = "POST"
  expected_status = 201
  interval        = 300
  paused          = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("spork_monitor.test", "target", "https://httpbin.org/status/201"),
					resource.TestCheckResourceAttr("spork_monitor.test", "name", "Full Config Monitor"),
					resource.TestCheckResourceAttr("spork_monitor.test", "type", "http"),
					resource.TestCheckResourceAttr("spork_monitor.test", "method", "POST"),
					resource.TestCheckResourceAttr("spork_monitor.test", "expected_status", "201"),
					resource.TestCheckResourceAttr("spork_monitor.test", "interval", "300"),
					resource.TestCheckResourceAttr("spork_monitor.test", "paused", "true"),
				),
			},
		},
	})
}
