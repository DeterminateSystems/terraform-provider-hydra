# terraform-provider-hydra

## Testing locally

```shell
$ nix-shell
nix-shell$ make install
nix-shell$ cd examples
nix-shell$ terraform init && terraform plan
```

## Regenerate API Bindings

```
$ nix-shell
nix-shell$ make api
```
