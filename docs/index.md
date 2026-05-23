---
page_title: "sendgrid Provider"
description: "Manages a focused slice of SendGrid: Authenticated Domains, their DNS validation, and Inbound Parse webhook rules."
---

# sendgrid Provider

Manages a focused slice of SendGrid (Authenticated Domains, their DNS validation, and Inbound Parse webhook rules). This is **not** a general-purpose SendGrid provider — it covers a specific set of resources.

## Example Usage

```hcl
terraform {
  required_providers {
    sendgrid = {
      source  = "weesp-ai/sendgrid"
      version = "~> 0.2"
    }
  }
}

variable "sendgrid_api_key" {
  type      = string
  sensitive = true
}

provider "sendgrid" {
  api_key = var.sendgrid_api_key
}
```

Set the value via `TF_VAR_sendgrid_api_key` in the calling environment (or any other Terraform variable mechanism) — the provider does not read API keys from arbitrary environment variables.

## Schema

### Required

- `api_key` (String, Sensitive) SendGrid API key. Required scopes: `mail.send`, `whitelabel`, `inbound_parse`.
