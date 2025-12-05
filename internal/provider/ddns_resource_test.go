package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestDDNSCreateUpdateWithoutRecreate(t *testing.T) {
	testDomain := os.Getenv("ALLINKL_TEST_DOMAIN")
	if testDomain == "" {
		t.Skip("ALLINKL_TEST_DOMAIN environment variable must be set")
	}

	currentSecondsAndMS := fmt.Sprintf("%d", time.Now().UnixNano())

	resourcePath := "allinkl_ddns.test"
	resourceConfigTemplate := `resource "allinkl_ddns" "test" {
  dyndns_comment   = "%s"
  dyndns_password  = "%s"
  dyndns_zone      = "%s"
  dyndns_label     = "%s"
  dyndns_target_ip = "%s"
}`

	initialComment := currentSecondsAndMS + "tftest"
	initialPassword := "password"
	initialLabel := currentSecondsAndMS + "tf.test"
	initialTargetIP := "1.2.3.4"

	updatedComment := initialComment + "_updated"
	updatedPassword := initialPassword + "_updated"
	updatedLabel := "updated." + initialLabel

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create initial DynDNS entry
			{
				Config: fmt.Sprintf(resourceConfigTemplate, initialComment, initialPassword, testDomain, initialLabel, initialTargetIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "dyndns_comment", initialComment),
					resource.TestCheckResourceAttr(resourcePath, "dyndns_password", initialPassword),
					resource.TestCheckResourceAttr(resourcePath, "dyndns_zone", testDomain),
					resource.TestCheckResourceAttr(resourcePath, "dyndns_label", initialLabel),
					resource.TestCheckResourceAttr(resourcePath, "dyndns_target_ip", initialTargetIP),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourcePath]
						if !ok {
							return fmt.Errorf("Not found: %s", resourcePath)
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
			// create update existing entry without replacement
			{
				Config: fmt.Sprintf(resourceConfigTemplate, updatedComment, updatedPassword, testDomain, initialLabel, initialTargetIP),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourcePath, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "dyndns_comment", updatedComment),
					resource.TestCheckResourceAttr(resourcePath, "dyndns_password", updatedPassword),
					resource.TestCheckResourceAttr(resourcePath, "dyndns_zone", testDomain),
					resource.TestCheckResourceAttr(resourcePath, "dyndns_label", initialLabel),
					resource.TestCheckResourceAttr(resourcePath, "dyndns_target_ip", initialTargetIP),
				),
			},
			// update property that forces replacement
			{
				Config: fmt.Sprintf(resourceConfigTemplate, updatedComment, updatedPassword, testDomain, updatedLabel, initialTargetIP),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourcePath, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "dyndns_label", updatedLabel),
				),
			},
		},
	})
}

func TestDDNSCreateFailsWithIllegalCharactersInPassword(t *testing.T) {
	testDomain := os.Getenv("ALLINKL_TEST_DOMAIN")
	if testDomain == "" {
		t.Skip("ALLINKL_TEST_DOMAIN environment variable must be set")
	}

	currentSecondsAndMS := fmt.Sprintf("%d", time.Now().UnixNano())

	resourceConfigTemplate := `resource "allinkl_ddns" "test" {
  dyndns_comment   = "%s"
  dyndns_password  = "%s"
  dyndns_zone      = "%s"
  dyndns_label     = "%s"
  dyndns_target_ip = "%s"
}`

	initialComment := currentSecondsAndMS + "tftest"
	initialPassword := "illegal password"
	initialLabel := currentSecondsAndMS + "tf.test"
	initialTargetIP := "1.2.3.4"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(resourceConfigTemplate, initialComment, initialPassword, testDomain, initialLabel, initialTargetIP),
				ExpectError: regexp.MustCompile(`password_syntax_incorrect: `),
			},
		},
	})
}
