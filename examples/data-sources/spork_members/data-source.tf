data "spork_members" "all" {}

output "member_emails" {
  value = [for m in data.spork_members.all.members : m.email]
}
