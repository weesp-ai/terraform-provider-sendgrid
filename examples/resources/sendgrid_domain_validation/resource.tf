resource "sendgrid_domain_validation" "agent" {
  domain_id = sendgrid_authenticated_domain.agent.domain_id

  # Block validation until the DNS records SendGrid expects are live.
  depends_on = [
    google_dns_record_set.agent_mail,
    google_dns_record_set.agent_dkim1,
    google_dns_record_set.agent_dkim2,
  ]
}
