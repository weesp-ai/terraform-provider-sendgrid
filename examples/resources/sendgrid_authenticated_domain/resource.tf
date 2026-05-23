resource "sendgrid_authenticated_domain" "agent" {
  domain = "agent.example.com"
}

# Pair the SendGrid-supplied DNS attributes with your DNS provider — Cloud DNS shown.
resource "google_dns_record_set" "agent_mail" {
  managed_zone = "example-com"
  name         = "${sendgrid_authenticated_domain.agent.mail_cname_host}."
  type         = "CNAME"
  ttl          = 300
  rrdatas      = ["${sendgrid_authenticated_domain.agent.mail_cname_data}."]
}
