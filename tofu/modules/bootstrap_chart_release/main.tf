locals {
  namespace   = coalesce(var.namespace, var.name)
  chart_path  = coalesce(var.chart_path, abspath("${path.module}/../../../bootstrap/${var.name}"))
  values_file = "${local.chart_path}/values.yaml"
}

resource "helm_release" "this" {
  name             = var.name
  namespace        = local.namespace
  chart            = local.chart_path
  create_namespace = var.create_namespace
  take_ownership   = true

  values = [file(local.values_file)]
}
