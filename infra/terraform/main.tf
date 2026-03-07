resource "kubernetes_manifest" "namespace" {
  manifest = yamldecode(file("${path.module}/../kubernetes/namespace.yaml"))
}

resource "kubernetes_manifest" "service_account" {
  manifest = yamldecode(file("${path.module}/../kubernetes/service-account.yaml"))
}

resource "kubernetes_manifest" "service" {
  manifest = yamldecode(file("${path.module}/../kubernetes/service.yaml"))
}

resource "kubernetes_manifest" "deployment" {
  manifest = yamldecode(file("${path.module}/../kubernetes/deployment.yaml"))
}

resource "kubernetes_manifest" "http_route" {
  manifest = yamldecode(file("${path.module}/../kubernetes/http-route.yaml"))
}

resource "kubernetes_manifest" "configmap" {
  manifest = yamldecode(file("${path.module}/../kubernetes/configmap.yaml"))
}

resource "kubernetes_manifest" "hpa" {
  manifest = yamldecode(file("${path.module}/../kubernetes/hpa.yaml"))
}

# Secret com credenciais AWS (criado após o namespace)
resource "kubernetes_secret" "secret" {
  metadata {
    name      = "processor-secret"
    namespace = "processor"
  }

  data = {
    AWS_REGION            = var.aws_region
    AWS_ACCESS_KEY_ID     = var.aws_access_key
    AWS_SECRET_ACCESS_KEY = var.aws_secret_key
  }

  type = "Opaque"

  # Garante que o namespace existe antes de criar o secret
  depends_on = [kubernetes_manifest.namespace]
}
