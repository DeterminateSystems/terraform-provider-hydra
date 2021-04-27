package hydra

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProvider *schema.Provider
var testAccProviders map[string]*schema.Provider

func init() {
	testAccProvider = Provider()

	testAccProviders = map[string]*schema.Provider{
		"hydra": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("HYDRA_HOST"); v == "" {
		t.Fatal("HYDRA_HOST must be set for acceptance tests\n",
			"NOTE: This should be a throwaway Hydra instance")
	}
	if v := os.Getenv("HYDRA_USERNAME"); v == "" {
		t.Fatal("HYDRA_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("HYDRA_PASSWORD"); v == "" {
		t.Fatal("HYDRA_PASSWORD must be set for acceptance tests")
	}
}
