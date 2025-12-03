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

func TestValidateDNSRecord(t *testing.T) {
	tests := []struct {
		name       string
		recordType string
		recordAux  int64
		want       bool
	}{
		// Test zero aux values are accepted for all record types
		{"zero aux with A record", "A", 0, true},
		{"zero aux with AAAA record", "AAAA", 0, true},
		{"zero aux with CNAME record", "CNAME", 0, true},
		{"zero aux with TXT record", "TXT", 0, true},
		{"zero aux with MX record", "MX", 0, true},
		{"zero aux with SRV record", "SRV", 0, true},

		// Test uppercase MX and SRV allow non-zero aux
		{"non-zero aux with MX record (uppercase)", "MX", 10, true},
		{"non-zero aux with SRV record (uppercase)", "SRV", 10, true},

		// Test lowercase mx and srv allow non-zero aux
		{"non-zero aux with mx record (lowercase)", "mx", 10, true},
		{"non-zero aux with srv record (lowercase)", "srv", 10, true},

		// Test non-zero aux is rejected for record types other than MX and SRV
		{"non-zero aux with A record", "A", 10, false},
		{"non-zero aux with AAAA record", "AAAA", 10, false},
		{"non-zero aux with CNAME record", "CNAME", 10, false},
		{"non-zero aux with TXT record", "TXT", 10, false},
		{"non-zero aux with NS record", "NS", 10, false},

		// Test lowercase non-MX/SRV types also reject non-zero aux
		{"non-zero aux with a record (lowercase)", "a", 10, false},
		{"non-zero aux with aaaa record (lowercase)", "aaaa", 10, false},
		{"non-zero aux with cname record (lowercase)", "cname", 10, false},
		{"non-zero aux with txt record (lowercase)", "txt", 10, false},

		// Test negative aux values
		{"negative aux with MX record", "MX", -10, true},
		{"negative aux with A record", "A", -10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateDNSRecord(tt.recordType, tt.recordAux)
			if got != tt.want {
				t.Errorf("ValidateDNSRecord(%q, %d) = %v, want %v", tt.recordType, tt.recordAux, got, tt.want)
			}
		})
	}
}

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
