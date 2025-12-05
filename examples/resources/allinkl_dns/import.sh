# dns entries are imported using their ID
# when editing a DNS record via the KAS, the browser URL will follow
# https://kas.all-inkl.com/index.php?...&id=$ID
# use the value of $ID to import the DNS record:
terraform import allinkl_dns.home 123456789