terraform {
  required_providers {
    sendgrid = {
      source  = "weesp-ai/sendgrid"
      version = "~> 0.2"
    }
  }
}

variable "sendgrid_api_key" {
  description = "SendGrid API key. Set via TF_VAR_sendgrid_api_key in the environment."
  type        = string
  sensitive   = true
}

provider "sendgrid" {
  api_key = var.sendgrid_api_key
}
