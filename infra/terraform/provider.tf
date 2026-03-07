provider "aws" {
  region = "sa-east-1"
}

data "aws_eks_cluster" "hackaton" {
  name = "hackaton-soat-terraform"
}

data "aws_eks_cluster_auth" "hackaton" {
  name = "hackaton-soat-terraform"
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.hackaton.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.hackaton.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.hackaton.token
}

provider "kubectl" {
  host                   = data.aws_eks_cluster.hackaton.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.hackaton.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.hackaton.token
  load_config_file       = false
}
