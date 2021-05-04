# Hydra Provider

This provider allows one to manage the projects and jobsets of a particular
[Hydra] instance.

## Example Usage

```terraform
terraform {
  required_providers {
    hydra = {
      version = "~> 0.1"
      source  = "DeterminateSystems/hydra"
    }
  }
}

provider "hydra" {
  host = "https://hydra.example.com"
}
```

## Authentication

The Hydra provider supports two means of providing authentication credentials:

- Environment variables
- Static credentials

-> The user specified by either a static `username` or `HYDRA_USERNAME`
environment variable should have the permissions necessary for project and
jobset creation, modification, and deletion.

### Environment Variables

You can provide your credentials via the `HYDRA_USERNAME` and `HYDRA_PASSWORD`
environment variables.

```terraform
provider "hydra" {
  host = "https://hydra.example.com"
}
```

### Static Credentials

!> **Warning:** Hard-coded credentials are not recommended in any Terraform
configuration and risks secret leakage should this file ever be committed to a
public version control system.

Static credentials can be provided by adding a `username` and `password` to the
Hydra provider block:

```terraform
provider "hydra" {
  host     = "https://hydra.example.com"
  username = "alice"
  password = "foobar"
}
```

## Argument Reference

* `host` - (Optional) This is the address of the Hydra instance. It must be
provided, but it can also be sourced from the `HYDRA_HOST` environment variable.

* `username` - (Optional) This is the user that Terraform will be logging in as.
It must be provided, but it can also be sourced from the `HYDRA_USERNAME`
environment variable.

* `password` - (Optional) This is the password for the Hydra user specified in
`username`. It must be provided, but it can also be sourced from the
`HYDRA_PASSWORD` environment variable.

[Hydra]: https://github.com/NixOS/hydra/
