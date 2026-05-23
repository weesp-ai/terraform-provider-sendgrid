---
page_title: "sendgrid_domain_validation Resource"
description: "Polls SendGrid until an Authenticated Domain passes DNS validation."
---

# sendgrid_domain_validation (Resource)

Polls SendGrid until an Authenticated Domain passes DNS validation. Place this resource downstream of the DNS records that satisfy the domain's CNAME requirements. The poll cadence is 10 seconds and times out after 5 minutes.

`Delete` is a no-op — validation is a logical resource, not a SendGrid object. `Update` is unreachable because `domain_id` forces a replacement.

## Example Usage

```hcl
resource "sendgrid_domain_validation" "agent" {
  domain_id = sendgrid_authenticated_domain.agent.domain_id

  depends_on = [
    google_dns_record_set.agent_mail,
    google_dns_record_set.agent_dkim1,
    google_dns_record_set.agent_dkim2,
  ]
}
```

## Schema

### Required

- `domain_id` (Number) ID of the SendGrid Authenticated Domain to validate.

### Read-Only

- `id` (String) Same as `domain_id`, stringified.
- `valid` (Boolean) True after SendGrid confirms the domain's DNS records.

## Import

```shell
# Import by numeric SendGrid domain ID.
terraform import sendgrid_domain_validation.agent 42
```
