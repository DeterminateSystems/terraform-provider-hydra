# Jobset Resource

The Jobset resource defines a [Hydra jobset] to be managed by Terraform.

## Example Usage

### Type `legacy`

```terraform
resource "hydra_jobset" "trunk" {
  project     = hydra_project.nixpkgs.name
  state       = "enabled"
  visible     = true
  name        = "trunk"
  type        = "legacy"
  description = "master branch"

  nix_expression {
    file  = "pkgs/top-level/release.nix"
    input = "nixpkgs"
  }

  check_interval    = 0
  scheduling_shares = 3000

  email_notifications = true
  email_override      = "example@example.com"
  keep_evaluations    = 3

  input {
    name              = "nixpkgs"
    type              = "git"
    value             = "https://github.com/NixOS/nixpkgs.git"
    notify_committers = false
  }

  input {
    name              = "officialRelease"
    type              = "boolean"
    value             = "false"
    notify_committers = false
  }
}
```

### Type `flake`

```terraform
resource "hydra_jobset" "trunk-flake" {
  project     = hydra_project.nixpkgs.name
  state       = "enabled"
  visible     = true
  name        = "trunk-flake"
  type        = "flake"
  description = "master branch"

  flake_uri = "github:NixOS/nixpkgs/master"

  check_interval    = 0
  scheduling_shares = 3000

  email_notifications = true
  email_override      = "example@example.com"
  keep_evaluations    = 3
}
```

## Argument Reference

* `project` - (Required) The name of the parent project.

* `state` - (Required) The state of the jobset. One of `disabled`, `enabled`,
`one-shot`, or `one-at-a-time`.

* `visible` - (Optional) Whether or not the jobset is visible.

* `name` - (Required) The name of the jobset.

* `type` - (Required) The type of the jobset. Either `legacy` or `flake`.

* `description` - (Optional) The description of the jobset.

* `flake_uri` - (Required when the `type` is `flake`, otherwise prohibited.) The
jobset's flake URI.

* `input` - (Required when the `type` is `legacy`, otherwise prohibited.)
Input(s) to be provided to the jobset.

  * `name` - (Required) The name of the input.

  * `type` - (Required) The type of the input.

  * `value` - (Required) The value of the jobset.

  * `notify_committers` - (Optional) Whether or not to notify committers.

* `nix_expression` - (Required when the `type` is `legacy`, otherwise
prohibited.) The jobset's entrypoint Nix expression.

  * `file` - (Required) The file containing the Nix expression.

  * `input` - (Required) The input where the `file` is located.

* `check_interval` - (Required) How frequently to check the jobset in seconds (0
disables polling).

* `scheduling_shares` - (Required) How many shares allocated to the jobset.

* `email_notifications` - (Optional) Whether or not to send email notifications.

* `email_override` - (Optional) An email, or a comma-separated list of emails,
to send email notifications to.

* `keep_evaluations` - (Required) How many of the jobset's evaluations to keep.

[Hydra jobset]: https://github.com/NixOS/hydra/blob/e9a06113c955e457fa59717c4964c302e852ee9b/doc/manual/src/projects.md#job-sets
