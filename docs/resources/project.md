# Project Resource

The Project resource defines a [Hydra project] to be managed by Terraform.

## Example Usage

```terraform
resource "hydra_project" "nixpkgs" {
  name         = "nixpkgs"
  display_name = "Nixpkgs"
  description  = "Nix Packages collection"
  homepage     = "http://nixos.org/nixpkgs"
  owner        = "alice"
  enabled      = true
  visible      = true

  declarative {
    file  = "static-declarative-project/declarative.json"
    type  = "git"
    value = "https://github.com/DeterminateSystems/hydra-examples.git main"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the project.

* `display_name` - (Required) The display name of the project.

* `description` - (Optional) A description of the project.

* `homepage` - (Optional) The homepage of the project.

* `owner` - (Required) The owner of the project (a Hydra user).

* `enabled` - (Optional) Whether or not the project is enabled.

* `visible` - (Optional) Whether or not the project is visible.

* `declarative` - (Optional) Configuration of the declarative project.

  * `file` - (Required) The file in `value` which contains the declarative spec file. Relative to the root of `input`.

  * `type` - (Required) The type of the declarative input.

  * `value` - (Required) The value of the declarative input.

[Hydra project]: https://github.com/NixOS/hydra/blob/e9a06113c955e457fa59717c4964c302e852ee9b/doc/manual/src/projects.md#creating-and-managing-projects
