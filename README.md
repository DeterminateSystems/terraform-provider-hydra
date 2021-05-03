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

[Hydra]: https://github.com/NixOS/hydra/
