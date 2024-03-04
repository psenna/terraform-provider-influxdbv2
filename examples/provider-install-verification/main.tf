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
  name = "neworg123"
  description = "New created org"
}