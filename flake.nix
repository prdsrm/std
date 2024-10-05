{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
  };

  outputs = {
    self,
    flake-utils,
    nixpkgs,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
         config.allowUnfree = true;
      };
    in {
      devShell = pkgs.mkShell {
        nativeBuildInputs = with pkgs; [
          # Golang & build tools
          gnumake
          go
          golangci-lint
          gopls
          delve
          # Postgres
          go-migrate
          postgresql_16
        ];
      };
    });
}
