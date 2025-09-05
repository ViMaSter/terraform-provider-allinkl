terraform {
  required_providers {
    allinkl = {
      source  = "ViMaSter/allinkl"
      version = "0.1.0"
    }
  }
}

provider "allinkl" {
  kas_auth_type = "plain"
  kas_login     = "w123456"  # or set ALLINKL_KAS_LOGIN environment variable
  kas_auth_data = "PASSWORD" # or set ALLINKL_KAS_AUTH_DATA environment variable
}

# based on all-inkl's requirements
resource "random_password" "ddns_password" {
  length           = 30
  min_upper        = 1
  min_lower        = 1
  min_numeric      = 1
  min_special      = 1
  special          = true
  override_special = "/-_#*+!§,()=:.@äöüÄÖÜß"
}

resource "allinkl_ddns" "home" {
  dyndns_comment   = "home"
  dyndns_password  = random_password.ddns_password.result
  dyndns_zone      = "my.tld"
  dyndns_label     = "subdomain"
  dyndns_target_ip = "1.2.3.4"
}