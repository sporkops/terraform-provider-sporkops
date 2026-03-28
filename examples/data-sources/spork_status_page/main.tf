data "spork_status_page" "main" {
  name = "Acme Status"
}

output "status_page_url" {
  value = "https://${data.spork_status_page.main.slug}.status.sporkops.com"
}
