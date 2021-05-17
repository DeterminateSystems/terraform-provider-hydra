terraform {
  required_providers {
    hydra = {
      version = "~> 0.1"
      source  = "DeterminateSystems/hydra"
    }
  }
}

provider "hydra" {
  host     = "http://0.0.0.0:63333"
  username = "alice"
  password = "foobar"
}

resource "hydra_project" "nixpkgs-declarative" {
  name         = "nixpkgs-declarative"
  display_name = "Nixpkgs"
  description  = "Nix Packages collection"
  homepage     = "https://nixos.org/nixpkgs"
  owner        = "alice"
  enabled      = false
  visible      = true

  declarative {
    file  = "static-declarative-project/declarative.json"
    type  = "git"
    value = "https://github.com/DeterminateSystems/hydra-examples.git main"
  }
}
