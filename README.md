# terraform-provider-hydra

The Terraform Hydra provider is a plugin for Terraform that allows for
declarative management of a [Hydra] instance.

<sup>**NOTE**: Only Hydra instances running [commit
`5520f4b`](https://github.com/NixOS/hydra/commit/5520f4b7b626c1ff41fc5fb5b990ee279507b6b3)
or later are officially supported. That commit is the last in a series of
patches that flesh out Hydra's API and its responses to make it suitable for
automation with this plugin.</sup>

## Running locally

This assumes a running instance of Hydra is available.

```shell
$ nix-shell
nix-shell$ make install
nix-shell$ cd examples
nix-shell$ terraform init && terraform plan
```

## Regenerate API bindings

This will fetch the latest `hydra-api.yaml` from [Hydra] and generate API
bindings against that specification.

```shell
$ nix-shell
nix-shell$ make api
```

## Running acceptance tests locally

NOTE: You should use a throwaway Hydra instance to prevent anything unexpected
happening.

```shell
$ nix-shell
nix-shell$ HYDRA_HOST=http://0:63333 HYDRA_USERNAME=alice HYDRA_PASSWORD=foobar make testacc
```

## Importing from an existing Hydra instance

Migrate from a manually configured Hydra to Terraform-managed configuration
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

Your Terraform network and state file will now have up to date data for all of
your existing project and jobset resources, and a `terraform plan` should be
clean: no difference should be detected.

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

[Hydra]: https://github.com/NixOS/hydra/
