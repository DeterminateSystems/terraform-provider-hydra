package hydra

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"terraform-provider-hydra/hydra/api"
)

func TestAccHydraProject_basic(t *testing.T) {
	// identifier must start with a letter
	name := fmt.Sprintf("p%s", acctest.RandString(7))
	rename := fmt.Sprintf("%s-2", name)
	badname := "123"
	resourceName := "hydra_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHydraProjectDestroy,
		Steps: []resource.TestStep{
			// Test creation of project
			{
				Config: testAccHydraProjectConfigBasic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName),
				),
			},
			// Test modification of project name
			{
				Config: testAccHydraProjectConfigBasic(rename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName),
				),
			},
			// Test invalid project identifier
			{
				Config:      testAccHydraProjectConfigBasic(badname),
				ExpectError: regexp.MustCompile("Invalid project identifier"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName),
				),
			},
		},
	})
}

func TestAccHydraProject_hiddenDisabled(t *testing.T) {
	// identifier must start with a letter
	name := fmt.Sprintf("p%s", acctest.RandString(7))
	resourceName := "hydra_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHydraProjectDestroy,
		Steps: []resource.TestStep{
			// Test creation of project
			{
				Config: testAccHydraProjectConfigHiddenDisabled(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName),
				),
			},
		},
	})
}

func TestAccHydraProject_declarative(t *testing.T) {
	// identifier must start with a letter
	name := fmt.Sprintf("p%s", acctest.RandString(7))
	resourceName := "hydra_project.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHydraProjectDestroy,
		Steps: []resource.TestStep{
			// Test creation of project
			{
				Config: testAccHydraProjectConfigDeclarative(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName),
				),
			},
		},
	})
}

// testAccCheckExampleResourceDestroy verifies the Project has been destroyed
func testAccCheckHydraProjectDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*api.ClientWithResponses)
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hydra_project" {
			continue
		}

		projectID := rs.Primary.ID

		get, err := client.GetProjectIdWithResponse(ctx, projectID)
		if err != nil {
			return err
		}
		defer get.HTTPResponse.Body.Close()

		// Check to make sure the project doesn't exist
		if get.HTTPResponse.StatusCode == http.StatusOK {
			return fmt.Errorf("Expected project %s to be destroyed", projectID)
		}
	}

	return nil
}

func testAccHydraProjectConfigHiddenDisabled(name string) string {
	return fmt.Sprintf(`
resource "hydra_project" "test" {
  name         = "%s"
  display_name = "Nixpkgs"
  description  = "Nix Packages collection"
  homepage     = "http://nixos.org/nixpkgs"
  owner        = "%s"
  enabled = false
  visible = false
}
`, name, os.Getenv("HYDRA_USERNAME"))
}

func testAccHydraProjectConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "hydra_project" "test" {
  name         = "%s"
  display_name = "Nixpkgs"
  description  = "Nix Packages collection"
  homepage     = "http://nixos.org/nixpkgs"
  owner        = "%s"
  enabled = true
  visible = true
}
`, name, os.Getenv("HYDRA_USERNAME"))
}

func testAccHydraProjectConfigDeclarative(name string) string {
	return fmt.Sprintf(`
resource "hydra_project" "test" {
  name         = "%s"
  display_name = "Nixpkgs"
  description  = "Nix Packages collection"
  homepage     = "http://nixos.org/nixpkgs"
  owner        = "%s"
  enabled      = false
  visible      = true

  declarative {
    file  = "static-declarative-project/declarative.json"
    type  = "git"
    value = "https://github.com/DeterminateSystems/hydra-examples.git main"
  }
}
`, name, os.Getenv("HYDRA_USERNAME"))
}

// testAccCheckProjectExists verifies the project was successfully created
func testAccCheckProjectExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Resource not found for %s", name)
		}

		projectID := rs.Primary.ID
		if projectID == "" {
			return fmt.Errorf("No ID is set for %s", name)
		}

		client := testAccProvider.Meta().(*api.ClientWithResponses)
		ctx := context.Background()

		get, err := client.GetProjectIdWithResponse(ctx, projectID)
		if err != nil {
			return err
		}
		defer get.HTTPResponse.Body.Close()

		// Check to make sure the project was created
		if get.HTTPResponse.StatusCode != http.StatusOK {
			return fmt.Errorf("Expected project %s to be created", projectID)
		}

		return nil
	}
}
