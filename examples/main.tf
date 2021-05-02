terraform {
  required_providers {
    hydra = {
      version = "~> 0.1"
      source  = "DeterminateSystems/hydra"
    }
  }
}

provider "hydra" {
  host     = "http://0:63333"
  username = "alice"
  password = "foobar"
}

resource "hydra_project" "nixpkgs" {
  name         = "nixpkgs"
  display_name = "Nixpkgs"
  description  = "Nix Packages collection"
  homepage     = "http://nixos.org/nixpkgs"
  owner        = "alice"
  # TODO: declarative configuration
  enabled = true
  visible = true
}

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
