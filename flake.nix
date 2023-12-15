{
  inputs = {
    nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1.514192.tar.gz";

    flake-compat.url = "https://flakehub.com/f/edolstra/flake-compat/1.0.1.tar.gz";
  };

  outputs = { self, ... }@inputs:
    let
      inherit (inputs.nixpkgs) lib;

      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";

      version = "${builtins.substring 0 8 lastModifiedDate}-${self.shortRev or "dirty"}";

      forSystems = s: f: lib.genAttrs s (system: f rec {
        inherit system;
        pkgs = import inputs.nixpkgs { inherit system; config.allowUnfree = true; };
      });

      forAllSystems = forSystems [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
    in
    {
      devShells = forAllSystems ({ system, pkgs, ... }:
        {
          default = pkgs.mkShell {
            name = "dev";
            buildInputs = with pkgs; [
              go
              terraform_1
              curl
              oapi-codegen
              # goimports # better than gofmt because it adds missing imports
              golint
              git-absorb
              findutils
            ];
          };
        });
    };
}
