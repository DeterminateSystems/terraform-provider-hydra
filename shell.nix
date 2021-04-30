let
  sources = import ./nix/sources.nix;
  pkgs = import sources.nixpkgs { };
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    terraform_0_14
    curl
    oapi-codegen
    goimports # better than gofmt because it adds missing imports
    git-absorb
    findutils
  ];
}
