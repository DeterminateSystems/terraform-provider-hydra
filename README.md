# terraform-provider-hydra

## Running locally

```shell
$ nix-shell
nix-shell$ make install
nix-shell$ cd examples
nix-shell$ terraform init && terraform plan
```

## Regenerate API Bindings

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
