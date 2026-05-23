resource "sendgrid_inbound_parse_rule" "agent" {
  hostname = sendgrid_authenticated_domain.agent.domain
  url      = "https://api.example.com/sendgrid/secret-token/ingest"

  depends_on = [sendgrid_domain_validation.agent]
}
