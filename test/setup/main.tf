terraform {
  required_providers {
    allinkl = {
      source = "vimaster/allinkl"
    }
  }
}

provider "allinkl" {
  kas_login = "w0117660"
  kas_auth_type = "plain"
  kas_auth_data = "89Pc%C3%B66.lEN%C3%84KfykBLHWv%C3%A4u8p_F0zma"
}

resource "allinkl_ddns" "ddns_test" {
    dyndns_comment = "terraformcomment"
    dyndns_password = "terraformpassword"
    dyndns_zone = "mahn.ke"
    dyndns_label = "terraformzone.by.vincent"
    dyndns_target_ip = "1.2.3.4"
}

output "ddns_login" {
    value = allinkl_ddns.ddns_test.login
}
