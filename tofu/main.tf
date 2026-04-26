data "onepassword_item" "onepassword_credentials" {
  vault = var.onepassword_vault
  title = var.onepassword_credentials_item
}

data "onepassword_item" "onepassword_token" {
  vault = var.onepassword_vault
  title = var.onepassword_token_item
}

module "cilium" {
  source = "./modules/bootstrap_chart_release"

  name = "cilium"
}

module "argocd" {
  source = "./modules/bootstrap_chart_release"

  name = "argocd"

  depends_on = [
    module.cilium,
  ]
}

module "root" {
  source = "./modules/bootstrap_chart_release"

  name             = "root"
  namespace        = module.argocd.namespace
  chart_path       = "${path.module}/../bootstrap/root"
  create_namespace = false

  depends_on = [
    module.argocd,
  ]
}

resource "kubernetes_namespace_v1" "onepassword" {
  metadata {
    name = "onepassword"
  }
}

resource "kubernetes_secret_v1" "onepassword" {
  metadata {
    name      = "onepassword-secret"
    namespace = kubernetes_namespace_v1.onepassword.metadata[0].name
  }

  type = "Opaque"

  data = {
    credentials = data.onepassword_item.onepassword_credentials.section[0].field[0].value
    token       = data.onepassword_item.onepassword_token.credential
  }

  depends_on = [
    kubernetes_namespace_v1.onepassword,
  ]
}

resource "terraform_data" "remove_helm_secrets" {
  triggers_replace = timestamp()

  provisioner "local-exec" {
    command = "kubectl delete secret -A -l owner=helm --ignore-not-found"
  }

  depends_on = [
    module.cilium,
    module.argocd,
    module.root,
  ]
}
