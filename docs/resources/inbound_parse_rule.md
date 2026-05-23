---
page_title: "sendgrid_inbound_parse_rule Resource"
description: "A SendGrid Inbound Parse webhook rule that forwards mail addressed to a hostname to an HTTPS URL."
---

# sendgrid_inbound_parse_rule (Resource)

A SendGrid Inbound Parse webhook rule that forwards mail addressed to a hostname to an HTTPS URL. SendGrid permits at most one rule per hostname; the provider adopts an existing rule with the same hostname instead of failing — making the resource safe to add to state for already-provisioned hostnames without an explicit `terraform import`.

## Example Usage

```hcl
resource "sendgrid_inbound_parse_rule" "agent" {
  hostname = sendgrid_authenticated_domain.agent.domain
  url      = "https://api.example.com/sendgrid/secret-token/ingest"

  depends_on = [sendgrid_domain_validation.agent]
}
```

## Schema

### Required

- `hostname` (String) Mail-flow hostname (e.g. `mail.example.com`). Changing this forces a new resource.
- `url` (String) HTTPS URL SendGrid POSTs the inbound mail to.

### Optional

- `spam_check` (Boolean) Whether SendGrid runs SpamAssassin on inbound mail before forwarding. Defaults to `false`.
- `send_raw` (Boolean) Whether SendGrid forwards the original RFC 822 message as `email`. Defaults to `true`.

### Read-Only

- `id` (String) Same as `hostname`.

## Import

```shell
# Import by hostname.
terraform import sendgrid_inbound_parse_rule.agent agent.example.com
```
