terraform {
  required_providers {
    influxdbv2 = {
      source = "registry.terraform.io/psenna/influxdbv2"
    }
  }
}

provider "influxdbv2" {
    host = "http://influxdb:8086"
    api_key = "V75L9W05AABQBCACF6F8CVDJTPFEXA"
}

data "influxdbv2_organization" "old_organization" {
  name = "firstorg"
}

resource "influxdbv2_organization" "new_organization" {
  name = "neworg"
  description = "New created org"
}

resource "influxdbv2_bucket" "new_bucket" {
  name = "new_bucket"
  org_id = influxdbv2_organization.new_organization.id
  rp = "0"
  retention_rules = [{
    every_seconds = 2592000,
    retention_type = "expire"
  }]
}