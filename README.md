# All-Inkl provider

A Terraform provider for [all-inkl.com](https://all-inkl.com/)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.12.0
- [Go](https://golang.org/doc/install) >= 1.24.5

## Examples

### Creating a DDNS entry for an IPv4

```tf
provider "allinkl" {
  kas_auth_type = "plain"
  kas_login     = "w123456"  # or set ALLINKL_KAS_LOGIN environment variable
  kas_auth_data = "PASSWORD" # or set ALLINKL_KAS_AUTH_DATA environment variable
}

resource "allinkl_ddns" "home" {
  dyndns_comment   = "home"
  dyndns_password  = "password"
  dyndns_zone      = "%s"
  dyndns_label     = "home.mydomain.tld"
  dyndns_target_ip = "1.2.3.4"
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

### Tests

> [!WARNING]
> Running acceptance tests will make changes to your All-Inkl account. Use a test account or proceed with caution.

1. Set `ALLINKL_KAS_LOGIN` to your KAS username (starts with `w`)
2. Set `ALLINKL_KAS_AUTH_DATA` to your KAS password
2. Set `ALLINKL_TEST_DOMAIN` to a domain in your account, that tests will use to generate DynDNS entries
3. Run `make testacc`

### Documentation

To generate or update documentation, run `go generate`.

### Publishing the Provider

1. Fork this repository
2. Push to a tag respecting [Semantic Versioning](https://semver.org/), prefixed with `v` (ex. `v0.1.0, v1.0.0`)
3. Wait for GitHub Actions to complete