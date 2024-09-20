package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDDNSResource(t *testing.T) {
	testDomain := os.Getenv("ALLINKL_TEST_DOMAIN")
	if testDomain == "" {
		t.Fatal("ALLINKL_TEST_DOMAIN environment variable must be set")
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "allinkl_ddns" "test" {
  dyndns_comment   = "terraformprovider test"
  dyndns_password  = "password"
  dyndns_zone      = "%s"
  dyndns_label     = "terraformprovider.test"
  dyndns_target_ip = "1.2.3.4"
}
`, testDomain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("allinkl_ddns.test", "dyndns_comment", "terraformprovider test"),
					resource.TestCheckResourceAttr("allinkl_ddns.test", "dyndns_password", "password"),
					resource.TestCheckResourceAttr("allinkl_ddns.test", "dyndns_zone", testDomain),
					resource.TestCheckResourceAttr("allinkl_ddns.test", "dyndns_label", "terraformprovider.test"),
					resource.TestCheckResourceAttr("allinkl_ddns.test", "dyndns_target_ip", "1.2.3.4"),
					// Check that dyndns_login starts with "dyn" and ends with at least one digit
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["allinkl_ddns.test"]
						if !ok {
							return fmt.Errorf("Not found: allinkl_ddns.test")
						}
						login := rs.Primary.Attributes["dyndns_login"]
						matched, err := regexp.MatchString(`^dyn[a-fA-F\d]+$`, login)
						if err != nil {
							return err
						}
						if !matched {
							return fmt.Errorf("dyndns_login does not match expected pattern: got %q", login)
						}
						return nil
					},
				),
			},
		},
	})
}
