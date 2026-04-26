variable "kubeconfig_path" {
  type        = string
  description = "Path to kubeconfig used by all providers."
  default     = "/Users/ant0n/.kube/config"
}

variable "kubeconfig_context" {
  type        = string
  description = "Optional kubeconfig context used by all providers."
  default     = "homelab"
}

variable "onepassword_vault" {
  type        = string
  description = "1Password vault containing all secret items."
  default     = "Kubernetes"
}

variable "onepassword_credentials_item" {
  type        = string
  description = "1Password item title for the 1Password Connect credentials JSON. (copy content from file into regular field)"
  default     = "Kubernetes Credentials File"
}

variable "onepassword_token_item" {
  type        = string
  description = "1Password item title for the 1Password Kubernetes access token."
  default     = "Kubernetes Access Token: Kubernetes"
}
