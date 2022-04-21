# Run "make init" to perform "terraform init"
# The easiest way to get a PostgreSQL is: docker run -e POSTGRES_PASSWORD="fill-me-in" -p 5432:5432 -t postgres

terraform {
  required_providers {
    csbpg = {
      source  = "cloudfoundry.org/cloud-service-broker/csbpg"
      version = "1.0.0"
    }
  }
}

provider "csbpg" {
  host            = "localhost"
  port            = 5432
  username        = "postgres"
  password        = "fill-me-in"
  database        = "postgres"
  data_owner_role = "dataowner"
}

resource "csbpg_binding_user" "binding_user" {
  username = "foo"
  password = "bar"
}
