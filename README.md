# terraform-provider-sendgrid

A Terraform provider for managing a focused slice of [SendGrid](https://sendgrid.com/) — Authenticated Domains, their DNS validation, and Inbound Parse webhook rules. It is **not** a general-purpose SendGrid provider; it covers a specific set of resources.

Published on the [Terraform Registry](https://registry.terraform.io/providers/weesp-ai/sendgrid). To use it, declare it in your Terraform configuration:

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

Set the value through Terraform's variable mechanism — for example, `export TF_VAR_sendgrid_api_key=SG.xxx` before running `terraform plan`. The provider deliberately does **not** read API keys from arbitrary environment variables.

## Resources

| Resource                          | Purpose                                                                       |
|-----------------------------------|-------------------------------------------------------------------------------|
| `sendgrid_authenticated_domain`   | A SendGrid Authenticated Domain (whitelabel) with DKIM/CNAME DNS attributes.  |
| `sendgrid_domain_validation`      | Polls SendGrid until a domain passes DNS validation. Use after DNS records.   |
| `sendgrid_inbound_parse_rule`     | An Inbound Parse webhook rule that forwards mail for a hostname to a URL.     |

See [`docs/`](./docs) for full attribute reference and [`examples/`](./examples) for usage patterns.

## Development

```bash
go build ./...                                       # build
go test ./...                                        # unit tests
TF_ACC=1 TF_VAR_sendgrid_api_key=sg... go test ./... # acceptance tests (live API)
```

To test changes against a Terraform project before tagging a release, use a [filesystem dev override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers):

```hcl
# ~/.terraformrc
provider_installation {
  dev_overrides {
    "weesp-ai/sendgrid" = "/path/to/terraform-provider-sendgrid"
  }
  direct {}
}
```

Then `go install ./...` and run `terraform plan` in your consuming project — Terraform will load the locally-built binary instead of fetching from the registry.

## Releasing

Pushing a `vX.Y.Z` tag triggers `.github/workflows/release.yml`, which uses [GoReleaser](https://goreleaser.com/) to:

1. Build provider binaries for the supported platforms.
2. Produce a `SHA256SUMS` checksums file and a detached GPG signature (`SHA256SUMS.sig`).
3. Create a GitHub Release containing the archives, checksums, signature, and the registry manifest.

The Terraform Registry detects the new GitHub Release via webhook and publishes the version; consumers pick it up on their next `terraform init -upgrade`.

Prerequisites (one-time setup): the repository must be public, the signing key's **public** half must be uploaded to the `weesp-ai` namespace on the registry, and the **private** key plus its passphrase must be stored as the `GPG_PRIVATE_KEY` and `PASSPHRASE` GitHub Actions secrets. See [HashiCorp's publishing guide](https://developer.hashicorp.com/terraform/registry/providers/publishing) for the full walkthrough.
