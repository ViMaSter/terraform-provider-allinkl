package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDDNSResource(t *testing.T) {
	testDomain := os.Getenv("ALLINKL_TEST_DOMAIN")
	if testDomain == "" {
		t.Fatal("ALLINKL_TEST_DOMAIN environment variable must be set")
	}

	now := time.Now()
	currentSecondsAndMS := fmt.Sprintf("%02d%03d", now.Unix()%100, now.Nanosecond()/1e6)

	testComment := currentSecondsAndMS + "tftest"
	testPassword := "password"
	testLabel := currentSecondsAndMS + "tf.test"
	testTargetIP := "1.2.3.4"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "allinkl_ddns" "test" {
  dyndns_comment   = "%s"
  dyndns_password  = "%s"
  dyndns_zone      = "%s"
  dyndns_label     = "%s"
  dyndns_target_ip = "%s"
}
`, testComment, testPassword, testDomain, testLabel, testTargetIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("allinkl_ddns.test", "dyndns_comment", testComment),
					resource.TestCheckResourceAttr("allinkl_ddns.test", "dyndns_password", testPassword),
					resource.TestCheckResourceAttr("allinkl_ddns.test", "dyndns_zone", testDomain),
					resource.TestCheckResourceAttr("allinkl_ddns.test", "dyndns_label", testLabel),
					resource.TestCheckResourceAttr("allinkl_ddns.test", "dyndns_target_ip", testTargetIP),
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
