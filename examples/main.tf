terraform {
  required_providers {
    hydra = {
      version = "0.1"
      source  = "DeterminateSystems/hydra"
    }
  }
}

provider "hydra" {

}

resource "hydra_project" "nixpkgs" {
  name         = "nixpkgs"
  display_name = "Nixpkgs"
  description  = "Nix Packages collection"
  homepage     = "http://nixos.org/nixpkgs"
  owner        = "eelco"
  # TODO: declarative configuration
  enabled = true
  visible = false
}
