let
  sources = import ../nix/sources.nix;
  pkgs = import sources.nixpkgs { };
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    curl
    jq
    shellcheck
    coreutils # basename, mktemp
  ];
}
