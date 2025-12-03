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

func TestDNSCreateUpdateWithoutRecreate(t *testing.T) {
	testDomain := os.Getenv("ALLINKL_TEST_DOMAIN")
	if testDomain == "" {
		t.Skip("ALLINKL_TEST_DOMAIN environment variable must be set")
	}

	currentSecondsAndMS := fmt.Sprintf("%d", time.Now().UnixNano())

	resourcePath := "allinkl_dns.test"
	resourceConfigTemplate := `resource "allinkl_dns" "test" {
  zone_host   = "%s"
  record_type = "%s"
  record_name = "%s"
  record_data = "%s"
  record_aux  = %d
}`

	initialType := "MX"
	initialName := currentSecondsAndMS + "tf.test"
	initialData := "mail.example.com."
	initialAux := 10

	updatedType := "MX"
	updatedName := "updated." + initialName
	updatedData := "mail2.example.com."
	updatedAux := 20

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create initial DNS entry
			{
				Config: fmt.Sprintf(resourceConfigTemplate, testDomain, initialType, initialName, initialData, initialAux),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "zone_host", testDomain),
					resource.TestCheckResourceAttr(resourcePath, "record_type", initialType),
					resource.TestCheckResourceAttr(resourcePath, "record_name", initialName),
					resource.TestCheckResourceAttr(resourcePath, "record_data", initialData),
					resource.TestCheckResourceAttr(resourcePath, "record_aux", fmt.Sprintf("%d", initialAux)),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourcePath]
						if !ok {
							return fmt.Errorf("Not found: %s", resourcePath)
						}
						id := rs.Primary.Attributes["record_id"]
						matched, err := regexp.MatchString(`^\d+$`, id)
						if err != nil {
							return err
						}
						if !matched {
							return fmt.Errorf("record_id does not match expected pattern: got %q", id)
						}
						return nil
					},
				),
			},
			// update existing entry without replacement
			{
				Config: fmt.Sprintf(resourceConfigTemplate, testDomain, updatedType, initialName, updatedData, updatedAux),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourcePath, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "zone_host", testDomain),
					resource.TestCheckResourceAttr(resourcePath, "record_type", updatedType),
					resource.TestCheckResourceAttr(resourcePath, "record_name", initialName),
					resource.TestCheckResourceAttr(resourcePath, "record_data", updatedData),
					resource.TestCheckResourceAttr(resourcePath, "record_aux", fmt.Sprintf("%d", updatedAux)),
				),
			},
			// update property that forces replacement
			{
				Config: fmt.Sprintf(resourceConfigTemplate, testDomain, updatedType, updatedName, updatedData, updatedAux),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourcePath, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "record_name", updatedName),
				),
			},
		},
	})
}

func TestDNSCreateFailsWithIllegalAux(t *testing.T) {
	testDomain := os.Getenv("ALLINKL_TEST_DOMAIN")
	if testDomain == "" {
		t.Skip("ALLINKL_TEST_DOMAIN environment variable must be set")
	}

	currentSecondsAndMS := fmt.Sprintf("%d", time.Now().UnixNano())

	resourceConfigTemplate := `resource "allinkl_dns" "test" {
  zone_host   = "%s"
  record_type = "%s"
  record_name = "%s"
  record_data = "%s"
  record_aux  = %d
}`

	initialType := "MX"
	initialName := currentSecondsAndMS + "tf.test"
	initialData := "mail.example.com."
	initialAux := -100

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(resourceConfigTemplate, testDomain, initialType, initialName, initialData, initialAux),
				ExpectError: regexp.MustCompile(`record_aux_syntax_incorrect`),
			},
		},
	})
}

func TestDNSCreateFailsForTypeWithNoAuxSupport(t *testing.T) {
	testDomain := os.Getenv("ALLINKL_TEST_DOMAIN")
	if testDomain == "" {
		t.Skip("ALLINKL_TEST_DOMAIN environment variable must be set")
	}

	currentSecondsAndMS := fmt.Sprintf("%d", time.Now().UnixNano())

	resourceConfigTemplate := `resource "allinkl_dns" "test" {
  zone_host   = "%s"
  record_type = "%s"
  record_name = "%s"
  record_data = "%s"
  record_aux  = %d
}`

	initialType := "A"
	initialName := currentSecondsAndMS + "tf.test"
	initialData := "203.0.113.10"
	initialAux := 10

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(resourceConfigTemplate, testDomain, initialType, initialName, initialData, initialAux),
				ExpectError: regexp.MustCompile(`record_aux must be empty or 0`),
			},
		},
	})
}
