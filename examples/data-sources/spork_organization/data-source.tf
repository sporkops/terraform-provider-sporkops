data "spork_organization" "current" {}

output "org_name" {
  value = data.spork_organization.current.name
}

output "monitoring_plan" {
  value = data.spork_organization.current.subscriptions[0].plan
}
