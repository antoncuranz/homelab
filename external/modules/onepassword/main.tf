resource "kubernetes_secret" "onepassword_secret" {
  metadata {
    name      = "onepassword-secret"
    namespace = "onepassword"
  }

  data = {
    "credentials" = var.onepassword_credentials
    "token"       = var.onepassword_token
  }
}
