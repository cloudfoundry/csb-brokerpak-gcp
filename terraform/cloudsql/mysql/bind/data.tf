locals {
  port          = 3306
  useSSL        = !var.allow_insecure_connections
  jdbcUrlSuffix = var.allow_insecure_connections ? "" : "\u0026disableSslHostnameVerification=true"
}
