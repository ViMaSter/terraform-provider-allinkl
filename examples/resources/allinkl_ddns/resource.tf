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