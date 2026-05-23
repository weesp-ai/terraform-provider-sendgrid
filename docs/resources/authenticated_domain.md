---
page_title: "sendgrid_authenticated_domain Resource"
description: "A SendGrid Authenticated Domain (whitelabel) with DKIM/CNAME DNS attributes."
---

# sendgrid_authenticated_domain (Resource)

A SendGrid Authenticated Domain (whitelabel) with DKIM/CNAME DNS attributes. SendGrid domains are immutable; changes to any input force a replacement. The provider adopts an existing domain with the same hostname instead of failing on conflict — this makes the resource safe to add to state for already-provisioned domains without an explicit `terraform import`.

## Example Usage

```hcl
resource "sendgrid_authenticated_domain" "agent" {
  domain = "agent.example.com"
}

resource "google_dns_record_set" "agent_mail" {
  managed_zone = "example-com"
  name         = "${sendgrid_authenticated_domain.agent.mail_cname_host}."
  type         = "CNAME"
  ttl          = 300
  rrdatas      = ["${sendgrid_authenticated_domain.agent.mail_cname_data}."]
}
```

## Schema

### Required

- `domain` (String) Fully-qualified hostname to authenticate (e.g. `mail.example.com`).

### Optional

- `automatic_security` (Boolean) Whether SendGrid manages the DKIM keys and CNAMEs. Defaults to `true`.
- `custom_spf` (Boolean) Whether SPF records are managed by the user. Defaults to `false`.
- `is_default` (Boolean) Whether this is the SendGrid account's default domain. Defaults to `false`.

### Read-Only

- `id` (String) The SendGrid domain ID, as a string.
- `domain_id` (Number) The SendGrid domain ID as an integer.
- `valid` (Boolean) Whether SendGrid considers this domain DNS-validated.
- `mail_cname_host` (String) DNS host for the SendGrid `em.*` mail CNAME.
- `mail_cname_data` (String) DNS target for the mail CNAME.
- `dkim1_host` (String) DNS host for the first DKIM CNAME.
- `dkim1_data` (String) DNS target for the first DKIM CNAME.
- `dkim2_host` (String) DNS host for the second DKIM CNAME.
- `dkim2_data` (String) DNS target for the second DKIM CNAME.

## Import

```shell
# Import by numeric SendGrid domain ID.
terraform import sendgrid_authenticated_domain.agent 42
```
