terraform {
  required_providers {
    allinkl = {
      source = "vimaster/allinkl"
    }
  }
}

provider "allinkl" {
}

resource "allinkl_dns" "sub_test" {
    record_name = "sub"
    record_type = "A"
    record_data = "1.2.3.4"
    record_aux = "0"
    zone_host = "vimaster.de"
}