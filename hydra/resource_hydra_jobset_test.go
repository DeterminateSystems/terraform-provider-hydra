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

func TestAccHydraJobset_basic(t *testing.T) {
	// identifier must start with a letter
	name := fmt.Sprintf("j%s", acctest.RandString(7))
	rename := fmt.Sprintf("%s-2", name)
	badname := "123"
	resourceName := "hydra_jobset.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHydraJobsetDestroy,
		Steps: []resource.TestStep{
			// Test creation of jobset
			{
				Config: testAccHydraJobsetConfigBasic(name, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
			// Test rename of jobset
			{
				Config: testAccHydraJobsetConfigBasic(name, rename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
			// Test invalid jobset identifier
			{
				Config:      testAccHydraJobsetConfigBasic(name, badname),
				ExpectError: regexp.MustCompile("Invalid jobset identifier"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
			// Test if jobset has all required fields set
			{
				Config:      testAccHydraJobsetConfigEmptyNixExpr(name, name),
				ExpectError: regexp.MustCompile(`Jobset type "legacy" requires a non-empty nix_expression.`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
			// Test if jobset has all required fields set
			{
				Config:      testAccHydraJobsetConfigEmptyInputs(name, name),
				ExpectError: regexp.MustCompile(`Jobset type "legacy" requires non-empty input`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
		},
	})
}

func TestAccHydraJobset_flake(t *testing.T) {
	// identifier must start with a letter
	name := fmt.Sprintf("j%s", acctest.RandString(7))
	rename := fmt.Sprintf("%s-2", name)
	badname := "123"
	resourceName := "hydra_jobset.test-flake"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHydraJobsetDestroy,
		Steps: []resource.TestStep{
			// Test creation of flake jobset
			{
				Config: testAccHydraJobsetConfigFlake(name, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
			// Test rename of flake jobset
			{
				Config: testAccHydraJobsetConfigFlake(name, rename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
			// Test invalid jobset identifier
			{
				Config:      testAccHydraJobsetConfigFlake(name, badname),
				ExpectError: regexp.MustCompile("Invalid jobset identifier"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
			// Test if jobset has all required fields set
			{
				Config:      testAccHydraJobsetConfigEmptyFlake(name, name),
				ExpectError: regexp.MustCompile(`Jobset type "flake" requires a non-empty flake_uri.`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
		},
	})
}

func TestAccHydraJobset_hiddenDisabled(t *testing.T) {
	// identifier must start with a letter
	name := fmt.Sprintf("j%s", acctest.RandString(7))
	flakeName := fmt.Sprintf("j%s", acctest.RandString(7))
	resourceName := "hydra_jobset.test"
	resourceNameFlake := "hydra_jobset.test-flake"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHydraJobsetDestroy,
		Steps: []resource.TestStep{
			// Test creation of hidden / disabled jobset
			{
				Config: testAccHydraJobsetConfigHiddenDisabled(name, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
			// Test creation of hidden / disabled flake jobset
			{
				Config: testAccHydraJobsetConfigHiddenDisabledFlake(flakeName, flakeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceNameFlake),
				),
			},
		},
	})
}

func TestAccHydraJobset_inputs(t *testing.T) {
	// identifier must start with a letter
	name := fmt.Sprintf("j%s", acctest.RandString(7))
	inputName1 := fmt.Sprintf("i%s", acctest.RandString(7))
	inputName2 := fmt.Sprintf("i%s", acctest.RandString(7))
	resourceName := "hydra_jobset.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHydraJobsetDestroy,
		Steps: []resource.TestStep{
			// Test creation of jobset with inputs
			{
				Config: testAccHydraJobsetConfigBasic(name, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
				),
			},
			// Test changing jobset inputs
			{
				Config: testAccHydraJobsetConfigChangedInput(name, name, inputName1, inputName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
					testAccCheckJobsetInputsChanged(resourceName, inputName1, inputName2),
				),
			},
		},
	})
}

func TestAccHydraJobset_legacyToFlakeAndBack(t *testing.T) {
	// identifier must start with a letter
	name := fmt.Sprintf("j%s", acctest.RandString(7))
	resourceName := "hydra_jobset.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHydraJobsetDestroy,
		Steps: []resource.TestStep{
			// Test creation of basic jobset
			{
				Config: testAccHydraJobsetConfigBasic(name, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
					testAccCheckJobsetType(resourceName, 0),
				),
			},
			// Test jobset changing from legacy to flake
			{
				Config: testAccHydraJobsetConfigLegacyToFlake(name, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
					testAccCheckJobsetType(resourceName, 1),
				),
			},
			// Test jobset changing from flake back to legacy
			{
				Config: testAccHydraJobsetConfigBasic(name, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobsetExists(resourceName),
					testAccCheckJobsetType(resourceName, 0),
				),
			},
		},
	})
}

// testAccCheckExampleResourceDestroy verifies the Jobset has been destroyed
func testAccCheckHydraJobsetDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*api.ClientWithResponses)
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hydra_jobset" {
			continue
		}

		jobsetID := rs.Primary.Attributes["name"]
		projectID := rs.Primary.Attributes["project"]

		get, err := client.GetJobsetProjectIdJobsetIdWithResponse(ctx, projectID, jobsetID)
		if err != nil {
			return err
		}
		defer get.HTTPResponse.Body.Close()

		// Check to make sure the jobset doesn't exist
		if get.HTTPResponse.StatusCode == http.StatusOK {
			return fmt.Errorf("Expected jobset %s in project %s to be destroyed", jobsetID, projectID)
		}
	}

	return nil
}

// testAccCheckJobsetExists verifies the jobset was successfully created
func testAccCheckJobsetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Resource not found for %s", name)
		}

		jobsetID := rs.Primary.Attributes["name"]
		if jobsetID == "" {
			return fmt.Errorf("No jobset is set for %s", name)
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

func testAccCheckJobsetInputsChanged(name string, inputName1 string, inputName2 string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Resource not found for %s", name)
		}

		jobsetID := rs.Primary.Attributes["name"]
		if jobsetID == "" {
			return fmt.Errorf("No jobset is set for %s", name)
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

		jobset := get.JSON200
		if jobset.Inputs != nil && len(jobset.Inputs.AdditionalProperties) == 2 &&
			(jobset.Inputs.AdditionalProperties[inputName1].Name == nil || *jobset.Inputs.AdditionalProperties[inputName1].Name != inputName1) &&
			(jobset.Inputs.AdditionalProperties[inputName2].Name == nil || *jobset.Inputs.AdditionalProperties[inputName2].Name != inputName2) {
			return fmt.Errorf("Expected inputs to have changed")
		}

		return nil
	}
}

func testAccCheckJobsetType(name string, jobsetType int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Resource not found for %s", name)
		}

		jobsetID := rs.Primary.Attributes["name"]
		if jobsetID == "" {
			return fmt.Errorf("No jobset is set for %s", name)
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

		jobsetResponse := get.JSON200
		if jobsetType == 0 {
			if (jobsetResponse.Flake != nil && *jobsetResponse.Flake != "") ||
				(jobsetResponse.Type != nil && *jobsetResponse.Type == 1) {
				return fmt.Errorf("Expected jobset %s in project %s to be type legacy", jobsetID, projectID)
			}
		} else if jobsetType == 1 {
			if (jobsetResponse.Flake == nil || *jobsetResponse.Flake == "") ||
				(jobsetResponse.Type != nil && *jobsetResponse.Type == 0) {
				return fmt.Errorf("Expected jobset %s in project %s to be type flake", jobsetID, projectID)
			}
		}

		return nil
	}
}

func testAccHydraJobsetConfigBasic(project string, jobset string) string {
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
  name        = "%s"
  type        = "legacy"
  description = ""

  nix_expression {
    file  = "release.nix"
    input = "ofborg"
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
}`, project, os.Getenv("HYDRA_USERNAME"), jobset)
}

func testAccHydraJobsetConfigFlake(project string, jobset string) string {
	return fmt.Sprintf(`
resource "hydra_project" "test-flake" {
  name         = "%s"
  display_name = "Nixpkgs"
  description  = "Nix Packages set"
  homepage     = "https://github.com/nixos/nixpkgs"
  owner        = "%s"
  enabled = true
  visible = true
}

resource "hydra_jobset" "test-flake" {
  project     = hydra_project.test-flake.name
  state       = "enabled"
  visible     = true
  name        = "%s"
  type        = "flake"
  description = "master branch"

  flake_uri = "github:NixOS/nixpkgs/master"

  check_interval    = 0
  scheduling_shares = 3000

  email_notifications = true
  email_override      = "example@example.com"
  keep_evaluations    = 3
}`, project, os.Getenv("HYDRA_USERNAME"), jobset)
}

func testAccHydraJobsetConfigHiddenDisabled(project string, jobset string) string {
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

resource "hydra_jobset" "test" {
  project     = hydra_project.test.name
  state       = "disabled"
  visible     = false
  name        = "%s"
  type        = "legacy"
  description = ""

  nix_expression {
    file  = "release.nix"
    input = "ofborg"
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
}`, project, os.Getenv("HYDRA_USERNAME"), jobset)
}

func testAccHydraJobsetConfigHiddenDisabledFlake(project string, jobset string) string {
	return fmt.Sprintf(`
resource "hydra_project" "test-flake" {
  name         = "%s"
  display_name = "Nixpkgs"
  description  = "Nix Packages set"
  homepage     = "https://github.com/nixos/nixpkgs"
  owner        = "%s"
  enabled = false
  visible = false
}

resource "hydra_jobset" "test-flake" {
  project     = hydra_project.test-flake.name
  state       = "disabled"
  visible     = false
  name        = "%s"
  type        = "flake"
  description = "master branch"

  flake_uri = "github:NixOS/nixpkgs/master"

  check_interval    = 0
  scheduling_shares = 3000

  email_notifications = true
  email_override      = "example@example.com"
  keep_evaluations    = 3
}`, project, os.Getenv("HYDRA_USERNAME"), jobset)
}

func testAccHydraJobsetConfigEmptyFlake(project string, jobset string) string {
	return fmt.Sprintf(`
resource "hydra_project" "test-flake" {
  name         = "%s"
  display_name = "Nixpkgs"
  description  = "Nix Packages set"
  homepage     = "https://github.com/nixos/nixpkgs"
  owner        = "%s"
  enabled = true
  visible = true
}

resource "hydra_jobset" "test-flake" {
  project     = hydra_project.test-flake.name
  state       = "enabled"
  visible     = true
  name        = "%s"
  type        = "flake"
  description = "master branch"

  check_interval    = 0
  scheduling_shares = 3000

  email_notifications = true
  email_override      = "example@example.com"
  keep_evaluations    = 3
}`, project, os.Getenv("HYDRA_USERNAME"), jobset)
}

func testAccHydraJobsetConfigEmptyNixExpr(project string, jobset string) string {
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

resource "hydra_jobset" "test" {
  project     = hydra_project.test.name
  state       = "disabled"
  visible     = false
  name        = "%s"
  type        = "legacy"
  description = ""

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
}`, project, os.Getenv("HYDRA_USERNAME"), jobset)
}

func testAccHydraJobsetConfigEmptyInputs(project string, jobset string) string {
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

resource "hydra_jobset" "test" {
  project     = hydra_project.test.name
  state       = "disabled"
  visible     = false
  name        = "%s"
  type        = "legacy"
  description = ""

  nix_expression {
    file  = "release.nix"
    input = "ofborg"
  }

  check_interval    = 0
  scheduling_shares = 3000

  email_notifications = false
  keep_evaluations    = 3
}`, project, os.Getenv("HYDRA_USERNAME"), jobset)
}

func testAccHydraJobsetConfigChangedInput(project string, jobset string, inputName1 string, inputName2 string) string {
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
  name        = "%s"
  type        = "legacy"
  description = ""

  nix_expression {
    file  = "release.nix"
    input = "%s"
  }

  check_interval    = 0
  scheduling_shares = 3000

  email_notifications = false
  keep_evaluations    = 3

  input {
    name              = "%s"
    type              = "git"
    value             = "https://github.com/NixOS/nixpkgs.git nixpkgs-unstable"
    notify_committers = false
  }

  input {
    name              = "%s"
    type              = "git"
    value             = "https://github.com/nixos/ofborg.git released"
    notify_committers = false
  }
}`, project, os.Getenv("HYDRA_USERNAME"), jobset, inputName2, inputName1, inputName2)
}

func testAccHydraJobsetConfigLegacyToFlake(project string, jobset string) string {
	return fmt.Sprintf(`
resource "hydra_project" "test" {
  name         = "%s"
  display_name = "Nixpkgs"
  description  = "Nix Packages set"
  homepage     = "https://github.com/nixos/nixpkgs"
  owner        = "%s"
  enabled = true
  visible = true
}

resource "hydra_jobset" "test" {
  project     = hydra_project.test.name
  state       = "enabled"
  visible     = true
  name        = "%s"
  type        = "flake"
  description = "master branch"

  flake_uri = "github:NixOS/nixpkgs/master"

  check_interval    = 0
  scheduling_shares = 3000

  email_notifications = true
  email_override      = "example@example.com"
  keep_evaluations    = 3
}`, project, os.Getenv("HYDRA_USERNAME"), jobset)
}
