resource "allinkl_dns" "mx_subdomain" {
  zone_host   = "example.com."
  record_type = "MX"
  record_name = "subdomain"
  record_data = "subdomain.example.com." # needs to end with ., if specifying a CNAME
  record_aux  = 10                       # for MX and SRV records only
}
  
# A record example
resource "allinkl_dns" "a_subdomain" {
  zone_host   = "example.com."
  record_type = "A"
  record_name = "subdomain"
  record_data = "1.2.3.4"
}