variable "cloudflare_email" {
  type = string
}

variable "cloudflare_api_key" {
  type      = string
  sensitive = true
}

variable "cloudflare_account_id" {
  type = string
}

variable "onepassword_credentials" {
  type = string
}

variable "onepassword_token" {
  type = string
}
