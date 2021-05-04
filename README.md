# terraform-provider-hydra

The Terraform Hydra provider is a plugin for Terraform that allows for
declarative management of a [Hydra] instance.

## Requirements

To use this provider, you will need the following:

* [Terraform] 0.13+
* A Hydra instance running [commit
`6e53767`](https://github.com/NixOS/hydra/commit/6e537671dfa21f89041cbe16f0b461fe44327038)
or later

## Getting started

To get started with this provider, you'll need to create a configuration file
that will tell Terraform to use this provider. This will look something like the
following snippet:

```terraform
terraform {
  required_providers {
    hydra = {
      version = "~> 0.1"
      source  = "DeterminateSystems/hydra"
    }
  }
}
```

After that's done, you'll need to specify where your Hydra instance can be
reached and provide credentials for this provider to be able to work its magic:

> **NOTE:** Hard-coded credentials are not recommended, so while it is possible
to use them (just uncomment the `username` and `password` items and fill them in
with valid values), you are urged to use the `HYDRA_USERNAME` and
`HYDRA_PASSWORD` environment variables.

```terraform
provider "hydra" {
  host = "https://hydra.example.com"
  # username = "alice"
  # password = "foobar"
}
```

Now that you can connect to Hydra, it's time to create a project with the
[`hydra_project` resource](./docs/resources/project.md):

```terraform
resource "hydra_project" "nixpkgs" {
  name         = "nixpkgs"
  display_name = "Nixpkgs"
  description  = "Nix Packages collection"
  homepage     = "https://nixos.org/nixpkgs"
  owner        = "alice"
  enabled      = true
  visible      = true
}
```

You can attach a jobset to this project with the [`hydra_jobset`
resource](./docs/resources/jobset.md):

> **NOTE:** The `check_interval` is 0 for this example to prevent Hydra from
starting an evaluation on the entirety of Nixpkgs. Change this to a non-zero
value if you would like to tell Hydra it can start evaluating this jobset.

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
  keep_evaluations  = 3

  email_notifications = true
  email_override      = "example@example.com"
}
```

That's it for the basic usage of this provider!

You may also want to check out the example configurations inside [the
`examples/` directory](./examples/).

### Importing from an existing Hydra instance

You can migrate from a hand-configured Hydra to Terraform-managed configuration
files using our included generator,
[`./tools/generator.sh`](./tools/generator.sh).

The generator enumerates the server's projects and jobsets, generating a `.tf`
file for each project. The generator also produces a script of `terraform
import` commands.

The workflow is:

1. Execute `generator.sh`
2. Commit the generated `.tf` files to your repository
3. Execute the generated `terraform import` script
4. Discard the `terraform import` script, as it should not be necessary anymore

Your Terraform network and state file will now have up-to-date data for all of
your existing project and jobset resources, and a `terraform plan` should report
no differences were detected.

```shell
$ cd tools
$ nix-shell
# Usage: generator.sh <server-root> <out-dir> <import-file>
#
#     Arguments:
#         <server-root>    The root of the Hydra server to import projects and jobsets from.
#         <out-dir>        The directory to output generated Terraform configuration files to.
#         <import-file>    Where to write the generated list of 'terraform import' statements.
nix-shell$ ./generator.sh hydra.example.com outdir generated-tf-import.sh
```

## Development

In addition to the dependencies for using this provider, hacking on this
provider also requires the following:

* [Go] 1.15+
* [`oapi-codegen`]

### Running locally

This assumes a running instance of Hydra is available.

```shell
$ nix-shell
nix-shell$ make install
nix-shell$ cd examples/default
nix-shell$ terraform init && terraform plan
```

### Regenerating API bindings

This will fetch the latest `hydra-api.yaml` from Hydra and generate API bindings
against that specification.

```shell
$ nix-shell
nix-shell$ make api
```

### Running acceptance tests locally

**NOTE:** You should use a throwaway Hydra instance to prevent anything
unexpected happening.

```shell
$ nix-shell
nix-shell$ HYDRA_HOST=http://0:63333 HYDRA_USERNAME=alice HYDRA_PASSWORD=foobar make testacc
```

## Contributing

Pull requests are welcome. When submitting one, please follow the checklist in
the template to ensure everything works properly.

The typical contribution workflow is as follows:

1. Make your change
1. Format it with `make fmt` (requires [`goimports`])
1. Verify it builds with `make build`
1. Install it with `make install`
1. Spin up a local Hydra server to test with (see the Hydra documentation on
[Executing Hydra During Development](https://github.com/NixOS/hydra/blob/6e537671dfa21f89041cbe16f0b461fe44327038/README.md#executing-hydra-during-development))
1. Extend one of the [examples](./examples/) so that it will exercise your
change (or write your own example!)
1. Remove the `.terraform.lock.hcl` file (if it exists) and run `terraform init
&& terraform apply`
1. Once everything looks good, write a test for your change
1. Commit and open a pull request (be sure to follow the checklist in the
template)

## License

[MPL-2.0](LICENSE)

[Hydra]: https://github.com/NixOS/hydra
[Terraform]: https://www.terraform.io/downloads.html
[Go]: https://golang.org/doc/install
[`oapi-codegen`]: https://github.com/deepmap/oapi-codegen
[`goimports`]: https://pkg.go.dev/golang.org/x/tools/cmd/goimports
