//go:build acceptance

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStatusPageResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_status_page" "test" {
  name = "Test Status Page"
  slug = "tf-test-basic"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("spork_status_page.test", "id"),
					resource.TestCheckResourceAttr("spork_status_page.test", "name", "Test Status Page"),
					resource.TestCheckResourceAttr("spork_status_page.test", "slug", "tf-test-basic"),
					resource.TestCheckResourceAttr("spork_status_page.test", "theme", "light"),
					resource.TestCheckResourceAttr("spork_status_page.test", "is_public", "true"),
					resource.TestCheckResourceAttrSet("spork_status_page.test", "created_at"),
				),
			},
			{
				ResourceName:      "spork_status_page.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStatusPageResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_status_page" "test" {
  name  = "Test Status Page"
  slug  = "tf-test-update"
  theme = "light"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("spork_status_page.test", "name", "Test Status Page"),
					resource.TestCheckResourceAttr("spork_status_page.test", "theme", "light"),
				),
			},
			{
				Config: `
resource "spork_status_page" "test" {
  name         = "Updated Status Page"
  slug         = "tf-test-update"
  theme        = "dark"
  accent_color = "#0066ff"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("spork_status_page.test", "name", "Updated Status Page"),
					resource.TestCheckResourceAttr("spork_status_page.test", "theme", "dark"),
					resource.TestCheckResourceAttr("spork_status_page.test", "accent_color", "#0066ff"),
				),
			},
		},
	})
}

func TestAccStatusPageResource_withComponents(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_monitor" "test" {
  target = "https://httpbin.org/get"
  name   = "Test Monitor for Status Page"
  type   = "http"
}

resource "spork_status_page" "test" {
  name = "Test Status Page with Components"
  slug = "tf-test-components"

  components {
    monitor_id   = spork_monitor.test.id
    display_name = "API"
    order        = 0
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("spork_status_page.test", "id"),
					resource.TestCheckResourceAttr("spork_status_page.test", "components.#", "1"),
					resource.TestCheckResourceAttr("spork_status_page.test", "components.0.display_name", "API"),
					resource.TestCheckResourceAttr("spork_status_page.test", "components.0.order", "0"),
				),
			},
		},
	})
}
