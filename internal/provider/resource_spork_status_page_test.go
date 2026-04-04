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

func TestAccStatusPageResource_withComponentGroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_monitor" "api" {
  target = "https://httpbin.org/get"
  name   = "TF Test API Monitor"
  type   = "http"
}

resource "spork_monitor" "web" {
  target = "https://httpbin.org/html"
  name   = "TF Test Web Monitor"
  type   = "http"
}

resource "spork_monitor" "db" {
  target = "https://httpbin.org/status/200"
  name   = "TF Test DB Monitor"
  type   = "http"
}

resource "spork_status_page" "test" {
  name = "TF Test Grouped Page"
  slug = "tf-test-groups"

  component_groups {
    name  = "Website"
    order = 0
  }

  component_groups {
    name  = "Backend"
    order = 1
  }

  components {
    monitor_id   = spork_monitor.api.id
    display_name = "API"
    group        = "Backend"
    order        = 0
  }

  components {
    monitor_id   = spork_monitor.web.id
    display_name = "Homepage"
    group        = "Website"
    order        = 0
  }

  components {
    monitor_id   = spork_monitor.db.id
    display_name = "Database"
    group        = "Backend"
    order        = 1
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("spork_status_page.test", "component_groups.#", "2"),
					resource.TestCheckResourceAttr("spork_status_page.test", "components.#", "3"),
					resource.TestCheckResourceAttr("spork_status_page.test", "components.0.group", "Backend"),
					resource.TestCheckResourceAttr("spork_status_page.test", "components.1.group", "Website"),
					resource.TestCheckResourceAttr("spork_status_page.test", "components.2.group", "Backend"),
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

// Data source tests

func TestAccStatusPageDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_status_page" "test" {
  name = "DS Test Page"
  slug = "tf-test-ds-id"
}

data "spork_status_page" "test" {
  id = spork_status_page.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.spork_status_page.test", "name", "DS Test Page"),
					resource.TestCheckResourceAttr("data.spork_status_page.test", "slug", "tf-test-ds-id"),
					resource.TestCheckResourceAttr("data.spork_status_page.test", "theme", "light"),
					resource.TestCheckResourceAttr("data.spork_status_page.test", "is_public", "true"),
				),
			},
		},
	})
}

func TestAccStatusPageDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_status_page" "test" {
  name = "DS Test By Name"
  slug = "tf-test-ds-name"
}

data "spork_status_page" "test" {
  name = spork_status_page.test.name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.spork_status_page.test", "id"),
					resource.TestCheckResourceAttr("data.spork_status_page.test", "name", "DS Test By Name"),
					resource.TestCheckResourceAttr("data.spork_status_page.test", "slug", "tf-test-ds-name"),
				),
			},
		},
	})
}

func TestAccStatusPagesDataSource_list(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "spork_status_page" "test" {
  name = "DS Test List"
  slug = "tf-test-ds-list"
}

data "spork_status_pages" "all" {
  depends_on = [spork_status_page.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.spork_status_pages.all", "status_pages.#"),
				),
			},
		},
	})
}
