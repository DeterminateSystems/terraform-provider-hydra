package hydra

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"terraform-provider-hydra/hydra/api"
)

func TestAccHydraJobset_basic(t *testing.T) {
	// identifier must start with a letter
	name := fmt.Sprintf("j%s", acctest.RandString(7))
	// rename := fmt.Sprintf("%s-2", name)
	// badname := "123"
	resourceName := "hydra_jobset.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHydraJobsetDestroy,
		Steps: []resource.TestStep{
			// Test creation of project
			{
				Config: testAccHydraJobsetConfigBasic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
			// Test modification of project name
			// {
			// 	Config: testAccHydraProjectConfigBasic(rename),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckProjectExists(resourceName),
			// 	),
			// },
			// // Test invalid project identifier
			// {
			// 	Config:      testAccHydraProjectConfigBasic(badname),
			// 	ExpectError: regexp.MustCompile("Invalid project identifier"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckProjectExists(resourceName),
			// 	),
			// },
		},
	})
}

// testAccCheckExampleResourceDestroy verifies the Project has been destroyed
func testAccCheckHydraJobsetDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*api.ClientWithResponses)
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hydra_project" {
			continue
		}

		jobsetID := rs.Primary.ID
		projectID := rs.Primary.Attributes["project"]

		get, err := client.GetJobsetProjectIdJobsetIdWithResponse(ctx, projectID, jobsetID)
		if err != nil {
			return err
		}
		defer get.HTTPResponse.Body.Close()

		// Check to make sure the project doesn't exist
		if get.HTTPResponse.StatusCode == http.StatusOK {
			return fmt.Errorf("Expected jobset %s in project %s to be destroyed", jobsetID, projectID)
		}
	}

	return nil
}

// func testAccHydraJobsetConfigHiddenDisabled(name string) string {
// 	return fmt.Sprintf(`
// resource "hydra_project" "test" {
//   name         = "%s"
//   display_name = "Nixpkgs"
//   description  = "Nix Packages collection"
//   homepage     = "http://nixos.org/nixpkgs"
//   owner        = "%s"
//   enabled = false
//   visible = false
// }

// resource "hydra_jobset" "test" {
//   project     = hydra_project.test.name
//   state       = "enabled"
//   visible     = true
//   name        = "release"
//   type        = "legacy"
//   description = ""

//   nix_expression {
//     file = "release.nix"
//     in   = "ofborg"
//   }

//   check_interval    = 0
//   scheduling_shares = 3000

//   email_notifications = false
//   keep_evaluations    = 3

//   input {
//     name              = "nixpkgs"
//     type              = "git"
//     value             = "https://github.com/NixOS/nixpkgs.git nixpkgs-unstable"
//     notify_committers = false
//   }

//   input {
//     name              = "ofborg"
//     type              = "git"
//     value             = "https://github.com/nixos/ofborg.git released"
//     notify_committers = false
//   }
// }`, name, os.Getenv("HYDRA_USERNAME"))
// }

func testAccHydraJobsetConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "hydra_project" "test" {
  name         = "%s"
  display_name = "Ofborg"
  description  = "ofborg automation"
  homepage     = "https://github.com/nixos/ofborg"
  owner        = "%s"
  enabled = true
  visible = true
}

resource "hydra_jobset" "test" {
  project     = hydra_project.test.name
  state       = "enabled"
  visible     = true
  name        = "release"
  type        = "legacy"
  description = ""

  nix_expression {
    file = "release.nix"
    in   = "ofborg"
  }

  check_interval    = 0
  scheduling_shares = 3000

  email_notifications = false
  keep_evaluations    = 3

  input {
    name              = "nixpkgs"
    type              = "git"
    value             = "https://github.com/NixOS/nixpkgs.git nixpkgs-unstable"
    notify_committers = false
  }

  input {
    name              = "ofborg"
    type              = "git"
    value             = "https://github.com/nixos/ofborg.git released"
    notify_committers = false
  }
}`, name, os.Getenv("HYDRA_USERNAME"))
}

// testAccCheckProjectExists verifies the project was successfully created
func testAccCheckJobsetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Resource not found for %s", name)
		}

		jobsetID := rs.Primary.ID
		if jobsetID == "" {
			return fmt.Errorf("No ID is set for %s", name)
		}
		projectID := rs.Primary.Attributes["project"]
		if projectID == "" {
			return fmt.Errorf("No project is set for %s", name)
		}

		client := testAccProvider.Meta().(*api.ClientWithResponses)
		ctx := context.Background()

		get, err := client.GetJobsetProjectIdJobsetIdWithResponse(ctx, projectID, jobsetID)
		if err != nil {
			return err
		}
		defer get.HTTPResponse.Body.Close()

		// Check to make sure the jobset was created
		if get.HTTPResponse.StatusCode != http.StatusOK {
			return fmt.Errorf("Expected jobset %s in project %s to be created", jobsetID, projectID)
		}

		return nil
	}
}
